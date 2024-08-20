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
	// if skyXRD.Spec.DeploymentPlan.Status != "Optimal" {
	// 	log.Info("SkyXRD [" + req.Name + "] Ignored. Status is: " + skyXRD.Spec.DeploymentPlan.Status)
	// 	return ctrl.Result{}, nil
	// }
	// // Status is Optimal
	// for task, plan := range skyXRD.Spec.DeploymentPlan.Tasks {
	// 	for _, provider := range plan {
	// 		log.Info("SkyXRD [" + req.Name + "] Task: " + task + " (" + provider.Name + ", " + provider.Region + ", " + provider.Type + ")")
	// 	}
	// }

	// _ = &cpext.CompositeResourceDefinition{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      skyXRD.Spec.AppName,
	// 		Namespace: skyXRD.Namespace,
	// 	},
	// }

	// // items, err := cpext.Compositions()
	// compositionList := &cpext.CompositionList{}
	// if err := r.List(ctx, compositionList); err != nil {
	// 	log.Error(err, "Failed to list compositions")
	// }
	// for _, comp := range compositionList.Items {
	// 	log.Info("SkyXRD [" + req.Name + "] CompositeTypeRef: " + comp.Spec.CompositeTypeRef.Kind + comp.Spec.CompositeTypeRef.APIVersion)
	// }

	var gvr = schema.GroupVersionResource{
		Group:    "xrds.skycluster.savitestbed.ca",
		Version:  "v1alpha1",
		Resource: "xprovidersetups",
	}

	type XRDParams struct {
		Provider string
		Region   string
		Zone     string
		App      string
		IpGroup  string
		IpSubnet string
	}
	params := XRDParams{
		Provider: "savi",
		Region:   "scinet",
		Zone:     "default",
		App:      "app1",
		IpGroup:  "30",
		IpSubnet: "212",
	}

	tmpl, err := template.New("xrd").Parse(xrdTemplate)
	if err != nil {
		log.Error(err, "Failed to parse template")
		return ctrl.Result{}, err
	}

	// Execute the template with the provided parameters
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		log.Error(err, "Failed to execute template")
		return ctrl.Result{}, err
	}

	// Decode YAML to unstructured object
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = dec.Decode(buf.Bytes(), nil, obj)
	if err != nil {
		log.Error(err, "Failed to decode YAML")
		return ctrl.Result{}, err
	}

	resourceClient := r.DynamicClient.Resource(gvr)

	// Create the object in the Kubernetes cluster
	// err = r.Create(context.Background(), obj)
	if _, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Error(err, "Failed to create XProviderSetup, maybe object already exists?")
			return ctrl.Result{}, err
		}
	}

	// log.Info("Created XProviderSetup: " + result.GetName())

	// Find the corresponding SkyApp object
	var skyApp corev1alpha1.SkyApp
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: skyXRD.Namespace,
		Name:      skyXRD.Spec.SkyAppRefName,
	}, &skyApp); err != nil {
		log.Error(err, "Failed to fetch referenced SkyApp")
		return ctrl.Result{}, err
	}

	// TO Consider: There may be multiple tasks that are placed in the same provider
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
		// This coming from SkyApp.Spec.AppConfig
		// e.g. frontend

		// ensure pp in providers exists in the deployPlan
		for _, pp := range providers {
			if _, ok := deployPlan[pp]; !ok {
				log.Info("SkyXRD [" + req.Name + "]:    Initilizing " + pp.Name + ", Region: " + pp.Region + ", Type: " + pp.Type)
				deployPlan[pp] = make([]corev1alpha1.VServiceComposition, 0)
			}
		}

		// find APIVersion and Kind for this taskName
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
					log.Info("SkyXRD [" + req.Name + "]        Task: " + task.Name + " " + vserviceConstraint.VirtualServiceName)
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

	log.Info("SkyXRD [" + req.Name + "] Deployment Plan:")
	// Let's print all the deployPlan
	for pp, vss := range deployPlan {
		log.Info("SkyXRD [" + req.Name + "] Provider: " + pp.Name + ", Region: " + pp.Region + ", Type: " + pp.Type)
		for _, vs := range vss {
			log.Info("SkyXRD [" + req.Name + "]     " + vs.APIVersion + ", " + vs.Kind)
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

// SetupWithManager sets up the controller with the Manager.
func (r *SkyXRDReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.SkyXRD{}).
		Watches(&cpextv1.Composition{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
