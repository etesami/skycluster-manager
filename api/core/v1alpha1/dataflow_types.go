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

type DataFlowConnection struct {
	// ConnectionSource is the source deployment sends data
	ConnectionSource DeploymentRef `json:"connectionSource"`

	// ConnectionDestination is the destination deployment receives data
	ConnectionDestination DeploymentRef `json:"connectionDestination"`
}

// DataflowSpec defines the desired state of Dataflow
type DataflowSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AppName is the name of the application, which should be the same as the "app" lable in Deployment
	AppName string `json:"appName"`

	// DataFlowConnections is the list of connections between deployments
	DataFlowConnections []DataFlowConnection `json:"dataFlowConnections"`
}

// DataflowStatus defines the observed state of Dataflow
type DataflowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dataflow is the Schema for the dataflows API
type Dataflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataflowSpec   `json:"spec,omitempty"`
	Status DataflowStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DataflowList contains a list of Dataflow
type DataflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dataflow{}, &DataflowList{})
}
