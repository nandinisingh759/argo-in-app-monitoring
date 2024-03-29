package analysis

import (
	argoinappiov1 "monitoring/api/v1"
)

// GetResult returns the metric result by name
func GetResult(run *argoinappiov1.MetricRun, metricName string) *argoinappiov1.MetricResult {
	for _, result := range run.Status.MetricResults {
		if result.Name == metricName {
			return &result
		}
	}
	return nil
}

// SetMetricResult updates the metric result
func SetMetricResult(metricResults []argoinappiov1.MetricResult, result argoinappiov1.MetricResult) []argoinappiov1.MetricResult {
	for i, r := range metricResults {
		if r.Name == result.Name {
			metricResults[i] = result
			return metricResults
		}
	}
	metricResults = append(metricResults, result)
	return metricResults
}

// SetResult updates the metric result
func SetResult(run *argoinappiov1.MetricRun, result argoinappiov1.MetricResult) {
	for i, r := range run.Status.MetricResults {
		if r.Name == result.Name {
			run.Status.MetricResults[i] = result
			return
		}
	}
	run.Status.MetricResults = append(run.Status.MetricResults, result)
}

func SetMetrics(run *argoinappiov1.MetricRun, metric argoinappiov1.Metric) {
	for i, r := range run.Spec.Metrics {
		if r.Name == metric.Name {
			run.Spec.Metrics[i] = metric
			return
		}
	}
	run.Spec.Metrics = append(run.Spec.Metrics, metric)
}

// MetricCompleted returns whether or not a metric was completed or not
func MetricCompleted(run *argoinappiov1.MetricRun, metricName string) bool {
	if result := GetResult(run, metricName); result != nil {
		return result.Phase.Completed()
	}
	return false
}

// IsTerminating returns whether or not the notifications run is terminating, either because a terminate
// was requested explicitly, or because a metric has already measured Failed, Error, or Inconclusive
// which causes the run to end prematurely.
func IsTerminating(run *argoinappiov1.MetricRun) bool {
	for _, res := range run.Status.MetricResults {
		switch res.Phase {
		case argoinappiov1.AnalysisPhaseFailed, argoinappiov1.AnalysisPhaseError, argoinappiov1.AnalysisPhaseInconclusive:
			return true
		}
	}
	return false
}

// LastMeasurement returns the last measurement started or completed for a specific metric
func LastMeasurement(run *argoinappiov1.MetricRun, metricName string) *argoinappiov1.Measurement {
	if result := GetResult(run, metricName); result != nil {
		totalMeasurements := len(result.Measurements)
		if totalMeasurements == 0 {
			return nil
		}
		return &result.Measurements[totalMeasurements-1]
	}
	return nil
}

// analysisStatusOrder is a list of completed notifications sorted in best to worst condition
var analysisStatusOrder = []argoinappiov1.AnalysisPhase{
	argoinappiov1.AnalysisPhaseSuccessful,
	argoinappiov1.AnalysisPhaseRunning,
	argoinappiov1.AnalysisPhasePending,
	argoinappiov1.AnalysisPhaseInconclusive,
	argoinappiov1.AnalysisPhaseError,
	argoinappiov1.AnalysisPhaseFailed,
}

// IsBad returns whether or not the new health status code warrants a notification to the user.
func IsBad(new argoinappiov1.AnalysisPhase) bool {
	if new == analysisStatusOrder[3] || new == analysisStatusOrder[4] || new == analysisStatusOrder[5] {
		return true
	}
	return false
}
