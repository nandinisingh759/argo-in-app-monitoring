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

// MetricCompleted returns whether or not a metric was completed or not
func MetricCompleted(run *argoinappiov1.MetricRun, metricName string) bool {
	if result := GetResult(run, metricName); result != nil {
		return result.Phase.Completed()
	}
	return false
}

// IsTerminating returns whether or not the analysis run is terminating, either because a terminate
// was requested explicitly, or because a metric has already measured Failed, Error, or Inconclusive
// which causes the run to end prematurely.
func IsTerminating(run *argoinappiov1.MetricRun) bool {
	if run.Spec.Terminate {
		return true
	}
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

// analysisStatusOrder is a list of completed analysis sorted in best to worst condition
var analysisStatusOrder = []argoinappiov1.AnalysisPhase{
	argoinappiov1.AnalysisPhaseSuccessful,
	argoinappiov1.AnalysisPhaseRunning,
	argoinappiov1.AnalysisPhasePending,
	argoinappiov1.AnalysisPhaseInconclusive,
	argoinappiov1.AnalysisPhaseError,
	argoinappiov1.AnalysisPhaseFailed,
}

// IsWorse returns whether or not the new health status code is a worser condition than the current.
// Both statuses must be already completed
func IsWorse(current, new argoinappiov1.AnalysisPhase) bool {
	currentIndex := 0
	newIndex := 0
	for i, code := range analysisStatusOrder {
		if current == code {
			currentIndex = i
		}
		if new == code {
			newIndex = i
		}
	}
	return newIndex > currentIndex
}
