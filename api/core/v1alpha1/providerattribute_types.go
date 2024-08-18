/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderMetrics struct {
	DstProviderRef ProviderRef `json:"dstProviderReference"`
	Latency        string      `json:"latency,omitempty"`
	EgressDataCost string      `json:"egressDataCost,omitempty"`
}

// ProviderAttributeSpec defines the desired state of ProviderAttribute
type ProviderAttributeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProviderReference ProviderRef       `json:"providerReference"`
	ProviderMetrics   []ProviderMetrics `json:"providerMetrics"`
}

// ProviderAttributeStatus defines the observed state of ProviderAttribute
type ProviderAttributeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ProviderAttribute is the Schema for the providerattributes API
type ProviderAttribute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderAttributeSpec   `json:"spec,omitempty"`
	Status ProviderAttributeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderAttributeList contains a list of ProviderAttribute
type ProviderAttributeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderAttribute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProviderAttribute{}, &ProviderAttributeList{})
}
