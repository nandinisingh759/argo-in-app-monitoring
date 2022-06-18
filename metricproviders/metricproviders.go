package metricproviders

import (
	"fmt"

	"monitoring/metricproviders/prometheus"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"

	v1alpha1 "monitoring/api/v1"
)

// Provider methods to query a external systems and generate a measurement
type Provider interface {
	// Run start a new external system call for a measurement
	// Should be idempotent and do nothing if a call has already been started
	Run(*v1alpha1.MetricRun, v1alpha1.Metric) v1alpha1.Measurement
	// Checks if the external system call is finished and returns the current measurement
	Resume(*v1alpha1.MetricRun, v1alpha1.Metric, v1alpha1.Measurement) v1alpha1.Measurement
	// Terminate will terminate an in-progress measurement
	Terminate(*v1alpha1.MetricRun, v1alpha1.Metric, v1alpha1.Measurement) v1alpha1.Measurement
	// GarbageCollect is used to garbage collect completed measurements to the specified limit
	GarbageCollect(*v1alpha1.MetricRun, v1alpha1.Metric, int) error
	// Type gets the provider type
	Type() string
	// GetMetadata returns any additional metadata which providers need to store/display as part
	// of the metric result. For example, Prometheus uses is to store the final resolved queries.
	GetMetadata(metric v1alpha1.Metric) map[string]string
}

type ProviderFactory struct {
	KubeClient kubernetes.Interface
	JobLister  batchlisters.JobLister
}

type ProviderFactoryFunc func(logCtx log.Entry, metric v1alpha1.Metric) (Provider, error)

// NewProvider creates the correct provider based on the provider type of the Metric
func (f *ProviderFactory) NewProvider(logCtx log.Entry, metric v1alpha1.Metric) (Provider, error) {
	switch provider := Type(metric); provider {
	case prometheus.ProviderType:
		api, err := prometheus.NewPrometheusAPI(metric)
		if err != nil {
			return nil, err
		}
		return prometheus.NewPrometheusProvider(api, logCtx), nil
	default:
		return nil, fmt.Errorf("no valid provider in metric '%s'", metric.Name)
	}
}

func Type(metric v1alpha1.Metric) string {
	if metric.Provider.Prometheus != nil {
		return prometheus.ProviderType
	}
	return "Unknown Provider"
}
