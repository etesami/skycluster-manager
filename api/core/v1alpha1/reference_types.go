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

type DeployLocation struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Zone     string `json:"zone"`
	Location string `json:"location"`
}
