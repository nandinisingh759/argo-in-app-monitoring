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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InAppMetricSpec defines the desired state of InAppMetric
type InAppMetricSpec struct {
	Command           string            `json:"command,omitempty"`
	Schedule          string            `json:"schedule,omitempty"`
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`
	//Suspend                 *bool             `json:"suspend,omitempty"`
	StartingDeadlineSeconds *int64   `json:"startingDeadlineSeconds,omitempty"`
	Metrics                 []Metric `json:"metrics" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,1,rep,name=metrics"`
}

type ConcurrencyPolicy string

const (
	AllowConcurrent   ConcurrencyPolicy = "Allow"
	ForbidConcurrent  ConcurrencyPolicy = "Forbid"
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// InAppMetricStatus defines the observed state of InAppMetric
type InAppMetricStatus struct {
	Active           []corev1.ObjectReference `json:"active,omitempty"`
	LastScheduleTime *metav1.Time             `json:"lastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InAppMetric is the Schema for the inappmetrics API
type InAppMetric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InAppMetricSpec   `json:"spec,omitempty"`
	Status InAppMetricStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InAppMetricList contains a list of InAppMetric
type InAppMetricList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InAppMetric `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InAppMetric{}, &InAppMetricList{})
}
