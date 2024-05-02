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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
//     - If SkyApp does not exist:
//     -- Nothing
//  2. Fetch the SkyApp (dataflow created the object)
//     2.1. Retrive the skydeployments and if they have same SkyApp name, set the OwnerReferences to SkyApp
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
			Name:      "sky-" + deployment.Name,
			Namespace: deployment.Namespace,
		}

		// check if the deployment has annotation skyappname set
		// if so, find the SkyApp and add it as OwnerReferences
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
			} else {
				logger.Info("SkyApp not found. Skip OwnerReferences.")
				// logger.Error(err, "SkyApp not found. Skip OwnerReferences.")
			}
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

func myHandler(_ context.Context, o client.Object) []reconcile.Request {
	// If this is not a managed Service we want to enqueue it
	logger := log.FromContext(context.Background()).WithName("[SkyDeploy Handler]")
	logger.Info("Dataflow Found", "Name", o.GetName(), "labels", o.GetLabels())
	// Check the lable and reconcile any SkyDeployment with the same lable
	if o.GetLabels()["app"] == "skycluster" {
		logger.Info("Reconciling for", "Name", o.GetName(), "namespace", o.GetNamespace())
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Namespace: o.GetNamespace(),
					Name:      o.GetName(),
				},
			},
		}
	}
	return nil
}
