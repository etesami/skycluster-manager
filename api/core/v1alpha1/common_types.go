package v1alpha1

type ProviderRef struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Type   string `json:"type"`
}

type TaskPlacement struct {
	Status string                   `json:"status"`
	Tasks  map[string][]ProviderRef `json:"tasks"`
}

// a map from providers to the list of compositions to be created and maintained
// type DeploymentPlan map[string]VirtualService

type SkyService struct {
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Type       string `json:"type,omitempty"`
}

// TODO: We may need tp update the DeployedServices type later to include
// more specific information about the deployed services.
// Alternatively, we can create a new type for each service type.
// This should be investigated further.
// DeployedServices: ProviderRef -> [SkyServiceName] -> {Type: "ctrl" or "agent" or "etc."}
type DeployedServices struct {
	Provider ProviderRef           `json:"provider"`
	Services map[string]SkyService `json:"vservices,omitempty"`
}

// Task -> Regions
// Regions:
// 	1. Should be prepared: XProviderSetup
//  2. Resources should be created: XSkyCluster (composition)

// What is offered across providers? A K8S cluster to run tasks
// How is the composition done?
//   SkyService: SkyK8SCluster, SkyVM
//      This is what we use in SkyApp object as virtual service, there is no notion of providers here
//      The SkyService is defined using configmaps, with its costs, in addition to its composed resources.
//
//      There is a slight difference here, while SkyK8SCluster spans across providers, a SkyVM is only for one provider
// 		  However, we hide the specification of providers from the user.
//

//   When we know the requirements for a SkyService (from the configmaps),
//   we can create all the individual resources needed to create the SkyService.
//   SkyService -> [XComposition (with provider name, region, zone)]

//   e.g.
//   SkyService: SkyK8SCluster (across set of providers)
// 	   Task1 (with VS: SkyK8S): [Provider1, maybe more e.g. Provider2, ...]
// 	   Task2 (with VS: SkyK8S): [Provider2]
// 	   Task3 (with VS: SkyK8S): [Provider3]
//   Based on the definition of the SkyService, for each provider we should have:
//      XCompositions: XProviderSetup, XSkyCluster (ctrl or agent)
//         		         These two are high level Crossplane resources
