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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/robfig/cron"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	argoinappiov1 "monitoring/api/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ref "k8s.io/client-go/tools/reference"
)

// InAppMetricReconciler reconciles a InAppMetric object
type InAppMetricReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
	Clock
}

/* Fake clock for testing */
type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

type Clock interface {
	Now() time.Time
}

var (
	scheduledTimeAnnotation = "monitoring/scheduled-at"
)

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
	_ = log.FromContext(ctx)

	var instance argoinappiov1.InAppMetric

	err := r.Get(context.TODO(), req.NamespacedName, &instance)
	if err != nil {
		ctrl.Log.Error(err, "Error getting instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	/* List all active jobs and update the status */
	var childJobs batchv1.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		ctrl.Log.Error(err, "Unable to list child Jobs")
		return ctrl.Result{}, err
	}
	// split jobs into active, successful and failed jobs
	var activeJobs []*batchv1.Job
	var successfulJobs []*batchv1.Job
	var failedJobs []*batchv1.Job
	var mostRecent *time.Time // find last run

	isJobFinished := func(job *batchv1.Job) (bool, batchv1.JobConditionType) {
		for _, c := range job.Status.Conditions {
			if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
				return true, c.Type
			}
		}
		return false, ""
	}

	getScheduledTimeForJob := func(job *batchv1.Job) (*time.Time, error) {
		timeRaw := job.Annotations[scheduledTimeAnnotation]
		if len(timeRaw) == 0 {
			return nil, nil
		}

		timeParsed, err := time.Parse(time.RFC3339, timeRaw)
		if err != nil {
			return nil, err
		}
		return &timeParsed, nil
	}

	for i, job := range childJobs.Items {
		_, finishedType := isJobFinished(&job)
		switch finishedType {
		case "": // ongoing
			activeJobs = append(activeJobs, &childJobs.Items[i])
		case batchv1.JobFailed:
			failedJobs = append(failedJobs, &childJobs.Items[i])
		case batchv1.JobComplete:
			successfulJobs = append(successfulJobs, &childJobs.Items[i])
		}

		scheduledTimeForJob, err := getScheduledTimeForJob(&job)
		if err != nil {
			ctrl.Log.Error(err, "unable to parse schedule time for child jobs", "job", &job)
			continue
		}
		if scheduledTimeForJob != nil {
			if mostRecent == nil {
				mostRecent = scheduledTimeForJob
			} else if mostRecent.Before(*scheduledTimeForJob) {
				mostRecent = scheduledTimeForJob
			}
		}
	}

	if mostRecent != nil {
		instance.Status.LastScheduleTime = &metav1.Time{Time: *mostRecent}
	} else {
		instance.Status.LastScheduleTime = nil
	}

	/* Append active jobs to current job */
	instance.Status.Active = nil
	for _, activeJob := range activeJobs {
		jobRef, err := ref.GetReference(r.Scheme, activeJob)
		if err != nil {
			ctrl.Log.Error(err, "unable to make reference to active job", "job", activeJob)
			continue
		}
		instance.Status.Active = append(instance.Status.Active, *jobRef)
	}

	ctrl.Log.Info("job count", "active jobs", len(activeJobs), "successful jobs", len(successfulJobs), "failed jobs", len(failedJobs))

	err = r.Status().Update(ctx, &instance)
	if err != nil {
		ctrl.Log.Error(err, "unable to update Job status")
		return ctrl.Result{}, err
	}

	/* Clean up old jobs */

	/* Check if suspended */

	/* Get next scheduled run */
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

	missedRun, nextRun, err := getNextSchedule(&instance, r.Now())
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

	ctrl.Log.WithValues("current run", missedRun)

	/* add in if we want to consider starting deadlines */

	/* for now just consider replacement concurrency */

	if instance.Spec.ConcurrencyPolicy == argoinappiov1.ReplaceConcurrent {
		for _, activeJob := range activeJobs {
			if err := r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				ctrl.Log.Error(err, "unable to delete active job", "job", activeJob)
				return ctrl.Result{}, err
			}
		}
	}

	/* Constructing a Job */
	constructJob := func(cr *argoinappiov1.InAppMetric, scheduledTime time.Time) (*batchv1.Job, error) {
		name := fmt.Sprintf("%s-%d", cr.Name, scheduledTime.Unix())

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name, // if the names are the same pass in a jobname
				Namespace: cr.Namespace,
				//Labels:    make(map[string]string),
				Annotations: make(map[string]string),
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "busybox",
								Image:   "busybox",
								Command: strings.Split(cr.Spec.Command, " "),
							},
						},
						RestartPolicy: corev1.RestartPolicyOnFailure,
					},
				},
			},
		}

		job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)

		if err := ctrl.SetControllerReference(cr, job, r.Scheme); err != nil {
			return nil, err
		}

		return job, nil
	}

	/* Make the job */
	job, err := constructJob(&instance, missedRun)
	if err != nil {
		ctrl.Log.Error(err, "unable to construct job")
		return scheduledResult, nil
	}

	if err := r.Create(ctx, job); err != nil {
		ctrl.Log.Error(err, "unable to create Job", "job", job)
		return ctrl.Result{}, err
	}

	ctrl.Log.Info("created job", "job", job)

	return scheduledResult, nil
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
