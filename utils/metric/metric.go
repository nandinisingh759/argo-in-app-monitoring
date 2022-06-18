package metric

import (
	v1alpha1 "monitoring/api/v1"

	timeutil "monitoring/utils/time"
)

// MarkMeasurementError sets an error message on a measurement along with finish time
func MarkMeasurementError(m v1alpha1.Measurement, err error) v1alpha1.Measurement {
	m.Phase = v1alpha1.AnalysisPhaseError
	m.Message = err.Error()
	if m.FinishedAt == nil {
		finishedTime := timeutil.MetaNow()
		m.FinishedAt = &finishedTime
	}
	return m
}
