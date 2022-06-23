/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"

	//"strings"
	"time"

	//"github.com/go-logr/logr"
	"github.com/robfig/cron"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"sigs.k8s.io/controller-runtime/pkg/log"

	argoinappiov1 "monitoring/api/v1"
	analysisutil "monitoring/utils/analysis"
	timeutil "monitoring/utils/time"

	batchv1 "k8s.io/api/batch/v1"
	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//ref "k8s.io/client-go/tools/reference"
	"monitoring/metricproviders/prometheus"

	log "github.com/sirupsen/logrus"
)

// InAppMetricReconciler reconciles a InAppMetric object
type InAppMetricReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	//Log    logr.Logger
	Clock
}

/* Fake clock for testing */
type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

type Clock interface {
	Now() time.Time
}

var (
	scheduledTimeAnnotation             = "monitoring/scheduled-at"
	EnvVarArgoRolloutsPrometheusAddress = "ARGO_ROLLOUTS_PROMETHEUS_ADDRESS"
)

func newMetricRun() *argoinappiov1.MetricRun {
	return &argoinappiov1.MetricRun{}
}

//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics/finalizers,verbs=update
//+kubebuilder:rbac:groups=argo-in-app.io,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argo-in-app.io,resources=jobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InAppMetric object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *InAppMetricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var inAppMetric argoinappiov1.InAppMetric
	//var metricRun argoinappiov1.MetricRun

	err := r.Get(context.TODO(), req.NamespacedName, &inAppMetric)
	if err != nil {
		ctrl.Log.Error(err, "Error getting instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	getNextSchedule := func(metric *argoinappiov1.InAppMetric, now time.Time) (lastMissed time.Time, next time.Time, err error) {
		sched, err := cron.ParseStandard(metric.Spec.Schedule)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", metric.Spec.Schedule, err)
		}

		var earliestTime time.Time
		if metric.Status.LastScheduleTime != nil {
			earliestTime = metric.Status.LastScheduleTime.Time
		} else {
			earliestTime = metric.ObjectMeta.CreationTimestamp.Time
		}

		if earliestTime.After(now) {
			return time.Time{}, sched.Next(now), nil
		}

		starts := 0
		for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
			lastMissed = t
			starts++
			if starts > 100 {
				return time.Time{}, time.Time{}, fmt.Errorf("too many missed starts")
			}
		}
		return lastMissed, sched.Next(now), nil
	}

	missedRun, nextRun, err := getNextSchedule(&inAppMetric, r.Now())
	if err != nil {
		ctrl.Log.Error(err, "unable to figure out schedule")
		return ctrl.Result{}, nil
	}

	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())}
	ctrl.Log.WithValues("now", r.Now(), "next run", nextRun)

	/* Run a new job if it's on schedule, not past the deadline, and not blocked by our concurrency policy */
	if missedRun.IsZero() {
		ctrl.Log.Info("no upcoming scheduled times")
		return scheduledResult, nil
	}

	run := newMetricRun()
	err = runMeasurements(run, inAppMetric.Spec.Metrics)
	if err != nil {
		ctrl.Log.Error(err, "error")
	} else {
		ctrl.Log.Info(strconv.Itoa(len(run.Status.MetricResults)))
	}

	return scheduledResult, nil
}

func runMeasurements(run *argoinappiov1.MetricRun, tasks []argoinappiov1.Metric) error {
	for _, task := range tasks {
		e := log.Entry{}
		api, err := prometheus.NewPrometheusAPI(task)
		if err != nil {
			ctrl.Log.Error(err, "error creating api")
		}

		p := prometheus.NewPrometheusProvider(api, e)

		metricResult := analysisutil.GetResult(run, task.Name)
		if metricResult == nil {
			metricResult = &argoinappiov1.MetricResult{
				Name:     task.Name,
				Phase:    argoinappiov1.AnalysisPhaseRunning,
				Metadata: p.GetMetadata(task),
			}
		}

		var newMeasurement argoinappiov1.Measurement
		newMeasurement = p.Run(run, task)

		if newMeasurement.Phase.Completed() {
			ctrl.Log.Info("Measurement Completed.")
			if newMeasurement.FinishedAt == nil {
				finishedAt := timeutil.MetaNow()
				newMeasurement.FinishedAt = &finishedAt
			}
			switch newMeasurement.Phase {
			case argoinappiov1.AnalysisPhaseSuccessful:
				metricResult.Successful++
				metricResult.Count++
				metricResult.ConsecutiveError = 0
			case argoinappiov1.AnalysisPhaseFailed:
				metricResult.Failed++
				metricResult.Count++
				metricResult.ConsecutiveError = 0
			case argoinappiov1.AnalysisPhaseInconclusive:
				metricResult.Inconclusive++
				metricResult.Count++
				metricResult.ConsecutiveError = 0
			case argoinappiov1.AnalysisPhaseError:
				metricResult.Error++
				metricResult.ConsecutiveError++
				ctrl.Log.Info(newMeasurement.Message)
			}
		}

		metricResult.Measurements = append(metricResult.Measurements, newMeasurement)
		analysisutil.SetResult(run, *metricResult)
	}
	return nil
}

func Contains(arr []argoinappiov1.Metric, e argoinappiov1.Metric) bool {
	for _, T := range arr {
		if T == e {
			return true
		}
	}
	return false
}

func Index(arr []argoinappiov1.Metric, e argoinappiov1.Metric) bool {
	for _, T := range arr {
		if T == e {
			return true
		}
	}
	return false
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = argoinappiov1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *InAppMetricReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&argoinappiov1.InAppMetric{}).
		Owns(&batchv1.Job{}).
		Complete(r)

}
