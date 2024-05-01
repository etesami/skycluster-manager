package v1alpha1

// DeploymentRef is a reference to a Deployment resource
type DeploymentRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// DataflowRef is a reference to a Dataflow resource
type DataflowRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
