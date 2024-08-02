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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type DataflowDestConstraints struct {
	Latency string `json:"latency"`
}

type DataflowDestinations struct {
	Name        string                    `json:"name"`
	Constraints []DataflowDestConstraints `json:"constraints,omitempty"`
}

type DataflowConnections struct {
	Source       string                 `json:"source"`
	Destinations []DataflowDestinations `json:"destinations"`
}

// DataflowAttributeSpec defines the desired state of DataflowAttribute
type DataflowAttributeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AppName     string                `json:"appName"`
	Connections []DataflowConnections `json:"connections"`
}

// DataflowAttributeStatus defines the observed state of DataflowAttribute
type DataflowAttributeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DataflowAttribute is the Schema for the dataflowattributes API
type DataflowAttribute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataflowAttributeSpec   `json:"spec,omitempty"`
	Status DataflowAttributeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataflowAttributeList contains a list of DataflowAttribute
type DataflowAttributeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataflowAttribute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataflowAttribute{}, &DataflowAttributeList{})
}
