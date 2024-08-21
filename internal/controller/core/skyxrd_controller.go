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
	"strconv"
	"strings"
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
			// In this case SkyXRD may be deleted.
			// We need to ensure all composition created by this object are deleted as well.
			// To do this, we need to delete finalizers from all the compositions that
			// are created by this object.
			// We can do this by listing all the compositions and check if they have
			// the OwenerReference to this object and delete their finalizers.
			// List all compositions and check the owner reference
			for _, gvr := range []schema.GroupVersionResource{
				{
					Group:    "xrds.skycluster.savitestbed.ca",
					Version:  "v1alpha1",
					Resource: "xprovidersetups",
				},
				{
					Group:    "xrds.skycluster.savitestbed.ca",
					Version:  "v1alpha1",
					Resource: "xskyclusters",
				},
			} {
				log.Info("SkyXRD [" + req.Name + "]   Checking resources for " + gvr.Resource)
				resourceClient := r.DynamicClient.Resource(gvr)
				list, err := resourceClient.List(ctx, metav1.ListOptions{})
				if err != nil {
					log.Error(err, "Failed to list resources")
					return ctrl.Result{}, err
				}
				for _, obj := range list.Items {
					// Check if the object has the owner reference to the SkyXRD object
					ownerRefs := obj.GetOwnerReferences()
					for _, ownerRef := range ownerRefs {
						if ownerRef.Name == req.Name {
							log.Info("SkyXRD [" + req.Name + "]    Removing finalizers obj " + obj.GetName())
							// Remove the finalizers
							obj.SetFinalizers([]string{})
							if _, err := resourceClient.Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
								log.Error(err, "Failed to remove finalizers")
								return ctrl.Result{}, err
							}
						}
					}
				}
			}
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

	// TODO: We need to deploy update procedure. Currently, we skip creating a
	// deployment plan if services are deployed already.

	if len(skyXRD.Status.DeployedServices) > 0 {
		// TODO: how to update deployed services?
		log.Info("SkyXRD [" + req.Name + "] Services are already deployed, skipping the deployment plan")
		return ctrl.Result{}, nil
	}

	deployPlan := make(map[corev1alpha1.ProviderRef][]corev1alpha1.VServiceComposition)

	// if the optimization status is not optimal we should not proceed
	if skyXRD.Spec.TaskPlacement.Status != "Optimal" {
		log.Info("SkyXRD [" + req.Name + "] Ignored. Optimization result is: " + skyXRD.Spec.TaskPlacement.Status)
		return ctrl.Result{}, nil
	}

	// TODO: Consider a case that a task is placed in multiple providers
	for taskName, providers := range skyXRD.Spec.TaskPlacement.Tasks {
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
					log.Info("SkyXRD [" + req.Name + "]     Requested VService: " + vserviceConstraint.VirtualServiceName)
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
	// At this stage we assume there is no deployed vservices.

	deployedServices := make([]corev1alpha1.DeployedServices, 0)
	preparedProviders := make(map[corev1alpha1.ProviderRef]corev1alpha1.SkyService)

	for pp, vss := range deployPlan {
		for _, vs := range vss {
			// We support only one type of virtual service for now: skyk8scluster
			log.Info("SkyXRD [" + req.Name + "] Deploying on " + vs.Kind + "." + vs.APIVersion)
			if strings.ToLower(vs.Kind) != "skyk8scluster" &&
				strings.ToLower(vs.Kind) != "skyk8sclusters" {
				log.Info("SkyXRD [" + req.Name + "]   Unsupported virtual service: " + vs.Kind)
				continue
			}

			// For each provider, this composite resource consists of XProviderSetup and XSkyCluster
			// The XProviderSetup is created only once for each provider
			if _, ok := preparedProviders[pp]; !ok {
				// Create XProviderSetup
				params := getParamsForProvider(pp, skyXRD.Spec.AppName)

				if err := r.createXProviderSetup(ctx, &skyXRD, params); err != nil {
					log.Error(err, "Failed to create XProviderSetup")
					return ctrl.Result{}, err
				}
				preparedProviders[pp] = corev1alpha1.SkyService{
					Name:       params.Name,
					APIVersion: params.APIVersion,
				}
				log.Info("SkyXRD [" + req.Name + "]    Created XProviderSetup (" + pp.Name + ", " + pp.Region + ", " + pp.Type + ")")
			}

			// Create XSkyCluster
			// We need to ensure there is only one ctrl object across all the providers
			var params XSkyClusterSetupParams
			if len(deployedServices) == 0 {
				// This is the first object, it should be ctrl
				params = getParamsForSkyCluster(pp, skyXRD.Spec.AppName, true, strconv.Itoa(len(deployedServices)+1))
			} else {
				// We already should have a ctrl object
				// create agents
				params = getParamsForSkyCluster(pp, skyXRD.Spec.AppName, false, strconv.Itoa(len(deployedServices)+1))
			}
			if err := r.createXSkyClusterSetup(ctx, &skyXRD, params); err != nil {
				log.Error(err, "Failed to create XProviderSetup")
				return ctrl.Result{}, err
			}
			deployedServices = append(deployedServices, corev1alpha1.DeployedServices{
				Provider: pp,
				Services: map[string]corev1alpha1.SkyService{
					params.Name: {
						Name:       params.Name,
						APIVersion: params.APIVersion,
						Type:       params.Type,
					},
				},
			})
			log.Info("SkyXRD [" + req.Name + "]    Created XSkyCluster (" + params.Name + ", " + pp.Type + ", " + pp.Name + ")")
		}
	}

	deployedProviders := make([]corev1alpha1.DeployedServices, 0)
	// Update the status of the SkyXRD object
	for pp, ds := range preparedProviders {
		deployedProviders = append(deployedProviders, corev1alpha1.DeployedServices{
			Provider: pp,
			Services: map[string]corev1alpha1.SkyService{
				ds.Name: {
					Name:       ds.Name,
					APIVersion: ds.APIVersion,
				},
			},
		})
	}
	skyXRD.Status.DeployedProviders = deployedProviders

	// Update the status of the SkyXRD object
	skyXRD.Status.DeployedServices = deployedServices
	if err := r.Status().Update(ctx, &skyXRD); err != nil {
		log.Error(err, "Failed to update SkyXRD status")
		return ctrl.Result{}, err
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
		Name:       "xprovidersetup1-" + pp.Name + "-" + pp.Region + "-default-" + appName,
		APIVersion: "xrds.skycluster.savitestbed.ca",
		Provider:   pp.Name,
		Region:     pp.Region,
		Zone:       "default",
		AppName:    appName,
		IpGroup:    ipGroup,
		IpSubnet:   ipSubnet,
	}

	return params
}

func getParamsForSkyCluster(pp corev1alpha1.ProviderRef, appName string, ctrl bool, num string) XSkyClusterSetupParams {

	params := XSkyClusterSetupParams{
		Name: "xskycluster1-" + pp.Name + "-" + pp.Region + "-" + func(b bool) string {
			if b {
				return "ctrl"
			} else {
				return "agent"
			}
		}(ctrl) + "-" + appName + "-" + num,
		APIVersion: "xrds.skycluster.savitestbed.ca",
		Provider:   pp.Name,
		Region:     pp.Region,
		AppName:    appName,
		Type: func(b bool) string {
			if b {
				return "ctrl"
			} else {
				return "agent"
			}
		}(ctrl),
		Num: num,
		Size: func(b bool) string {
			if b {
				return "xlarge"
			} else {
				return "small"
			}
		}(ctrl),
		IsController: func(b bool) string {
			if b {
				return "true"
			} else {
				return "false"
			}
		}(ctrl),
	}

	return params
}

func (r *SkyXRDReconciler) createXSkyClusterSetup(ctx context.Context, skyxrd *corev1alpha1.SkyXRD, params XSkyClusterSetupParams) error {
	log := log.FromContext(ctx)

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "xskyclusters",
	}

	tmpl, err := template.New("xrd").Parse(xSkyClusterSetupTemplate)
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
			log.Error(err, "Failed to create XSkyClusterSetup, maybe object already exists?")
			return err
		}
	}
	// log.Info("Created XSkyClusterSetup: " + obj.GetName())
	return nil
}

func (r *SkyXRDReconciler) createXProviderSetup(ctx context.Context, skyxrd *corev1alpha1.SkyXRD, params XProviderSetupParams) error {
	log := log.FromContext(ctx)

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "xprovidersetups",
	}

	tmpl, err := template.New("xrd").Parse(xProviderSetupTemplate)
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
