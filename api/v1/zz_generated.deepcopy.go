//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudWatchMetric) DeepCopyInto(out *CloudWatchMetric) {
	*out = *in
	if in.MetricDataQueries != nil {
		in, out := &in.MetricDataQueries, &out.MetricDataQueries
		*out = make([]CloudWatchMetricDataQuery, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudWatchMetric.
func (in *CloudWatchMetric) DeepCopy() *CloudWatchMetric {
	if in == nil {
		return nil
	}
	out := new(CloudWatchMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudWatchMetricDataQuery) DeepCopyInto(out *CloudWatchMetricDataQuery) {
	*out = *in
	if in.Expression != nil {
		in, out := &in.Expression, &out.Expression
		*out = new(string)
		**out = **in
	}
	if in.Label != nil {
		in, out := &in.Label, &out.Label
		*out = new(string)
		**out = **in
	}
	if in.MetricStat != nil {
		in, out := &in.MetricStat, &out.MetricStat
		*out = new(CloudWatchMetricStat)
		(*in).DeepCopyInto(*out)
	}
	if in.Period != nil {
		in, out := &in.Period, &out.Period
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.ReturnData != nil {
		in, out := &in.ReturnData, &out.ReturnData
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudWatchMetricDataQuery.
func (in *CloudWatchMetricDataQuery) DeepCopy() *CloudWatchMetricDataQuery {
	if in == nil {
		return nil
	}
	out := new(CloudWatchMetricDataQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudWatchMetricStat) DeepCopyInto(out *CloudWatchMetricStat) {
	*out = *in
	in.Metric.DeepCopyInto(&out.Metric)
	out.Period = in.Period
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudWatchMetricStat.
func (in *CloudWatchMetricStat) DeepCopy() *CloudWatchMetricStat {
	if in == nil {
		return nil
	}
	out := new(CloudWatchMetricStat)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudWatchMetricStatMetric) DeepCopyInto(out *CloudWatchMetricStatMetric) {
	*out = *in
	if in.Dimensions != nil {
		in, out := &in.Dimensions, &out.Dimensions
		*out = make([]CloudWatchMetricStatMetricDimension, len(*in))
		copy(*out, *in)
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudWatchMetricStatMetric.
func (in *CloudWatchMetricStatMetric) DeepCopy() *CloudWatchMetricStatMetric {
	if in == nil {
		return nil
	}
	out := new(CloudWatchMetricStatMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudWatchMetricStatMetricDimension) DeepCopyInto(out *CloudWatchMetricStatMetricDimension) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudWatchMetricStatMetricDimension.
func (in *CloudWatchMetricStatMetricDimension) DeepCopy() *CloudWatchMetricStatMetricDimension {
	if in == nil {
		return nil
	}
	out := new(CloudWatchMetricStatMetricDimension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatadogMetric) DeepCopyInto(out *DatadogMetric) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatadogMetric.
func (in *DatadogMetric) DeepCopy() *DatadogMetric {
	if in == nil {
		return nil
	}
	out := new(DatadogMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GraphiteMetric) DeepCopyInto(out *GraphiteMetric) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GraphiteMetric.
func (in *GraphiteMetric) DeepCopy() *GraphiteMetric {
	if in == nil {
		return nil
	}
	out := new(GraphiteMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InAppMetric) DeepCopyInto(out *InAppMetric) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InAppMetric.
func (in *InAppMetric) DeepCopy() *InAppMetric {
	if in == nil {
		return nil
	}
	out := new(InAppMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InAppMetric) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InAppMetricList) DeepCopyInto(out *InAppMetricList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]InAppMetric, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InAppMetricList.
func (in *InAppMetricList) DeepCopy() *InAppMetricList {
	if in == nil {
		return nil
	}
	out := new(InAppMetricList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InAppMetricList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InAppMetricSpec) DeepCopyInto(out *InAppMetricSpec) {
	*out = *in
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]Metric, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InAppMetricSpec.
func (in *InAppMetricSpec) DeepCopy() *InAppMetricSpec {
	if in == nil {
		return nil
	}
	out := new(InAppMetricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InAppMetricStatus) DeepCopyInto(out *InAppMetricStatus) {
	*out = *in
	if in.LastScheduleTime != nil {
		in, out := &in.LastScheduleTime, &out.LastScheduleTime
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InAppMetricStatus.
func (in *InAppMetricStatus) DeepCopy() *InAppMetricStatus {
	if in == nil {
		return nil
	}
	out := new(InAppMetricStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JobMetric) DeepCopyInto(out *JobMetric) {
	*out = *in
	in.Metadata.DeepCopyInto(&out.Metadata)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JobMetric.
func (in *JobMetric) DeepCopy() *JobMetric {
	if in == nil {
		return nil
	}
	out := new(JobMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KayentaMetric) DeepCopyInto(out *KayentaMetric) {
	*out = *in
	out.Threshold = in.Threshold
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]KayentaScope, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KayentaMetric.
func (in *KayentaMetric) DeepCopy() *KayentaMetric {
	if in == nil {
		return nil
	}
	out := new(KayentaMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KayentaScope) DeepCopyInto(out *KayentaScope) {
	*out = *in
	out.ControlScope = in.ControlScope
	out.ExperimentScope = in.ExperimentScope
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KayentaScope.
func (in *KayentaScope) DeepCopy() *KayentaScope {
	if in == nil {
		return nil
	}
	out := new(KayentaScope)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KayentaThreshold) DeepCopyInto(out *KayentaThreshold) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KayentaThreshold.
func (in *KayentaThreshold) DeepCopy() *KayentaThreshold {
	if in == nil {
		return nil
	}
	out := new(KayentaThreshold)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Measurement) DeepCopyInto(out *Measurement) {
	*out = *in
	if in.StartedAt != nil {
		in, out := &in.StartedAt, &out.StartedAt
		*out = (*in).DeepCopy()
	}
	if in.FinishedAt != nil {
		in, out := &in.FinishedAt, &out.FinishedAt
		*out = (*in).DeepCopy()
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ResumeAt != nil {
		in, out := &in.ResumeAt, &out.ResumeAt
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Measurement.
func (in *Measurement) DeepCopy() *Measurement {
	if in == nil {
		return nil
	}
	out := new(Measurement)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metric) DeepCopyInto(out *Metric) {
	*out = *in
	if in.Count != nil {
		in, out := &in.Count, &out.Count
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.FailureLimit != nil {
		in, out := &in.FailureLimit, &out.FailureLimit
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.InconclusiveLimit != nil {
		in, out := &in.InconclusiveLimit, &out.InconclusiveLimit
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.ConsecutiveErrorLimit != nil {
		in, out := &in.ConsecutiveErrorLimit, &out.ConsecutiveErrorLimit
		*out = new(intstr.IntOrString)
		**out = **in
	}
	in.Provider.DeepCopyInto(&out.Provider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metric.
func (in *Metric) DeepCopy() *Metric {
	if in == nil {
		return nil
	}
	out := new(Metric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricProvider) DeepCopyInto(out *MetricProvider) {
	*out = *in
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(PrometheusMetric)
		**out = **in
	}
	if in.Kayenta != nil {
		in, out := &in.Kayenta, &out.Kayenta
		*out = new(KayentaMetric)
		(*in).DeepCopyInto(*out)
	}
	if in.Web != nil {
		in, out := &in.Web, &out.Web
		*out = new(WebMetric)
		(*in).DeepCopyInto(*out)
	}
	if in.Datadog != nil {
		in, out := &in.Datadog, &out.Datadog
		*out = new(DatadogMetric)
		**out = **in
	}
	if in.Wavefront != nil {
		in, out := &in.Wavefront, &out.Wavefront
		*out = new(WavefrontMetric)
		**out = **in
	}
	if in.NewRelic != nil {
		in, out := &in.NewRelic, &out.NewRelic
		*out = new(NewRelicMetric)
		**out = **in
	}
	if in.Job != nil {
		in, out := &in.Job, &out.Job
		*out = new(JobMetric)
		(*in).DeepCopyInto(*out)
	}
	if in.CloudWatch != nil {
		in, out := &in.CloudWatch, &out.CloudWatch
		*out = new(CloudWatchMetric)
		(*in).DeepCopyInto(*out)
	}
	if in.Graphite != nil {
		in, out := &in.Graphite, &out.Graphite
		*out = new(GraphiteMetric)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricProvider.
func (in *MetricProvider) DeepCopy() *MetricProvider {
	if in == nil {
		return nil
	}
	out := new(MetricProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricResult) DeepCopyInto(out *MetricResult) {
	*out = *in
	if in.Measurements != nil {
		in, out := &in.Measurements, &out.Measurements
		*out = make([]Measurement, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricResult.
func (in *MetricResult) DeepCopy() *MetricResult {
	if in == nil {
		return nil
	}
	out := new(MetricResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricRun) DeepCopyInto(out *MetricRun) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricRun.
func (in *MetricRun) DeepCopy() *MetricRun {
	if in == nil {
		return nil
	}
	out := new(MetricRun)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MetricRun) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricRunList) DeepCopyInto(out *MetricRunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MetricRun, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricRunList.
func (in *MetricRunList) DeepCopy() *MetricRunList {
	if in == nil {
		return nil
	}
	out := new(MetricRunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MetricRunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricRunSpec) DeepCopyInto(out *MetricRunSpec) {
	*out = *in
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]Metric, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricRunSpec.
func (in *MetricRunSpec) DeepCopy() *MetricRunSpec {
	if in == nil {
		return nil
	}
	out := new(MetricRunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricRunStatus) DeepCopyInto(out *MetricRunStatus) {
	*out = *in
	if in.MetricResults != nil {
		in, out := &in.MetricResults, &out.MetricResults
		*out = make([]MetricResult, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.StartedAt != nil {
		in, out := &in.StartedAt, &out.StartedAt
		*out = (*in).DeepCopy()
	}
	out.RunSummary = in.RunSummary
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricRunStatus.
func (in *MetricRunStatus) DeepCopy() *MetricRunStatus {
	if in == nil {
		return nil
	}
	out := new(MetricRunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NewRelicMetric) DeepCopyInto(out *NewRelicMetric) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NewRelicMetric.
func (in *NewRelicMetric) DeepCopy() *NewRelicMetric {
	if in == nil {
		return nil
	}
	out := new(NewRelicMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrometheusMetric) DeepCopyInto(out *PrometheusMetric) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrometheusMetric.
func (in *PrometheusMetric) DeepCopy() *PrometheusMetric {
	if in == nil {
		return nil
	}
	out := new(PrometheusMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunSummary) DeepCopyInto(out *RunSummary) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunSummary.
func (in *RunSummary) DeepCopy() *RunSummary {
	if in == nil {
		return nil
	}
	out := new(RunSummary)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScopeDetail) DeepCopyInto(out *ScopeDetail) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScopeDetail.
func (in *ScopeDetail) DeepCopy() *ScopeDetail {
	if in == nil {
		return nil
	}
	out := new(ScopeDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WavefrontMetric) DeepCopyInto(out *WavefrontMetric) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WavefrontMetric.
func (in *WavefrontMetric) DeepCopy() *WavefrontMetric {
	if in == nil {
		return nil
	}
	out := new(WavefrontMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebMetric) DeepCopyInto(out *WebMetric) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]WebMetricHeader, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebMetric.
func (in *WebMetric) DeepCopy() *WebMetric {
	if in == nil {
		return nil
	}
	out := new(WebMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebMetricHeader) DeepCopyInto(out *WebMetricHeader) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebMetricHeader.
func (in *WebMetricHeader) DeepCopy() *WebMetricHeader {
	if in == nil {
		return nil
	}
	out := new(WebMetricHeader)
	in.DeepCopyInto(out)
	return out
}
