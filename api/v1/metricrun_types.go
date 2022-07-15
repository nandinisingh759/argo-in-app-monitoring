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

package v1

import (
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetricRunSpec defines the desired state of MetricRun
type MetricRunSpec struct {
	// Array of metrics from inAppMetric Spec
	Metrics []Metric `json:"metrics,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,1,rep,name=metrics"`
}

// Metric defines a metric in which to perform analysis
type Metric struct {
	// Name is the name of the metric
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Interval defines an interval string (e.g. 30s, 5m, 1h) between each measurement.
	// If omitted, will perform a single measurement
	Interval DurationString `json:"interval,omitempty" protobuf:"bytes,2,opt,name=interval,casttype=DurationString"`
	// InitialDelay how long the AnalysisRun should wait before starting this metric
	InitialDelay DurationString `json:"initialDelay,omitempty" protobuf:"bytes,3,opt,name=initialDelay,casttype=DurationString"`
	// Count is the number of times to run the measurement. If both interval and count are omitted,
	// the effective count is 1. If only interval is specified, metric runs indefinitely.
	// If count > 1, interval must be specified.
	Count *intstrutil.IntOrString `json:"count,omitempty" protobuf:"bytes,4,opt,name=count"`
	// SuccessCondition is an expression which determines if a measurement is considered successful
	// Expression is a goevaluate expression. The keyword `result` is a variable reference to the
	// value of measurement. Results can be both structured data or primitive.
	// Examples:
	//   result > 10
	//   (result.requests_made * result.requests_succeeded / 100) >= 90
	SuccessCondition string `json:"successCondition,omitempty" protobuf:"bytes,5,opt,name=successCondition"`
	// FailureCondition is an expression which determines if a measurement is considered failed
	// If both success and failure conditions are specified, and the measurement does not fall into
	// either condition, the measurement is considered Inconclusive
	FailureCondition string `json:"failureCondition,omitempty" protobuf:"bytes,6,opt,name=failureCondition"`
	// FailureLimit is the maximum number of times the measurement is allowed to fail, before the
	// entire metric is considered Failed (default: 0)
	FailureLimit *intstrutil.IntOrString `json:"failureLimit,omitempty" protobuf:"bytes,7,opt,name=failureLimit"`
	// InconclusiveLimit is the maximum number of times the measurement is allowed to measure
	// Inconclusive, before the entire metric is considered Inconclusive (default: 0)
	InconclusiveLimit *intstrutil.IntOrString `json:"inconclusiveLimit,omitempty" protobuf:"bytes,8,opt,name=inconclusiveLimit"`
	// ConsecutiveErrorLimit is the maximum number of times the measurement is allowed to error in
	// succession, before the metric is considered error (default: 4)
	ConsecutiveErrorLimit *intstrutil.IntOrString `json:"consecutiveErrorLimit,omitempty" protobuf:"bytes,9,opt,name=consecutiveErrorLimit"`
	// Provider configuration to the external system to use to verify the analysis
	Provider MetricProvider `json:"provider" protobuf:"bytes,10,opt,name=provider"`
}

type MetricProvider struct {
	// Prometheus specifies the prometheus metric to query
	Prometheus *PrometheusMetric `json:"prometheus,omitempty" protobuf:"bytes,1,opt,name=prometheus"`
	// Kayenta specifies a Kayenta metric
	Kayenta *KayentaMetric `json:"kayenta,omitempty" protobuf:"bytes,2,opt,name=kayenta"`
	// Web specifies a generic HTTP web metric
	Web *WebMetric `json:"web,omitempty" protobuf:"bytes,3,opt,name=web"`
	// Datadog specifies a datadog metric to query
	Datadog *DatadogMetric `json:"datadog,omitempty" protobuf:"bytes,4,opt,name=datadog"`
	// Wavefront specifies the wavefront metric to query
	Wavefront *WavefrontMetric `json:"wavefront,omitempty" protobuf:"bytes,5,opt,name=wavefront"`
	// NewRelic specifies the newrelic metric to query
	NewRelic *NewRelicMetric `json:"newRelic,omitempty" protobuf:"bytes,6,opt,name=newRelic"`
	// Job specifies the job metric run
	Job *JobMetric `json:"job,omitempty" protobuf:"bytes,7,opt,name=job"`
	// CloudWatch specifies the cloudWatch metric to query
	CloudWatch *CloudWatchMetric `json:"cloudWatch,omitempty" protobuf:"bytes,8,opt,name=cloudWatch"`
	// Graphite specifies the Graphite metric to query
	Graphite *GraphiteMetric `json:"graphite,omitempty" protobuf:"bytes,9,opt,name=graphite"`
}

// AnalysisPhase is the overall phase of an AnalysisRun, MetricResult, or Measurement
type AnalysisPhase string

// Possible AnalysisPhase values
const (
	AnalysisPhasePending      AnalysisPhase = "Pending"
	AnalysisPhaseRunning      AnalysisPhase = "Running"
	AnalysisPhaseSuccessful   AnalysisPhase = "Successful"
	AnalysisPhaseFailed       AnalysisPhase = "Failed"
	AnalysisPhaseError        AnalysisPhase = "Error"
	AnalysisPhaseInconclusive AnalysisPhase = "Inconclusive"
)

// Completed returns whether or not the analysis status is considered completed
func (as AnalysisPhase) Completed() bool {
	switch as {
	case AnalysisPhaseSuccessful, AnalysisPhaseFailed, AnalysisPhaseError, AnalysisPhaseInconclusive:
		return true
	}
	return false
}

type PrometheusMetric struct {
	// Address is the HTTP address and port of the prometheus server
	Address string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
	// Query is a raw prometheus query to perform
	Query string `json:"query,omitempty" protobuf:"bytes,2,opt,name=query"`
}

// WavefrontMetric defines the wavefront query to perform canary analysis
type WavefrontMetric struct {
	// Address is the HTTP address and port of the wavefront server
	Address string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
	// Query is a raw wavefront query to perform
	Query string `json:"query,omitempty" protobuf:"bytes,2,opt,name=query"`
}

// NewRelicMetric defines the newrelic query to perform canary analysis
type NewRelicMetric struct {
	// Profile is the name of the secret holding NR account configuration
	Profile string `json:"profile,omitempty" protobuf:"bytes,1,opt,name=profile"`
	// Query is a raw newrelic NRQL query to perform
	Query string `json:"query" protobuf:"bytes,2,opt,name=query"`
}

// JobMetric defines a job to run which acts as a metric
type JobMetric struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec     batchv1.JobSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
}

// GraphiteMetric defines the Graphite query to perform canary analysis
type GraphiteMetric struct {
	// Address is the HTTP address and port of the Graphite server
	Address string `json:"address,omitempty" protobuf:"bytes,1,opt,name=address"`
	// Query is a raw Graphite query to perform
	Query string `json:"query,omitempty" protobuf:"bytes,2,opt,name=query"`
}

// CloudWatchMetric defines the cloudwatch query to perform canary analysis
type CloudWatchMetric struct {
	Interval          DurationString              `json:"interval,omitempty" protobuf:"bytes,1,opt,name=interval,casttype=DurationString"`
	MetricDataQueries []CloudWatchMetricDataQuery `json:"metricDataQueries" protobuf:"bytes,2,rep,name=metricDataQueries"`
}

// CloudWatchMetricDataQuery defines the cloudwatch query
type CloudWatchMetricDataQuery struct {
	Id         string                  `json:"id,omitempty" protobuf:"bytes,1,opt,name=id"`
	Expression *string                 `json:"expression,omitempty" protobuf:"bytes,2,opt,name=expression"`
	Label      *string                 `json:"label,omitempty" protobuf:"bytes,3,opt,name=label"`
	MetricStat *CloudWatchMetricStat   `json:"metricStat,omitempty" protobuf:"bytes,4,opt,name=metricStat"`
	Period     *intstrutil.IntOrString `json:"period,omitempty" protobuf:"varint,5,opt,name=period"`
	ReturnData *bool                   `json:"returnData,omitempty" protobuf:"bytes,6,opt,name=returnData"`
}

type CloudWatchMetricStat struct {
	Metric CloudWatchMetricStatMetric `json:"metric,omitempty" protobuf:"bytes,1,opt,name=metric"`
	Period intstrutil.IntOrString     `json:"period,omitempty" protobuf:"varint,2,opt,name=period"`
	Stat   string                     `json:"stat,omitempty" protobuf:"bytes,3,opt,name=stat"`
	Unit   string                     `json:"unit,omitempty" protobuf:"bytes,4,opt,name=unit"`
}

type CloudWatchMetricStatMetric struct {
	Dimensions []CloudWatchMetricStatMetricDimension `json:"dimensions,omitempty" protobuf:"bytes,1,rep,name=dimensions"`
	MetricName string                                `json:"metricName,omitempty" protobuf:"bytes,2,opt,name=metricName"`
	Namespace  *string                               `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
}

type CloudWatchMetricStatMetricDimension struct {
	Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Value string `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
}

type KayentaMetric struct {
	Address string `json:"address" protobuf:"bytes,1,opt,name=address"`

	Application string `json:"application" protobuf:"bytes,2,opt,name=application"`

	CanaryConfigName string `json:"canaryConfigName" protobuf:"bytes,3,opt,name=canaryConfigName"`

	MetricsAccountName       string `json:"metricsAccountName" protobuf:"bytes,4,opt,name=metricsAccountName"`
	ConfigurationAccountName string `json:"configurationAccountName" protobuf:"bytes,5,opt,name=configurationAccountName"`
	StorageAccountName       string `json:"storageAccountName" protobuf:"bytes,6,opt,name=storageAccountName"`

	Threshold KayentaThreshold `json:"threshold" protobuf:"bytes,7,opt,name=threshold"`

	Scopes []KayentaScope `json:"scopes" protobuf:"bytes,8,rep,name=scopes"`
}

type KayentaThreshold struct {
	Pass     int64 `json:"pass" protobuf:"varint,1,opt,name=pass"`
	Marginal int64 `json:"marginal" protobuf:"varint,2,opt,name=marginal"`
}

type KayentaScope struct {
	Name            string      `json:"name" protobuf:"bytes,1,opt,name=name"`
	ControlScope    ScopeDetail `json:"controlScope" protobuf:"bytes,2,opt,name=controlScope"`
	ExperimentScope ScopeDetail `json:"experimentScope" protobuf:"bytes,3,opt,name=experimentScope"`
}

type ScopeDetail struct {
	Scope  string `json:"scope" protobuf:"bytes,1,opt,name=scope"`
	Region string `json:"region" protobuf:"bytes,2,opt,name=region"`
	Step   int64  `json:"step" protobuf:"varint,3,opt,name=step"`
	Start  string `json:"start" protobuf:"bytes,4,opt,name=start"`
	End    string `json:"end" protobuf:"bytes,5,opt,name=end"`
}

type WebMetric struct {
	// Method is the method of the web metric (empty defaults to GET)
	Method WebMetricMethod `json:"method,omitempty" protobuf:"bytes,1,opt,name=method"`
	// URL is the address of the web metric
	URL string `json:"url" protobuf:"bytes,2,opt,name=url"`
	// +patchMergeKey=key
	// +patchStrategy=merge
	// Headers are optional HTTP headers to use in the request
	Headers []WebMetricHeader `json:"headers,omitempty" patchStrategy:"merge" patchMergeKey:"key" protobuf:"bytes,3,rep,name=headers"`
	// Body is the body of the we metric (must be POST/PUT)
	Body string `json:"body,omitempty" protobuf:"bytes,4,opt,name=body"`
	// TimeoutSeconds is the timeout for the request in seconds (default: 10)
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty" protobuf:"varint,5,opt,name=timeoutSeconds"`
	// JSONPath is a JSON Path to use as the result variable (default: "{$}")
	JSONPath string `json:"jsonPath,omitempty" protobuf:"bytes,6,opt,name=jsonPath"`
	// Insecure skips host TLS verification
	Insecure bool `json:"insecure,omitempty" protobuf:"varint,7,opt,name=insecure"`
}

// WebMetricMethod is the available HTTP methods
type WebMetricMethod string

// Possible HTTP method values
const (
	WebMetricMethodGet  WebMetricMethod = "GET"
	WebMetricMethodPost WebMetricMethod = "POST"
	WebMetricMethodPut  WebMetricMethod = "PUT"
)

type WebMetricHeader struct {
	Key   string `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type DatadogMetric struct {
	Interval DurationString `json:"interval,omitempty" protobuf:"bytes,1,opt,name=interval,casttype=DurationString"`
	Query    string         `json:"query" protobuf:"bytes,2,opt,name=query"`
}

// DurationString is a string representing a duration (e.g. 30s, 5m, 1h)
type DurationString string

// Duration converts DurationString into a time.Duration
func (d DurationString) Duration() (time.Duration, error) {
	return time.ParseDuration(string(d))
}

// MetricRunStatus defines the observed state of MetricRun
type MetricRunStatus struct {
	// Phase is the status of the analysis run
	Phase AnalysisPhase `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=AnalysisPhase"`
	// Message is a message explaining current status
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// MetricResults contains the metrics collected during the run
	MetricResults []MetricResult `json:"metricResults,omitempty" protobuf:"bytes,3,rep,name=metricResults"`
	// StartedAt indicates when the analysisRun first started
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,4,opt,name=startedAt"`
	// RunSummary contains the final results from the metric executions
	RunSummary RunSummary `json:"runSummary,omitempty" protobuf:"bytes,5,opt,name=runSummary"`
}

// RunSummary contains the final results from the metric executions
type RunSummary struct {
	// This is equal to the sum of Successful, Failed, Inconclusive
	Count int32 `json:"count,omitempty" protobuf:"varint,1,opt,name=count"`
	// Successful is the number of times the metric was measured Successful
	Successful int32 `json:"successful,omitempty" protobuf:"varint,2,opt,name=successful"`
	// Failed is the number of times the metric was measured Failed
	Failed int32 `json:"failed,omitempty" protobuf:"varint,3,opt,name=failed"`
	// Inconclusive is the number of times the metric was measured Inconclusive
	Inconclusive int32 `json:"inconclusive,omitempty" protobuf:"varint,4,opt,name=inconclusive"`
	// Error is the number of times an error was encountered during measurement
	Error int32 `json:"error,omitempty" protobuf:"varint,5,opt,name=error"`
}

// MetricResult contain a list of the most recent measurements for a single metric along with
// counters on how often the measurement
type MetricResult struct {
	// Name is the name of the metric
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Phase is the overall aggregate status of the metric
	Phase AnalysisPhase `json:"phase" protobuf:"bytes,2,opt,name=phase,casttype=AnalysisPhase"`
	// Measurements holds the most recent measurements collected for the metric
	Measurements []Measurement `json:"measurements,omitempty" protobuf:"bytes,3,rep,name=measurements"`
	// Message contains a message describing current condition (e.g. error messages)
	Message string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// Count is the number of times the metric was measured without Error
	// This is equal to the sum of Successful, Failed, Inconclusive
	Count int32 `json:"count,omitempty" protobuf:"varint,5,opt,name=count"`
	// Successful is the number of times the metric was measured Successful
	Successful int32 `json:"successful,omitempty" protobuf:"varint,6,opt,name=successful"`
	// Failed is the number of times the metric was measured Failed
	Failed int32 `json:"failed,omitempty" protobuf:"varint,7,opt,name=failed"`
	// Inconclusive is the number of times the metric was measured Inconclusive
	Inconclusive int32 `json:"inconclusive,omitempty" protobuf:"varint,8,opt,name=inconclusive"`
	// Error is the number of times an error was encountered during measurement
	Error int32 `json:"error,omitempty" protobuf:"varint,9,opt,name=error"`
	// ConsecutiveError is the number of times an error was encountered during measurement in succession
	// Resets to zero when non-errors are encountered
	ConsecutiveError int32 `json:"consecutiveError,omitempty" protobuf:"varint,10,opt,name=consecutiveError"`
	// DryRun indicates whether this metric is running in a dry-run mode or not
	DryRun bool `json:"dryRun,omitempty" protobuf:"varint,11,opt,name=dryRun"`
	// Metadata stores additional metadata about this metric. It is used by different providers to store
	// the final state which gets used while taking measurements. For example, Prometheus uses this field
	// to store the final resolved query after substituting the template arguments.
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,12,rep,name=metadata"`

	// ADD IN DESCRIPTION
	Counter   int32 `json:"currCount,omitempty"`
	CountFlag bool  `json:"countFlag,omitempty"`
}

// Measurement is a point in time result value of a single metric, and the time it was measured
type Measurement struct {
	// Phase is the status of this single measurement
	Phase AnalysisPhase `json:"phase" protobuf:"bytes,1,opt,name=phase,casttype=AnalysisPhase"`
	// Message contains a message describing current condition (e.g. error messages)
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// StartedAt is the timestamp in which this measurement started to be measured
	StartedAt *metav1.Time `json:"startedAt,omitempty" protobuf:"bytes,3,opt,name=startedAt"`
	// FinishedAt is the timestamp in which this measurement completed and value was collected
	FinishedAt *metav1.Time `json:"finishedAt,omitempty" protobuf:"bytes,4,opt,name=finishedAt"`
	// Value is the measured value of the metric
	Value string `json:"value,omitempty" protobuf:"bytes,5,opt,name=value"`
	// Metadata stores additional metadata about this metric result, used by the different providers
	// (e.g. kayenta run ID, job name)
	Metadata map[string]string `json:"metadata,omitempty" protobuf:"bytes,6,rep,name=metadata"`
	// ResumeAt is the  timestamp when the analysisRun should try to resume the measurement
	ResumeAt *metav1.Time `json:"resumeAt,omitempty" protobuf:"bytes,7,opt,name=resumeAt"`
}

func (m *Metric) EffectiveCount() *intstrutil.IntOrString {
	// Need to check if type is String
	if m.Count == nil || m.Count.IntValue() == 0 {
		if m.Interval == "" {
			one := intstrutil.FromInt(1)
			return &one
		}
		return nil
	}
	return m.Count
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MetricRun is the Schema for the metricruns API
type MetricRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricRunSpec   `json:"spec,omitempty"`
	Status MetricRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MetricRunList contains a list of MetricRun
type MetricRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetricRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetricRun{}, &MetricRunList{})
}
