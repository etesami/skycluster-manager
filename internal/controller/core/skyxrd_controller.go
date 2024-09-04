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
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	// cpextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"

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
// +kubebuilder:rbac:groups=xrds.skycluster.savitestbed.ca,resources=skyk8scluster/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *SkyXRDReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("SkyXRD [" + req.Name + "] Reconciler started")

	// We are watching SkyXRDs objects and also SkyK8SCluster objects
	// (possibly we need to watch SkyProviderSetup objects as well)
	// If (SkyK8SCluster) object:
	// 		1. Check if the object is ready
	// 		2. If ready, submit the application to the overlay K8S

	// If (SkyXRD) object:
	// 		1. Check if the object has the status "Optimal"
	// 		2. If the status is "Optimal", prepare the composite objects according to the deployment plan

	gvk := schema.GroupVersionKind{
		Group:   "xrds.skycluster.savitestbed.ca",
		Version: "v1alpha1",
		Kind:    "SkyK8SCluster",
	}

	// Create an unstructured object
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(gvk)

	// Fetch the object using the client
	if err := r.Get(ctx, req.NamespacedName, unstructuredObj); err != nil {
		if !errors.IsNotFound(err) {
			// Handle the error if it's not a NotFound error
			log.Error(err, "Failed to get Unstructured object SkyK8SCluster")
			return ctrl.Result{}, err
		}
		// If the error is not found we need to proceed to the next object (i.e. SkyXRD)
	} else {
		// Now, you have the unstructured object
		kubeCfg, found, err := unstructured.NestedString(unstructuredObj.Object, "status", "k3s", "kubeconfig")
		if err != nil {
			log.Error(err, "Failed to get kubeconfig from SkyK8SCluster object")
			return ctrl.Result{}, err
		}
		if found {
			// We have the kubeconfig and we can submit the application to the overlay K8S
			log.Info("SkyXRD  Successfully retrieved the kubeconfig!")
			appName, found, _ := unstructured.NestedString(unstructuredObj.Object, "metadata", "labels", "skycluster/app-name")
			if !found {
				// TODO: make sure the appName is supplied
				appName = "default"
			}
			if err := r.submitAppToRemoteCluster(ctx, kubeCfg, req.Namespace, appName); err != nil {
				log.Error(err, "Failed to submit the application to the remote cluster")
				return ctrl.Result{}, err
			}
			// TODO: We should have a mechanism to follow up the status of the application
		}
	}

	// This is the SkyXRD object reconciliation
	// Get skyXRD the resource
	var skyXRD corev1alpha1.SkyXRD
	if err := r.Get(ctx, req.NamespacedName, &skyXRD); err != nil {
		if errors.IsNotFound(err) {
			// log.Info("SkyXRD [" + req.Name + "] not found, why?")
			// In this case SkyXRD may be deleted.
			// We need to ensure all composite resources created by this object are deleted as well.
			// To do this, we need to delete finalizers from all the resources that
			// are created by this object.
			// We can do this by listing all the compositions and check if they have
			// the OwenerReference to this object and delete their finalizers.
			// List all compositions and check the owner reference
			for _, gvr := range []schema.GroupVersionResource{
				{
					Group:    "xrds.skycluster.savitestbed.ca",
					Version:  "v1alpha1",
					Resource: "skyprovidersetups",
				},
				{
					Group:    "xrds.skycluster.savitestbed.ca",
					Version:  "v1alpha1",
					Resource: "skyk8sclusters",
				},
			} {
				// log.Info("SkyXRD [" + req.Name + "]   Checking resources for " + gvr.Resource)
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
							// log.Info("SkyXRD [" + req.Name + "]    Removing finalizers obj " + obj.GetName())
							// Remove the finalizers
							// TODO: suprisingly, the following code does not work
							// and does not throw any error. The finalizers are not removed.
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

	// SkyXRD object is found, we need to check the status
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
	// For example, for SkyK8SCluster, we should create only one SkyProviderSetup
	// and multiple SkyK8SCluster composed virtual service if multiple tasks are placed
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
	// At this stage we assume there is no deployed vservices, clean start.

	// SkyXRD constains a list of providers (XSkyProvider),
	// a list of K8S controllers (SkyK8SCluster with type ctrl) and
	// a list of K8S agents (SkyK8SCluster with type agent).

	// The controller creates the SkyProviderSetup, then
	// Once the SkyProviderSetups are ready, creates SkyK8SCluster objects for each provider
	// One is ctrl node and the rest are agent SkyK8SCluster.

	deployedServices := make([]corev1alpha1.DeployedServices, 0)
	preparedProviders := make(map[corev1alpha1.ProviderRef]corev1alpha1.SkyService)

	for pp, vss := range deployPlan {
		for _, vs := range vss {
			// We support only one type of virtual service for now: skyk8scluster
			if strings.ToLower(vs.Kind) != "skyk8scluster" &&
				strings.ToLower(vs.Kind) != "skyk8sclusters" {
				log.Info("SkyXRD [" + req.Name + "]   Unsupported virtual service: " + vs.Kind)
				continue
			}

			// For each provider, this composite resource consists of SkyProviderSetup and SkyK8SCluster
			// The SkyProviderSetup is created only once for each provider
			if _, ok := preparedProviders[pp]; !ok {
				// Create SkyProviderSetup
				params := getParamsForProvider(pp, skyXRD.Spec.AppName)

				if err := r.createSkyProviderSetup(ctx, &skyXRD, params); err != nil {
					log.Error(err, "Failed to create SkyProviderSetup")
					return ctrl.Result{}, err
				}
				preparedProviders[pp] = corev1alpha1.SkyService{
					Name:       params.Name,
					APIVersion: params.APIVersion,
				}
				log.Info("SkyXRD [" + req.Name + "]    Created SkyProviderSetup (" + pp.Name + ", " + pp.Region + ", " + pp.Type + ")")
			}

			// Create SkyK8SCluster
			// We need to ensure there is only one ctrl object across all the providers
			var params skyK8SClusterSetupParams
			if len(deployedServices) == 0 {
				// This is the first object, it should be ctrl
				params = getParamsForSkyCluster(pp, skyXRD.Spec.AppName, true, strconv.Itoa(len(deployedServices)+1))
			} else {
				// We already should have a ctrl object
				// create agents
				params = getParamsForSkyCluster(pp, skyXRD.Spec.AppName, false, strconv.Itoa(len(deployedServices)+1))
			}
			if err := r.createSkyK8SClusterSetup(ctx, &skyXRD, params); err != nil {
				log.Error(err, "Failed to create SkyProviderSetup")
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
			log.Info("SkyXRD [" + req.Name + "]    Created SkyK8SCluster (" + params.Name + ", " + pp.Type + ", " + pp.Name + ")")
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
	// Update the status of the SkyXRD object
	skyXRD.Status.DeployedProviders = deployedProviders
	skyXRD.Status.DeployedServices = deployedServices
	if err := r.Status().Update(ctx, &skyXRD); err != nil {
		log.Error(err, "Failed to update SkyXRD status")
		return ctrl.Result{}, err
	}

	//    The controller observes the claim objects (skyProviderSetup and skyK8SCluster)
	//    and wait for them to become ready.
	//    We now have a Sky K8S cluster.

	// 2. Given we have an overlay K8S now, the actual application should be submitted.
	//    The kubeconfig of the overlay K8S should be fetched.
	//    Within the controller, we use the kubeconfig to create a new clientset
	//    then submit all deployment, service, etc. objects to the overlay K8S.

	return ctrl.Result{}, nil
}

// Func to return appropriate params for the XRD template given the provider
// TODO: Potentially we should use config maps to retrive this data dynamically
// rather than hardcoding it in the code.
func getParamsForProvider(pp corev1alpha1.ProviderRef, appName string) skyProviderSetupParams {

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

	params := skyProviderSetupParams{
		Name:       "skyprovidersetup1-" + pp.Name + "-" + pp.Region + "-default-" + appName,
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

func getParamsForSkyCluster(pp corev1alpha1.ProviderRef, appName string, ctrl bool, num string) skyK8SClusterSetupParams {

	params := skyK8SClusterSetupParams{
		Name: "skyk8scluster1-" + pp.Name + "-" + pp.Region + "-" + func(b bool) string {
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

func (r *SkyXRDReconciler) createSkyK8SClusterSetup(ctx context.Context, skyxrd *corev1alpha1.SkyXRD, params skyK8SClusterSetupParams) error {
	log := log.FromContext(ctx)

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "skyk8sclusters",
	}

	tmpl, err := template.New("xrd").Parse(skyK8SClusterSetupTemplate)
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
		{
			APIVersion: skyxrd.APIVersion,
			Kind:       skyxrd.Kind,
			Name:       skyxrd.Name,
			UID:        skyxrd.UID,
			Controller: func(b bool) *bool { return &b }(true),
		},
	})

	resourceClient := r.DynamicClient.Resource(gvr).Namespace(skyxrd.Namespace)
	// Create the object in the Kubernetes cluster
	// err = r.Create(context.Background(), obj)
	if _, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create SkyK8SClusterSetup, maybe object already exists?")
			return err
		}
	}
	// log.Info("Created SkyK8SClusterSetup: " + obj.GetName())
	return nil
}

func (r *SkyXRDReconciler) createSkyProviderSetup(ctx context.Context, skyxrd *corev1alpha1.SkyXRD, params skyProviderSetupParams) error {
	log := log.FromContext(ctx)

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "skyprovidersetups",
	}

	tmpl, err := template.New("xrd").Parse(skyProviderSetupTemplate)
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
		{
			APIVersion: skyxrd.APIVersion,
			Kind:       skyxrd.Kind,
			Name:       skyxrd.Name,
			UID:        skyxrd.UID,
			Controller: func(b bool) *bool { return &b }(true),
		},
	})

	resourceClient := r.DynamicClient.Resource(gvr).Namespace(skyxrd.Namespace)
	// Create the object in the Kubernetes cluster
	// err = r.Create(context.Background(), obj)
	if _, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create SkyProviderSetup")
			return err
		}
	}
	// log.Info("Created SkyProviderSetup: " + obj.GetName())
	return nil
}

func (r *SkyXRDReconciler) submitAppToRemoteCluster(ctx context.Context, kubeconfig string, namespace string, appName string) error {
	log := log.FromContext(ctx)
	// log.Info("SkyXRD  Submitting the application to the remote cluster")
	// List all deployments with the label "sky" in the current cluster
	deployments := &appsv1.DeploymentList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"skycluster-manager.savitestbed.ca/managed-by": "skycluster",
			"skycluster-manager.savitestbed.ca/app-name":   appName,
		}),
	}
	if err := r.Client.List(ctx, deployments, listOpts...); err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyXRD  No deployments found with given labels")
			return nil
		} else {
			log.Error(err, "Failed to list deployments with given labels")
			return err
		}
	}

	// Services
	services := &corev1.ServiceList{}
	listOpts = []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"skycluster-manager.savitestbed.ca/managed-by": "skycluster",
			"skycluster-manager.savitestbed.ca/app-name":   appName,
		}),
	}
	if err := r.Client.List(ctx, services, listOpts...); err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyXRD No services found with given labels")
			return nil
		} else {
			log.Error(err, "Failed to list services with given labels")
			return err
		}
	}

	configMaps := &corev1.ConfigMapList{}
	listOpts = []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"skycluster-manager.savitestbed.ca/managed-by": "skycluster",
			"skycluster-manager.savitestbed.ca/app-name":   appName,
		}),
	}
	if err := r.Client.List(ctx, configMaps, listOpts...); err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyXRD No configmaps found with given labels")
			return nil
		} else {
			log.Error(err, "Failed to list configmaps with given labels")
			return err
		}
	}

	// Build the config from the kubeconfig string
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return err
	}

	// Create a clientset for the remote cluster
	remoteClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Submit each deployment to the remote cluster
	for _, dep := range deployments.Items {
		lastAppliedConfig, exists := dep.Annotations["kubectl.kubernetes.io/last-applied-configuration"]
		if !exists {
			log.Error(nil, "Annotation 'kubectl.kubernetes.io/last-applied-configuration' not found on deployment")
			return err
		}
		var updatedDep appsv1.Deployment
		if err := json.Unmarshal([]byte(lastAppliedConfig), &updatedDep); err != nil {
			log.Error(err, "Failed to unmarshal 'last-applied-configuration' annotation Deployment "+dep.Name)
			return err
		}
		if _, err := remoteClientset.AppsV1().Deployments(dep.Namespace).Create(ctx, &updatedDep, metav1.CreateOptions{}); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "Failed to create deployment in remote cluster")
				return err
			} else {
				log.Info("[SkyXRD] Deployment " + dep.Name + " already exists in remote cluster!")
			}
		}
		log.Info("[SkyXRD] Deployment " + dep.Name + " created in remote cluster!")
	}

	for _, svc := range services.Items {
		lastAppliedConfig, exists := svc.Annotations["kubectl.kubernetes.io/last-applied-configuration"]
		if !exists {
			log.Error(nil, "Annotation 'kubectl.kubernetes.io/last-applied-configuration' not found on svc"+svc.Name)
			return err
		}
		var updatedSvc corev1.Service
		if err := json.Unmarshal([]byte(lastAppliedConfig), &updatedSvc); err != nil {
			log.Error(err, "Failed to unmarshal 'last-applied-configuration' annotation SVC")
			return err
		}
		if _, err := remoteClientset.CoreV1().Services(svc.Namespace).Create(ctx, &updatedSvc, metav1.CreateOptions{}); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "Failed to create service in remote cluster")
				return err
			} else {
				log.Info("[SkyXRD] Service " + svc.Name + " already exists in remote cluster!")
			}
		}
		log.Info("[SkyXRD] Service " + svc.Name + " created in remote cluster!")
	}
	for _, cm := range configMaps.Items {
		lastAppliedConfig, exists := cm.Annotations["kubectl.kubernetes.io/last-applied-configuration"]
		if !exists {
			log.Error(nil, "Annotation 'kubectl.kubernetes.io/last-applied-configuration' not found on configmap "+cm.Name)
			return err
		}
		var updatedCM corev1.ConfigMap
		if err := json.Unmarshal([]byte(lastAppliedConfig), &updatedCM); err != nil {
			log.Error(err, "Failed to unmarshal 'last-applied-configuration' annotation configmap "+cm.Name)
			return err
		}
		if _, err := remoteClientset.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, &updatedCM, metav1.CreateOptions{}); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "Failed to create configmap in remote cluster")
				return err
			} else {
				log.Info("[SkyXRD] ConfigMap " + cm.Name + " already exists in remote cluster")
			}
		}
		log.Info("[SkyXRD] ConfigMap " + cm.Name + " created in remote cluster!")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkyXRDReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// u := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   "xrds.skycluster.savitestbed.ca",
		Version: "v1alpha1",
		Kind:    "SkyK8SCluster",
	}
	unstructuredObj := &unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(gvk)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.SkyXRD{}).
		// Watches(&corev1alpha1.ILPTask{}, &handler.EnqueueRequestForObject{}).
		Watches(
			unstructuredObj,
			&handler.EnqueueRequestForObject{}, builder.WithPredicates(
				predicate.ResourceVersionChangedPredicate{},
			)).
		Watches(
			unstructuredObj,
			&handler.EnqueueRequestForObject{}, builder.WithPredicates(
				predicate.Funcs{
					UpdateFunc: func(e event.UpdateEvent) bool {
						oldObj, ok1 := e.ObjectOld.(*unstructured.Unstructured)
						newObj, ok2 := e.ObjectNew.(*unstructured.Unstructured)
						if !ok1 || !ok2 {
							return false
						}
						oldStatus, _, _ := unstructured.NestedMap(oldObj.Object, "status")
						newStatus, _, _ := unstructured.NestedMap(newObj.Object, "status")
						return !reflect.DeepEqual(oldStatus, newStatus)
					},
				},
			)).
		Complete(r)
}
