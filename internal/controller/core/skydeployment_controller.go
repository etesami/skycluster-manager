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
	"context"
	"fmt"

	pegraph "github.com/etesami/pegraph"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// SkyDeploymentReconciler reconciles a SkyDeployment object
type SkyDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=skydeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=skydeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=skydeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;update;patch;delete
//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=skyapps,verbs=create;update;patch;delete

// Reconcile function workflow
//  1. Fetch the Deployment if it has the "manage-by" annotation set to skycluster if so:
//     2.1. Check if the deployment has annotation skyappname set
//     -- Create or update the SkyDeploy resource
//     - If SkyApp object exists if so
//     -- Set the OwnerReferences of SkyDeploy to SkyApp
//     TODO: Run algorithm to update the SkyDeploy locations
//     ---- Requires SkyDeploymentList, Dataflow (owner of SkyApp) object (find using SkyApp name or ownerReferences)
//     - If SkyApp does not exist:
//     -- Nothing
//  2. Fetch the SkyApp (dataflow created the object)
//     2.1 Retrive the skydeployments and if they have same SkyApp name, set the OwnerReferences to SkyApp
//     2.2 Run algorithm to update the SkyDeploy locations
//     ---- Requires SkyDeploymentList, Dataflow (owner of SkyApp) object (find using SkyApp name or ownerReferences)
func (r *SkyDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("[SkyDeploy]")

	deployFound := true
	// Fetch the Deployment
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			deployFound = false
			logger.Info("Not a deployment. Continue ...")
			// return ctrl.Result{}, nil
		} else {
			logger.Error(err, "[WARNING] Failed to get Deployment")
			return ctrl.Result{}, nil
		}
	}

	var skyDeploy corev1alpha1.SkyDeployment
	if deployFound {
		// Check if it has the "manage-by" annotation set to skycluster
		if deployment.Annotations["managed-by"] != "skycluster" {
			return ctrl.Result{}, nil
		}

		// Create SkyDeployment object
		skyDeploy.Spec.DeploymentRef = corev1alpha1.DeploymentRef{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		}
		skyDeploy.ObjectMeta = metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		}

		// check if the deployment has annotation skyappname set
		// if so, find the SkyApp and add it as OwnerReferences
		// TODO: Should we keep reference to SkyApp? Why not deployment?
		if deployment.Annotations["skyappname"] != "" {
			skyDeploy.Spec.AppName = deployment.Annotations["skyappname"]
			skyApp := &corev1alpha1.SkyApp{}
			err = r.Get(ctx, types.NamespacedName{
				Name:      skyDeploy.Spec.AppName,
				Namespace: deployment.Namespace,
			}, skyApp)
			if err == nil {
				skyDeploy.SetOwnerReferences([]metav1.OwnerReference{
					{
						APIVersion: skyApp.APIVersion,
						Kind:       skyApp.Kind,
						Name:       skyApp.Name,
						UID:        skyApp.UID,
					},
				})
				// TODO: Run algorithm to update the SkyDeploy locations
			} else {
				// set the owner reference to the deployment temporarily
				// so it gets deleted if the deployment is deleted
				skyDeploy.SetOwnerReferences([]metav1.OwnerReference{
					{
						APIVersion: deployment.APIVersion,
						Kind:       deployment.Kind,
						Name:       deployment.Name,
						UID:        deployment.UID,
					},
				})
			}
		} else {
			logger.Info("SkyApp not found. Skip OwnerReferences.")
			// logger.Error(err, "SkyApp not found. Skip OwnerReferences.")
		}
	}

	skyAppFound := true
	// Not a deployment, we look for SkyApp
	skyApp := &corev1alpha1.SkyApp{}
	err = r.Get(ctx, req.NamespacedName, skyApp)
	if err != nil {
		if errors.IsNotFound(err) {
			skyAppFound = false
			// logger.Info("Not a SkyApp. Continue...")
		} else {
			logger.Error(err, "[WARNING] Failed to get SkyApp")
			return ctrl.Result{}, nil
		}
	}

	if skyAppFound {
		logger.Info("SkyApp Found", "Name", skyApp.Name)
		// Get the SkyDeployments and if they have the same SkyApp name, set the OwnerReferences to SkyApp
		skyDeployList := &corev1alpha1.SkyDeploymentList{}
		selector := client.MatchingFields{
			"spec.appName": skyApp.Name,
		}
		err = r.List(ctx, skyDeployList, selector)
		if err != nil {
			logger.Error(err, "[WARNING] Failed to list SkyDeployments")
			return ctrl.Result{}, nil
		}
		// print the number of resources found
		logger.Info("SkyDeployments Found. Setting the owner...", "Count", len(skyDeployList.Items))
		for _, sd := range skyDeployList.Items {
			sd.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: skyApp.APIVersion,
					Kind:       skyApp.Kind,
					Name:       skyApp.Name,
					UID:        skyApp.UID,
				},
			})
			err = r.Update(ctx, &sd)
			if err != nil {
				logger.Error(err, "[WARNING] Failed to update SkyDeployment")
			}
		}
		// Run algorithm to update the SkyDeploy locations
		// Requires SkyDeploymentList, Dataflow (owner of SkyApp) object (find using SkyApp name or ownerReferences)
		// ownerRef := metav1.GetOwnerReferences(skyApp)
		ownerRef := skyApp.GetOwnerReferences()
		if len(ownerRef) > 0 {
			// assume the first owner is the dataflow
			logger.Info("Dataflow Found (ownership)", "Name", ownerRef[0].Name)
			// Get the owner resource
			dataflow := &corev1alpha1.Dataflow{}
			err = r.Get(ctx, client.ObjectKey{Namespace: skyApp.Namespace, Name: ownerRef[0].Name}, dataflow)
			if err != nil {
				logger.Error(err, "[WARNING] Failed to get Dataflow")
			}
			logger.Info("Run the algorithm...")
			buildPEGraph(skyDeployList, dataflow)
			logger.Info("The algorithm completed. Updating the SkyDeployments...")
			for _, sd := range skyDeployList.Items {
				// logger.Info("Algorithm results:", "skydeploy", sd.Spec.DeployLocation)
				// update each skydeploy
				err = r.Update(ctx, &sd)
				if err != nil {
					logger.Error(err, "[WARNING] Failed to update SkyDeployment to add locations")
				}
			}
		} else {
			logger.Info("Dataflow not found. Skip the algorithm...")
		}
	} else if deployFound {
		// Create the object
		err = r.Create(ctx, &skyDeploy)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				logger.Info("SkyDeploy already exists")
			} else {
				logger.Error(err, "[WARNING] Failed to create SkyDeploy")
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkyDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Set the index for the SkyDeployment
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &corev1alpha1.SkyDeployment{}, "spec.appName", func(rawObj client.Object) []string {
			// grab the job object, extract the owner...
			skydeploy := rawObj.(*corev1alpha1.SkyDeployment)
			return []string{skydeploy.Spec.AppName}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.SkyDeployment{}).
		Watches(
			&appsv1.Deployment{},
			&handler.EnqueueRequestForObject{},
		).
		Watches(
			&corev1alpha1.SkyApp{},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}

var (
	providers = []*corev1alpha1.DeployLocation{
		&corev1alpha1.DeployLocation{
			Name:     "aws-us-west-2",
			Provider: "AWS",
			Region:   "us-west-2",
			Location: "cloud",
		},
		&corev1alpha1.DeployLocation{
			Name:     "savi-scinet",
			Provider: "SAVI",
			Region:   "SCINET",
			Location: "nte",
		},
		&corev1alpha1.DeployLocation{
			Name:     "savi-vaughan",
			Provider: "SAVI",
			Region:   "VAUGHAN",
			Location: "edge",
		},
	}
)

func buildPEGraph(skyDeployList *corev1alpha1.SkyDeploymentList, dataflow *corev1alpha1.Dataflow) {
	// Define locations
	locations := []*pegraph.Location{
		&pegraph.Location{Name: providers[0].Name},
		&pegraph.Location{Name: providers[1].Name},
		&pegraph.Location{Name: providers[2].Name},
	}

	// Define nodes
	nodes := []*pegraph.Node{}
	// Add skyDeploy items to the nodes
	for _, sd := range skyDeployList.Items {
		fmt.Println("Adding node: ", sd.Name)
		nodes = append(nodes, &pegraph.Node{Name: sd.Name, Location: nil})
	}

	// nodeA := &pegraph.Node{Name: "A", Location: nil}
	// nodeB := &pegraph.Node{Name: "B", Location: nil}
	// nodeC := &pegraph.Node{Name: "C", Location: nil}
	// nodeD := &pegraph.Node{Name: "D", Location: nil}
	// nodeE := &pegraph.Node{Name: "E", Location: nil}
	// nodeF := &pegraph.Node{Name: "F", Location: nil}

	// Create the initial graph
	skygraph := &pegraph.Graph{
		Nodes: nodes,
		Edges: make(map[string][]string),
	}

	connections := dataflow.Spec.DataFlowConnections
	// iterate over the connections and add the edge from source to destination
	for _, conn := range connections {
		fmt.Println("Adding edge: ", conn.ConnectionSource.Name, " -> ", conn.ConnectionDestination.Name)
		skygraph.Edges[conn.ConnectionSource.Name] = append(skygraph.Edges[conn.ConnectionSource.Name], conn.ConnectionDestination.Name)
	}

	// Add initial edges
	// skygraph.Edges[nodeA.Name] = append(skygraph.Edges[nodeA.Name], nodeB.Name) // A->B
	// skygraph.Edges[nodeB.Name] = append(skygraph.Edges[nodeB.Name], nodeC.Name) // B->C
	// skygraph.Edges[nodeB.Name] = append(skygraph.Edges[nodeB.Name], nodeE.Name) // B->E
	// skygraph.Edges[nodeC.Name] = append(skygraph.Edges[nodeC.Name], nodeD.Name) // C->D
	// skygraph.Edges[nodeC.Name] = append(skygraph.Edges[nodeC.Name], nodeF.Name) // C->F
	// skygraph.Edges[nodeD.Name] = append(skygraph.Edges[nodeD.Name], nodeE.Name) // D->E
	// skygraph.Edges[nodeE.Name] = append(skygraph.Edges[nodeE.Name], nodeF.Name) // E->F

	// Define performance data
	perfData := map[string]pegraph.PerfData{}
	// perfData := map[string]pegraph.PerfData{
	// 	nodeA.Name: {Latency: 10, Bandwidth: 100},
	// 	nodeB.Name: {Latency: 20, Bandwidth: 200},
	// 	nodeC.Name: {Latency: 30, Bandwidth: 300},
	// 	nodeD.Name: {Latency: 30, Bandwidth: 300},
	// 	nodeE.Name: {Latency: 30, Bandwidth: 300},
	// 	nodeF.Name: {Latency: 30, Bandwidth: 300},
	// }

	// Define location registry
	locationReqList := map[string][]*pegraph.Location{
		nodes[0].Name: {locations[2]},
		nodes[2].Name: {locations[0]},
	}

	// Define location allowlist
	locationAllowlist := map[string][]*pegraph.Location{
		nodes[0].Name: {locations[2]},
		nodes[1].Name: {locations[1], locations[2]},
		nodes[2].Name: {locations[0]},
	}

	// Generate the initial policy-enriched application graph
	// skygraph.PrintGraph()
	fmt.Println("---- ---- ----")
	initialPEAGraph := skygraph.GenerateInitialPEAGraph(locationReqList, locationAllowlist, perfData, locations)

	// initialPEAGraph.PrintGraph()
	// fmt.Println("---- ---- ----")

	// Generate the final policy-enriched application graph
	initialPEAGraph.GeneratePEAGraph(skygraph, locationAllowlist)
	// initialPEAGraph.PrintGraph()
	initialPEAGraph.DrawGraph()

	// Based on the generated graph, find the location of each node
	for _, node := range initialPEAGraph.Nodes {
		fmt.Println("Node: ", node.Name, " Location: ", node.Location.Name)
		// find provider and location given the name from the list
		for i := range skyDeployList.Items {
			if skyDeployList.Items[i].Name == node.Name {
				// fmt.Println("skydeploy Name:", sd.Name)
				for _, loc := range providers {
					if loc.Name == node.Location.Name {
						fmt.Println("Location Name:", loc.Name, " Provider:", loc.Provider, " Region:", loc.Region, " Location:", loc.Location)
						skyDeployList.Items[i].Spec.DeployLocation = corev1alpha1.DeployLocation{
							Name:     loc.Name,
							Provider: loc.Provider,
							Region:   loc.Region,
							Location: loc.Location,
						}
					}
				}
			}
		}

	}

}
