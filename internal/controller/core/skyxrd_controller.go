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

package core

import (
	"bytes"
	"context"
	"text/template"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cpextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// SkyXRDReconciler reconciles a SkyXRD object
type SkyXRDReconciler struct {
	client.Client
	DynamicClient dynamic.Interface
	Scheme        *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyxrds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyxrds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyxrds/finalizers,verbs=update
// +kubebuilder:rbac:groups=xrds.skycluster.savitestbed.ca,resources=xprovidersetups/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *SkyXRDReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("SkyXRD [" + req.Name + "] Reconciler started")

	// Get the resource
	var skyXRD corev1alpha1.SkyXRD
	if err := r.Get(ctx, req.NamespacedName, &skyXRD); err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyXRD [" + req.Name + "] not found, why?")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "Unable to fetch SkyXRD, something is wrong.")
			return ctrl.Result{}, err
		}
	}

	// TODO: uncomment
	if skyXRD.Spec.TaskPlacement.Status != "Optimal" {
		log.Info("SkyXRD [" + req.Name + "] Ignored. Status is: " + skyXRD.Spec.TaskPlacement.Status)
		return ctrl.Result{}, nil
	}

	// Status is Optimal, preparing the composite objects according to deployment plan

	// Find the corresponding SkyApp object
	var skyApp corev1alpha1.SkyApp
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: skyXRD.Namespace,
		Name:      skyXRD.Spec.SkyAppRefName,
	}, &skyApp); err != nil {
		log.Error(err, "Failed to fetch referenced SkyApp")
		return ctrl.Result{}, err
	}

	// To Consider: There may be multiple tasks that are placed in the same provider
	// and they require the same vservice (e.g. SkyK8SCluster). This case, we need to
	// figure out which composed virtual service should be created only once,
	// and which ones should be created multiple times.
	// For example, for SkyK8SCluster, we should create only one XProviderSetup
	// and multiple XSkyCluster composed virtual service if multiple tasks are placed
	// in the same provider with SkyK8SCluster vservice requirements.

	// var deployPlan corev1alpha1.DeploymentPlan
	deployPlan := make(map[corev1alpha1.ProviderRef][]corev1alpha1.VServiceComposition)
	// TODO: skyservices or vservices
	// var vservices []corev1alpha1.VServiceComposition

	// TODO: Consider a case that a task is placed in multiple providers
	for taskName, providers := range skyXRD.Spec.TaskPlacement.Tasks {
		log.Info("SkyXRD [" + req.Name + "] Task: " + taskName)
		// We need to get the vservice for each task
		// This coming from SkyApp.Spec.AppConfig, e.g. frontend

		// ensure pp in providers exists in the deployPlan
		for _, pp := range providers {
			if _, ok := deployPlan[pp]; !ok {
				log.Info("SkyXRD [" + req.Name + "]:  Initilizing " + pp.Name + ", Region: " + pp.Region + ", Type: " + pp.Type)
				deployPlan[pp] = make([]corev1alpha1.VServiceComposition, 0)
			}
		}

		// Find APIVersion and Kind for this taskName
		// and add it to the deployPlan for each provider
		// we ensure if same service is required by multiple providers
		// each provider has its own composition
		vss := make([]corev1alpha1.VServiceComposition, 0)
		for _, task := range skyApp.Spec.AppConfig {
			if task.Name == taskName {
				// Get the virtual services for this task
				// The task may have multiple virtual services
				// But we manually adjust the vservice to skyk8s (this is the only one supported now)
				// this means we only support one type of virtual service for now
				// we can break the loop
				// e.g. "skyk8scluster"
				for _, vserviceConstraint := range task.Constraints.VirtualServiceConstraints {
					// TODO: remove this section
					// manually adjust the vservice to skyk8s (this is the only one supported now)
					vserviceConstraint.VirtualServiceName = "skyk8scluster"
					log.Info("SkyXRD [" + req.Name + "]     Task: " + task.Name + " " + vserviceConstraint.VirtualServiceName)
					// Now we can retrive the information from virtual services with api and kind
					var vs corev1alpha1.VirtualService
					if err := r.Get(ctx, client.ObjectKey{
						Namespace: skyXRD.Namespace,
						Name:      vserviceConstraint.VirtualServiceName,
					}, &vs); err != nil {
						log.Error(err, "Failed to fetch referenced VirtualService")
						return ctrl.Result{}, err
					}
					// e.g. apiVersion: xrds.skycluster.savitestbed.ca/v1alpha1
					// e.g. kind: VirtualService
					vss = append(vss, vs.Spec.VServiceComposition...)
					// TODO: For now we assume only one virtual service is required by each task
					// Remove the break if we need to support multiple virtual services
					break
				}
				// We are looking for a particular task, so we can break here
				break
			}
		}

		// Now add the virtual service composition to the deployPlan for all the providers
		// this task is going to be placed
		for _, pp := range providers {
			deployPlan[pp] = append(deployPlan[pp], vss...)
		}
	}

	// Let's print all the deployPlan
	for pp, vss := range deployPlan {
		log.Info("SkyXRD [" + req.Name + "] Provider: " + pp.Name + ", Region: " + pp.Region + ", Type: " + pp.Type)
		for _, vs := range vss {
			log.Info("SkyXRD [" + req.Name + "]     " + vs.Kind + "." + vs.APIVersion)
		}
	}

	// Now we have the deployPlan, we can create the composite objects
	// create a map of deployed resources per provider
	providerSetup := make(map[corev1alpha1.ProviderRef]bool, 0)

	for pp, vss := range deployPlan {
		// log.Info("SkyXRD [" + req.Name + "] Provider: " + pp.Name + ", Region: " + pp.Region + ", Type: " + pp.Type)
		for _, vs := range vss {
			// We support only one type of virtual service for now: skyk8scluster
			// For each provider, this composite resource consists of XProviderSetup and XSkyCluster
			// The XProviderSetup is created only once for each provider
			// For now, we submit both XProviderSetup and XSkyCluster at the same time and
			// manage dependencies with XRD go templates
			log.Info("SkyXRD [" + req.Name + "]     " + vs.Kind + "." + vs.APIVersion)
			if !providerSetup[pp] {
				// Create XProviderSetup

				// Figure out values per provider out the values
				params := getParamsForProvider(pp, skyXRD.Spec.AppName)

				if err := r.createXProviderSetup(ctx, &skyXRD, params); err != nil {
					log.Error(err, "Failed to create XProviderSetup")
					return ctrl.Result{}, err
				}
				log.Info("SkyXRD [" + req.Name + "] Created XProviderSetup (" + pp.Name + ", " + pp.Region + ", " + pp.Type + ")")
				providerSetup[pp] = true
			}

		}
	}

	//    SkyXRD constains a list of providers (XSkyProvider),
	//		a list of K8S controllers (XSkyCluster with type ctrl) and
	// 		a list of K8S agents (XSkyCluster with type agent).

	// 		The controller creates the XProviderSetup, then
	// 		Once the XProviderSetups are ready, creates XSkyCluster objects for each provider
	//    One is ctrl node and the rest are agent XSkyCluster.

	//    The controller observes the XRD objects and wait for them to become ready.
	//    We now have a Sky K8S cluster.

	// 2. Given we have an overlay K8S now, the actual application should be submitted.
	//    The kubeconfig of the overlay K8S should be fetched.
	//    TODO: How to fetch the kubeconfig and how to use it to create a new clientset?
	//    Within the controller, use the kubeconfig to create a new clientset
	//    submit all deployment, service, etc. objects to the overlay K8S.

	return ctrl.Result{}, nil
}

// Func to return appropriate params for the XRD template given the provider
func getParamsForProvider(pp corev1alpha1.ProviderRef, appName string) XProviderSetupParams {

	ipGroup := "30"
	ipSubnet := "210"
	switch pp.Name {
	case "savi":
		switch pp.Region {
		case "scinet":
			ipGroup = "30"
			ipSubnet = "211"
		case "vaughan":
			ipGroup = "29"
			ipSubnet = "211"
		}
	case "aws":
		ipGroup = "27"
		switch pp.Region {
		case "ca-central-1":
			ipSubnet = "211"
		case "us-east-1":
			ipSubnet = "212"
		}
	case "gcp":
		ipGroup = "28"
		switch pp.Region {
		case "us-west1":
			ipSubnet = "211"
		case "us-east1":
			ipSubnet = "212"
		}
	case "azure":
		ipGroup = "26"
		switch pp.Region {
		case "centralus":
			ipSubnet = "211"
		case "canadaeast":
			ipSubnet = "212"
		case "canadacentral":
			ipSubnet = "213"
		}
	// The default shoud not be reached
	default:
		ipGroup = "30"
		ipSubnet = "212"
	}

	params := XProviderSetupParams{
		Provider: pp.Name,
		Region:   pp.Region,
		Zone:     "default",
		App:      appName,
		IpGroup:  ipGroup,
		IpSubnet: ipSubnet,
	}

	return params
}

func (r *SkyXRDReconciler) createXProviderSetup(ctx context.Context, skyxrd *corev1alpha1.SkyXRD, params XProviderSetupParams) error {
	log := log.FromContext(ctx)

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "xprovidersetups",
	}

	tmpl, err := template.New("xrd").Parse(xProviderSetupParam)
	if err != nil {
		log.Error(err, "Failed to parse template")
		return err
	}

	// Execute the template with the provided parameters
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		log.Error(err, "Failed to execute template")
		return err
	}

	// Decode YAML to unstructured object
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = dec.Decode(buf.Bytes(), nil, obj)
	if err != nil {
		log.Error(err, "Failed to decode YAML")
		return err
	}

	// set the ownership
	obj.SetOwnerReferences([]metav1.OwnerReference{
		metav1.OwnerReference{
			APIVersion: skyxrd.APIVersion,
			Kind:       skyxrd.Kind,
			Name:       skyxrd.Name,
			UID:        skyxrd.UID,
			Controller: func(b bool) *bool { return &b }(true),
		},
	})

	resourceClient := r.DynamicClient.Resource(gvr)
	// Create the object in the Kubernetes cluster
	// err = r.Create(context.Background(), obj)
	if _, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create XProviderSetup, maybe object already exists?")
			return err
		}
	}
	// log.Info("Created XProviderSetup: " + obj.GetName())
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkyXRDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.SkyXRD{}).
		Watches(&cpextv1.Composition{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
