/*
Copyright 2025.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SkyClusterSpec defines the desired state of SkyCluster.
type SkyClusterSpec struct {
	// DataflowPolicyRef is the reference to the DataflowPolicy
	// The DataflowPolicy is used to define the dataflow policy and
	// Should has the same name and namespace as the SkyCluster
	DataflowPolicyRef ResourceSpec `json:"dataflowPolicyRef,omitempty"`
	// DeploymentPolicyRef is the reference to the DeploymentPolicy
	// The DeploymentPolicy is used to define the deployment policy and
	// Should has the same name and namespace as the SkyCluster
	DeploymentPolciyRef ResourceSpec `json:"deploymentPolicyRef,omitempty"`
	// SkyProviders is the list of the SkyProvider that will be provisioned
	// This list should be populated when the optimizer is executed
	SkyProviders []ProviderRefSpec `json:"skyProviders,omitempty"`
	// SkyServices is the list of the SkyService that will be used by the optimization
	// and will be scheduled to be deployed in one of the SkyProviders
	SkyServices []corev1.ObjectReference `json:"skyServices,omitempty"`
	// SkyApps is the list of the deployments that will be used by the optimization
	// The deployments should be created prior to the SkyCluster creation
	SkyApps []corev1.ObjectReference `json:"skyApps,omitempty"`
}

// SkyClusterStatus defines the observed state of SkyCluster.
type SkyClusterStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SkyCluster is the Schema for the skyclusters API.
type SkyCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkyClusterSpec   `json:"spec,omitempty"`
	Status SkyClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SkyClusterList contains a list of SkyCluster.
type SkyClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkyCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkyCluster{}, &SkyClusterList{})
}
