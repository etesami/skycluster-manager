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

type LocationConstraints struct {
	ProviderName string `json:"providerName,omitempty"`
	Region       string `json:"region,omitempty"`
	ProviderType string `json:"providerType,omitempty"`
}

type VirtualServiceConstraints struct {
	VirtualServiceName string `json:"virtualServiceName"`
}

type AppConstranints struct {
	LocationConstraints       []LocationConstraints       `json:"locationConstraints,omitempty"`
	VirtualServiceConstraints []VirtualServiceConstraints `json:"virtualServiceConstraints,omitempty"`
}

type AppConfigDetail struct {
	Name        string          `json:"name"`
	Constraints AppConstranints `json:"constraints"`
}

// SkyAppSpec defines the desired state of SkyApp
type SkyAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AppName   string            `json:"appName"`
	Namespace string            `json:"namespace"`
	AppConfig []AppConfigDetail `json:"appConfig"`
}

// SkyAppStatus defines the observed state of SkyApp
type SkyAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SkyApp is the Schema for the skyapps API
type SkyApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkyAppSpec   `json:"spec,omitempty"`
	Status SkyAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SkyAppList contains a list of SkyApp
type SkyAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SkyApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SkyApp{}, &SkyAppList{})
}
