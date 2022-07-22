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
	"sort"

	"strconv"

	"strings"
	"time"

	"github.com/robfig/cron"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	argoinappiov1 "monitoring/api/v1"
	analysisutil "monitoring/utils/analysis"
	timeutil "monitoring/utils/time"

	"monitoring/utils/defaults"
	logutil "monitoring/utils/log"

	"monitoring/metricproviders"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

const (
	// SuccessfulAssessmentRunTerminatedResult is used for logging purposes when the metrics evaluation
	// is successful and the run is terminated.
	SuccessfulAssessmentRunTerminatedResult = "Metric Assessment Result - Successful: Run Terminated"
	scheduledTimeAnnotation                 = "argo-in-app.io/scheduled-at"
	// DefaultErrorRetryInterval is the default interval to retry a measurement upon error, in the
	// event an interval was not specified
	DefaultErrorRetryInterval = 10 * time.Second
)

// InAppMetricReconciler reconciles a InAppMetric object
type InAppMetricReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

/* Fake clock for testing */
type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

type Clock interface {
	Now() time.Time
}

/** Returns address of new MetricRun object **/
func newMetricRun() *argoinappiov1.MetricRun {
	return &argoinappiov1.MetricRun{}
}

//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=argo-in-app.io,resources=inappmetrics/finalizers,verbs=update
//+kubebuilder:rbac:groups=argo-in-app.io,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argo-in-app.io,resources=jobs/status,verbs=get
func (r *InAppMetricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var inAppMetric argoinappiov1.InAppMetric
	metricRunList := &argoinappiov1.MetricRunList{}
	opts := []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
	}

	// Get instance of inAppMetric object created by the yaml
	err := r.Get(context.TODO(), req.NamespacedName, &inAppMetric)
	if err != nil {
		ctrl.Log.Error(err, "Error getting instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all metric runs currently in the cluster
	err = r.List(context.TODO(), metricRunList, opts...)
	if err != nil {
		ctrl.Log.Error(err, "Could not list metric runs")
		return ctrl.Result{}, err
	}

	var mostRecentTime *time.Time

	// Order MetricRuns by their name -> which is the time they were made
	sort.Slice(metricRunList.Items, func(i, j int) bool {
		return metricRunList.Items[i].Name < metricRunList.Items[j].Name
	})

	var run *argoinappiov1.MetricRun

	// Iterate through all metric runs to update mostRecentTime
	for _, r := range metricRunList.Items {
		scheduledTimeForRun, err := getScheduledTimeForRun(&r)
		if err != nil {
			ctrl.Log.Error(err, "unable to parse schedule time for run", "run", r)
			continue
		}
		if scheduledTimeForRun != nil {
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForRun
			} else if mostRecentTime.Before(*scheduledTimeForRun) {
				mostRecentTime = scheduledTimeForRun
			}
		}
		for _, metric := range r.Spec.Metrics {
			if analysisutil.GetResult(&r, metric.Name) != nil {
				counter := analysisutil.GetResult(&r, metric.Name).Counter
				if metric.Count != nil && int(counter) != 0 && int(counter) != metric.Count.IntValue() {
					run = &r // this is an incomplete metric
				}
			}

		}
	}

	// Use mostRecentTime to update LastScheduleTIme
	if mostRecentTime != nil {
		inAppMetric.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
	} else {
		inAppMetric.Status.LastScheduleTime = nil
	}

	// If user has specified a RunLimit (it will not be zero) delete any old runs that exceed the runLimit
	if inAppMetric.Spec.RunLimit != 0 {
		for i, runs := range metricRunList.Items {
			if int32(i) >= int32(len(metricRunList.Items))-int32(inAppMetric.Spec.RunLimit) {
				break
			}
			if err = r.Delete(context.TODO(), &runs); client.IgnoreNotFound(err) != nil {
				ctrl.Log.Error(err, "unable to delete old run", "run", runs)
			} else {
				ctrl.Log.Info("deleted old run", "run", runs)
			}
		}
	}

	// Get the missedRun and nextRun times
	missedRun, nextRun, err := getNextSchedule(&inAppMetric, r.Now())
	if err != nil {
		ctrl.Log.Error(err, "unable to figure out schedule")
		return ctrl.Result{}, nil
	}

	// Create scheduledResult with the nextRun time
	scheduledResultTime := nextRun.Sub(r.Now())
	scheduledResult := ctrl.Result{RequeueAfter: scheduledResultTime}

	// If missedRun is zero then it is not time to run any query so return the scheduledResult
	if missedRun.IsZero() && run == nil { // if run is not nil, then we need to run more measurements and this means it has been enqueued
		ctrl.Log.Info("no upcoming scheduled times")
		return scheduledResult, nil
	}

	// Create newMetricRun
	if run == nil {
		run = newMetricRun()
		run.Namespace = inAppMetric.Namespace
		timeNow := timeutil.MetaNow().Unix()
		run.Name = "metricrun-" + strconv.FormatInt(timeNow, 10) // name of the metricRun is the time it was created in unix format
		run.Annotations = make(map[string]string)
		run.Annotations[scheduledTimeAnnotation] = missedRun.Format(time.RFC3339) // this annotation is used to later update lastScheduledTime
		updateMetricsSpec(run, inAppMetric.Spec.Metrics)                          // Populate metricRun spec with the metrics from inAppMetric

		// Grab annotations from inAppMetric and populate run's annotations
		for k, v := range inAppMetric.Annotations {
			run.Annotations[k] = v
		}

		// Create metricRun in cluster
		err = r.Client.Create(context.TODO(), run)
		if err != nil {
			ctrl.Log.Error(err, "resource not created")
		}
	}

	// Run measurements for given metrics
	logger := logutil.WithMetricRun(run)
	metricResults, err := runMeasurements(run, inAppMetric.Spec.Metrics)
	if err != nil {
		message := fmt.Sprintf("Unable to resolve metric arguments: %v", err)
		logger.Warn(message)
		run.Status.Phase = argoinappiov1.AnalysisPhaseError
		run.Status.Message = message
		return ctrl.Result{}, err
	}

	// Use measurement results to find status and message and populate metric run runSummary
	newPhase, newMessage, currSummary := assessRunStatus(run, inAppMetric.Spec.Metrics)
	currPhase := run.Status.Phase
	currMessage := run.Status.Message
	if newPhase != run.Status.Phase {
		currPhase = newPhase
		currMessage = newMessage
	}

	err = r.Get(context.TODO(), types.NamespacedName{Name: run.Name, Namespace: run.Namespace}, run)
	if err != nil {
		ctrl.Log.Error(err, "Error getting metricRun instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	now := timeutil.MetaNow()

	// Populate status of metricRun
	status := argoinappiov1.MetricRunStatus{
		Phase:         currPhase,
		Message:       currMessage,
		MetricResults: metricResults,
		StartedAt:     &now,
		RunSummary:    currSummary,
	}

	run.Status = status

	nextReconcileTime := calculateNextReconcileTime(run, inAppMetric.Spec.Metrics)
	if nextReconcileTime != nil {
		enqueueSeconds := nextReconcileTime.Sub(timeutil.Now())
		if enqueueSeconds < 0 {
			enqueueSeconds = 0
		}
		ctrl.Log.Info("Enqueueing:", "enqueue seconds", enqueueSeconds, "scheduled time", scheduledResultTime)
		if enqueueSeconds < scheduledResultTime {
			scheduledResult = ctrl.Result{RequeueAfter: enqueueSeconds}
		}
	}

	// Update status of metricRun
	if err = r.Client.Status().Update(context.TODO(), run); err != nil {
		ctrl.Log.Error(err, "unable to update metric run status")
		return ctrl.Result{}, err
	}

	// Update status of inAppMetric
	if err = r.Status().Update(context.TODO(), &inAppMetric); err != nil {
		ctrl.Log.Error(err, "unable to update InAppMetric status")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Second,
		}, err
	}

	return scheduledResult, nil
}

// calculateNextReconcileTime calculates the next time that this MetricRun should be reconciled,
// based on the earliest time of all metrics intervals, counts, and their finishedAt timestamps
func calculateNextReconcileTime(run *argoinappiov1.MetricRun, metrics []argoinappiov1.Metric) *time.Time {
	var reconcileTime *time.Time
	for _, metric := range metrics {
		if analysisutil.MetricCompleted(run, metric.Name) {
			// NOTE: this also covers the case where metric.Count is reached
			continue
		}
		//logCtx := logutil.WithAnalysisRun(run).WithField("metric", metric.Name)
		lastMeasurement := analysisutil.LastMeasurement(run, metric.Name)
		if lastMeasurement == nil {
			if metric.InitialDelay != "" {
				startTime := timeutil.MetaNow()
				if run.Status.StartedAt != nil {
					startTime = *run.Status.StartedAt
				}
				parsedInterval, err := parseMetricInterval(metric.InitialDelay)
				if err != nil {
					continue
				}
				endInitialDelay := startTime.Add(parsedInterval)
				if reconcileTime == nil || reconcileTime.After(endInitialDelay) {
					reconcileTime = &endInitialDelay
				}
				continue
			}
			// no measurement was started . we should never get here
			ctrl.Log.Info("Metric never started. Not factored into enqueue time.")
			continue
		}
		if lastMeasurement.FinishedAt == nil {
			// unfinished in-flight measurement.
			if lastMeasurement.ResumeAt != nil {
				if reconcileTime == nil || reconcileTime.After(lastMeasurement.ResumeAt.Time) {
					reconcileTime = &lastMeasurement.ResumeAt.Time
				}
			}
			continue
		}
		metricResult := analysisutil.GetResult(run, metric.Name)
		effectiveCount := metric.EffectiveCount()
		if effectiveCount != nil && metricResult.Count >= int32(effectiveCount.IntValue()) {
			// we have reached desired count
			continue
		}
		var interval time.Duration
		if lastMeasurement.Phase == argoinappiov1.AnalysisPhaseError {
			interval = DefaultErrorRetryInterval
		} else if metric.Interval != "" {
			parsedInterval, err := parseMetricInterval(metric.Interval)
			if err != nil {
				continue
			}
			interval = parsedInterval
		} else {
			// if we get here, an interval was not set (meaning reoccurrence was not desired), and
			// there was no error (meaning we don't need to retry). no need to requeue this metric.
			// NOTE: we shouldn't ever get here since it means we are not doing proper bookkeeping
			// of count.
			ctrl.Log.Info("Skipping requeue. No interval or error (count: %d, effectiveCount: %s)", "metricResult count", metricResult.Count) //, metric.EffectiveCount().String())
			continue
		}
		// Take the earliest time of all metrics
		metricReconcileTime := lastMeasurement.FinishedAt.Add(interval)
		if reconcileTime == nil || reconcileTime.After(metricReconcileTime) {
			reconcileTime = &metricReconcileTime
		}
	}
	return reconcileTime
}

// parseMetricInterval is a helper method to parse the given metric interval and return the
// parsed duration or error (if any)
func parseMetricInterval(metricDurationString argoinappiov1.DurationString) (time.Duration, error) {
	metricInterval, err := metricDurationString.Duration()
	if err != nil {
		ctrl.Log.Error(err, "Failed to parse interval: %v")
		return -1, err
	}
	return metricInterval, nil
}

/** Populates given metricRun's metricRun.Spec.Metrics with the give metrics **/
func updateMetricsSpec(run *argoinappiov1.MetricRun, tasks []argoinappiov1.Metric) {
	for _, task := range tasks {
		analysisutil.SetMetrics(run, task)
	}
}

/** Get scheduled time for given run using annotation **/
func getScheduledTimeForRun(run *argoinappiov1.MetricRun) (*time.Time, error) {
	timeRaw := run.Annotations[scheduledTimeAnnotation]
	if len(timeRaw) == 0 {
		return nil, nil
	}

	timeParsed, err := time.Parse(time.RFC3339, timeRaw)
	if err != nil {
		return nil, err
	}
	return &timeParsed, nil
}

/** Get the last missed time and the next time for runs using the cron schedule within inAppMetric **/
func getNextSchedule(metric *argoinappiov1.InAppMetric, now time.Time) (lastMissed time.Time, next time.Time, err error) {
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

	for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
		lastMissed = t
	}
	return lastMissed, sched.Next(now), nil
}

/** Runs measurements for given metrics, and populates metricResults in the status of the given run **/
func runMeasurements(run *argoinappiov1.MetricRun, tasks []argoinappiov1.Metric) ([]argoinappiov1.MetricResult, error) {
	var metricResults []argoinappiov1.MetricResult
	if run.Status.MetricResults == nil {
		metricResults = []argoinappiov1.MetricResult{}
	} else {
		metricResults = run.Status.MetricResults
	}

	for _, task := range tasks {
		e := log.Entry{}

		provider, err := metricproviders.NewProvider(e, task)
		if err != nil {
			return nil, err
		}

		metricResult := analysisutil.GetResult(run, task.Name)
		if metricResult == nil {
			metricResult = &argoinappiov1.MetricResult{
				Name:     task.Name,
				Phase:    argoinappiov1.AnalysisPhaseRunning,
				Metadata: provider.GetMetadata(task),
				Counter:  1,
			}
		} else {
			if task.Count == nil || (task.Count.IntValue() == int(metricResult.Counter)) {
				continue
			} else {
				metricResult.Counter++
			}
		}

		var newMeasurement argoinappiov1.Measurement
		newMeasurement = provider.Run(run, task)

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
		metricResults = analysisutil.SetMetricResult(metricResults, *metricResult)
	}
	return metricResults, nil
}

/** Given a metricRun, computes runSummary, phase and message for given metric tasks **/
func assessRunStatus(run *argoinappiov1.MetricRun, metrics []argoinappiov1.Metric) (argoinappiov1.AnalysisPhase, string, argoinappiov1.RunSummary) {
	var worstStatus argoinappiov1.AnalysisPhase
	var worstMessage string
	terminating := analysisutil.IsTerminating(run)
	everythingCompleted := true

	if run.Status.StartedAt == nil {
		now := timeutil.MetaNow()
		run.Status.StartedAt = &now
	}

	// Initialize Run summary object
	runSummary := argoinappiov1.RunSummary{
		Count:        0,
		Successful:   0,
		Failed:       0,
		Inconclusive: 0,
		Error:        0,
	}

	for _, metric := range metrics {
		runSummary.Count++

		if result := analysisutil.GetResult(run, metric.Name); result != nil {
			logger := logutil.WithMetricRun(run).WithField("metric", metric.Name)
			metricStatus := assessMetricStatus(metric, *result, terminating)
			if result.Phase != metricStatus {
				logger.Infof("Metric '%s' transitioned from %s -> %s", metric.Name, result.Phase, metricStatus)
				if lastMeasurement := analysisutil.LastMeasurement(run, metric.Name); lastMeasurement != nil {
					result.Message = lastMeasurement.Message
				}
				result.Phase = metricStatus
				analysisutil.SetResult(run, *result)
			}
			if !metricStatus.Completed() {
				// if any metric is in-progress, then entire notifications run will be considered running
				everythingCompleted = false
			} else {
				phase, message := assessMetricFailureInconclusiveOrError(metric, *result)
				if worstStatus == "" || analysisutil.IsWorse(worstStatus, metricStatus) {
					worstStatus = metricStatus
					if message != "" {
						worstMessage = fmt.Sprintf("Metric \"%s\" assessed %s due to %s", metric.Name, metricStatus, message)
						if result.Message != "" {
							worstMessage += fmt.Sprintf(": \"Error Message: %s\"", result.Message)
						}
					}
				}
				// Update Run Summary
				switch phase {
				case argoinappiov1.AnalysisPhaseError:
					runSummary.Error++
				case argoinappiov1.AnalysisPhaseFailed:
					runSummary.Failed++
				case argoinappiov1.AnalysisPhaseInconclusive:
					runSummary.Inconclusive++
				case argoinappiov1.AnalysisPhaseSuccessful:
					runSummary.Successful++
				default:
					// We'll mark the status as success by default if it doesn't match anything.
					runSummary.Successful++
				}
			}
		} else {
			everythingCompleted = false
		}
	}

	worstMessage = strings.TrimSpace(worstMessage)
	if terminating {
		if worstStatus == "" {
			// we have yet to take a single measurement, but have already been instructed to stop
			log.Infof(SuccessfulAssessmentRunTerminatedResult)
			return argoinappiov1.AnalysisPhaseSuccessful, worstMessage, runSummary
		}
		log.Infof("Metric Assessment Result - %s: Run Terminated", worstStatus)
		return worstStatus, worstMessage, runSummary
	}
	if !everythingCompleted || worstStatus == "" {
		return argoinappiov1.AnalysisPhaseRunning, "", runSummary
	}
	return worstStatus, worstMessage, runSummary

}

func assessMetricFailureInconclusiveOrError(metric argoinappiov1.Metric, result argoinappiov1.MetricResult) (argoinappiov1.AnalysisPhase, string) {
	var message string
	var phase argoinappiov1.AnalysisPhase

	failureLimit := int32(0)
	if metric.FailureLimit != nil {
		failureLimit = int32(metric.FailureLimit.IntValue())
	}
	if result.Failed > failureLimit {
		phase = argoinappiov1.AnalysisPhaseFailed
		message = fmt.Sprintf("failed (%d) > failureLimit (%d)", result.Failed, failureLimit)
	}

	inconclusiveLimit := int32(0)
	if metric.InconclusiveLimit != nil {
		inconclusiveLimit = int32(metric.InconclusiveLimit.IntValue())
	}
	if result.Inconclusive > inconclusiveLimit {
		phase = argoinappiov1.AnalysisPhaseInconclusive
		message = fmt.Sprintf("inconclusive (%d) > inconclusiveLimit (%d)", result.Inconclusive, inconclusiveLimit)
	}

	consecutiveErrorLimit := defaults.GetConsecutiveErrorLimitOrDefault(&metric)
	if result.ConsecutiveError > consecutiveErrorLimit {
		phase = argoinappiov1.AnalysisPhaseError
		message = fmt.Sprintf("consecutiveErrors (%d) > consecutiveErrorLimit (%d)", result.ConsecutiveError, consecutiveErrorLimit)
	}
	return phase, message
}

// assessMetricStatus assesses the status of a single metric based on:
// * current or latest measurement status
// * parameters given by the metric (failureLimit, count, etc...)
// * whether we are terminating (e.g. due to failing run, or termination request)
func assessMetricStatus(metric argoinappiov1.Metric, result argoinappiov1.MetricResult, terminating bool) argoinappiov1.AnalysisPhase {
	if result.Phase.Completed() {
		return result.Phase
	}
	logger := log.WithField("metric", metric.Name)
	if len(result.Measurements) == 0 {
		if terminating {
			logger.Infof(SuccessfulAssessmentRunTerminatedResult)
			return argoinappiov1.AnalysisPhasePending
		}
	}
	lastMeasurement := result.Measurements[len(result.Measurements)-1]
	if !lastMeasurement.Phase.Completed() {
		// we still have an in-flight measurement
		return argoinappiov1.AnalysisPhaseRunning
	}
	// Check if metric was considered Failed, Inconclusive, or Error
	// If true, then return AnalysisRunPhase as Failed, Inconclusive, or Error respectively
	phaseFailureInconclusiveOrError, message := assessMetricFailureInconclusiveOrError(metric, result)
	if phaseFailureInconclusiveOrError != "" {
		logger.Infof("Metric Assessment Result - %s: %s", phaseFailureInconclusiveOrError, message)
		return phaseFailureInconclusiveOrError
	}

	// If a count was specified, and we reached that count, then metric is considered Successful.
	// The Error, Failed, Inconclusive counters are ignored because those checks have already been
	// taken into consideration above, and we do not want to fail if failures < failureLimit.
	effectiveCount := metric.EffectiveCount()
	if effectiveCount != nil && result.Count >= int32(effectiveCount.IntValue()) {
		logger.Infof("Metric Assessment Result - %s: Count (%s) Reached", argoinappiov1.AnalysisPhaseSuccessful, effectiveCount.String())
		return argoinappiov1.AnalysisPhaseSuccessful
	}
	// if we get here, this metric runs indefinitely
	if terminating {
		logger.Infof(SuccessfulAssessmentRunTerminatedResult)
		return argoinappiov1.AnalysisPhaseSuccessful
	}
	return argoinappiov1.AnalysisPhaseRunning
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
