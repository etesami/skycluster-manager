package v1alpha1

type ProviderRef struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Type   string `json:"type"`
}
