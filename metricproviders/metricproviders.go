package metricproviders

import (
	"fmt"

	"monitoring/metricproviders/cloudwatch"
	"monitoring/metricproviders/datadog"
	"monitoring/metricproviders/graphite"
	"monitoring/metricproviders/kayenta"
	"monitoring/metricproviders/newrelic"
	"monitoring/metricproviders/prometheus"
	"monitoring/metricproviders/wavefront"
	"monitoring/metricproviders/webmetric"

	log "github.com/sirupsen/logrus"

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

// NewProvider creates the correct provider based on the provider type of the Metric
func NewProvider(logCtx log.Entry, metric v1alpha1.Metric) (Provider, error) {
	switch provider := Type(metric); provider {
	case prometheus.ProviderType:
		api, err := prometheus.NewPrometheusAPI(metric)
		if err != nil {
			return nil, err
		}
		return prometheus.NewPrometheusProvider(api, logCtx), nil
	/*case job.ProviderType:
	return job.NewJobProvider(logCtx, f.KubeClient, f.JobLister), nil*/
	case kayenta.ProviderType:
		c := kayenta.NewHttpClient()
		return kayenta.NewKayentaProvider(logCtx, c), nil
	case webmetric.ProviderType:
		c := webmetric.NewWebMetricHttpClient(metric)
		p, err := webmetric.NewWebMetricJsonParser(metric)
		if err != nil {
			return nil, err
		}
		return webmetric.NewWebMetricProvider(logCtx, c, p), nil
	/*case datadog.ProviderType:
		return datadog.NewDatadogProvider(logCtx, f.KubeClient)
	case wavefront.ProviderType:
		client, err := wavefront.NewWavefrontAPI(metric, f.KubeClient)
		if err != nil {
			return nil, err
		}
		return wavefront.NewWavefrontProvider(client, logCtx), nil
	case newrelic.ProviderType:
		client, err := newrelic.NewNewRelicAPIClient(metric, f.KubeClient)
		if err != nil {
			return nil, err
		}
		return newrelic.NewNewRelicProvider(client, logCtx), nil*/
	case graphite.ProviderType:
		client, err := graphite.NewAPIClient(metric, logCtx)
		if err != nil {
			return nil, err
		}
		return graphite.NewGraphiteProvider(client, logCtx), nil
	case cloudwatch.ProviderType:
		clinet, err := cloudwatch.NewCloudWatchAPIClient(metric)
		if err != nil {
			return nil, err
		}
		return cloudwatch.NewCloudWatchProvider(clinet, logCtx), nil
	default:
		return nil, fmt.Errorf("no valid provider in metric '%s'", metric.Name)
	}
}

func Type(metric v1alpha1.Metric) string {
	if metric.Provider.Prometheus != nil {
		return prometheus.ProviderType
		/*} else if metric.Provider.Job != nil {
		return job.ProviderType*/
	} else if metric.Provider.Kayenta != nil {
		return kayenta.ProviderType
	} else if metric.Provider.Web != nil {
		return webmetric.ProviderType
	} else if metric.Provider.Datadog != nil {
		return datadog.ProviderType
	} else if metric.Provider.Wavefront != nil {
		return wavefront.ProviderType
	} else if metric.Provider.NewRelic != nil {
		return newrelic.ProviderType
	} else if metric.Provider.CloudWatch != nil {
		return cloudwatch.ProviderType
	}
	return "Unknown Provider"
}
