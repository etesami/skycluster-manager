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

type DataflowAttributeRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type SkyAppRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ILPTaskSpec defines the desired state of ILPTask
type ILPTaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ILPTask. Edit ilptask_types.go to remove/update
	AppName              string               `json:"appName"`
	ProblemDefinition    string               `json:"problemDefinition,omitempty"`
	DataflowAttributeRef DataflowAttributeRef `json:"dataflowAttributeRef,omitempty"`
	SkyAppRef            SkyAppRef            `json:"skyAppRef,omitempty"`
}

// ILPTaskStatus defines the observed state of ILPTask
type ILPTaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Result   string `json:"result"`
	Solution string `json:"solution"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ILPTask is the Schema for the ilptasks API
type ILPTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ILPTaskSpec   `json:"spec,omitempty"`
	Status ILPTaskStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ILPTaskList contains a list of ILPTask
type ILPTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ILPTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ILPTask{}, &ILPTaskList{})
}
