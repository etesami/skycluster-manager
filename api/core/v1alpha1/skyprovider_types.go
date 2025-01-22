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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SkyProviderSpec defines the desired state of SkyProvider.
type SkyProviderSpec struct {
	SecGroup SecGroup `json:"secGroup,omitempty"`
	// IpCidrRange       string          `json:"ipCidrRange,omitempty"`
	KeypairRef        KeypairRefSpec  `json:"keypairRef,omitempty"`
	ProviderRef       ProviderRefSpec `json:"providerRef,omitempty"`
	SkyDependencyList []SkyDependency `json:"skyDependencies,omitempty"`
}

// SkyProviderStatus defines the observed state of SkyProvider.
type SkyProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SkyProvider is the Schema for the skyproviders API.
type SkyProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkyProviderSpec   `json:"spec,omitempty"`
	Status SkyProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SkyProviderList contains a list of SkyProvider.
type SkyProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkyProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkyProvider{}, &SkyProviderList{})
}

// Implement the interface for SkyProvider
func (obj *SkyProvider) GetProviderRef() ProviderRefSpec {
	return obj.Spec.ProviderRef
}
