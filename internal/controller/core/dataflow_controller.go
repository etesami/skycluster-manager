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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// DataflowReconciler reconciles a Dataflow object
type DataflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=dataflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=dataflows/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=dataflows/finalizers,verbs=update
////+kubebuilder:rbac:groups=core.skycluster.savitestbed.ca,resources=skyapps,verbs=get;list;watch;create;update;patch;delete

// Reconcile functions workflow:
// 1. Fetch the Dataflow
// 2. Create a new SkyApp object and set OwnerReferences, we do not expect to see SkyApp already created
// 3. Get the skyapp and if it does not exist create it (this trigers the reconcilation of SkyDeplopoyments)
func (r *DataflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("[Dataflow]")

	// Fetch the Dataflow
	dataflow := &corev1alpha1.Dataflow{}
	err := r.Get(ctx, req.NamespacedName, dataflow)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "[WARNING] Failed to get Dataflow")
		return ctrl.Result{}, err
	}

	logger.Info("Dataflow Found", "Name", dataflow.Name, "AppName", dataflow.Spec.AppName)

	// Create a new SkyApp object and set OwnerReferences
	skyApp := &corev1alpha1.SkyApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataflow.Spec.AppName,
			Namespace: dataflow.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: dataflow.APIVersion,
					Kind:       dataflow.Kind,
					Name:       dataflow.Name,
					UID:        dataflow.UID,
				},
			},
		},
	}

	skyAppFound := true
	// get the skyapp and if it does not exist create it
	err = r.Get(ctx, client.ObjectKeyFromObject(skyApp), skyApp)
	if err != nil {
		if errors.IsNotFound(err) {
			skyAppFound = false
			logger.Info("SkyApp not found, creating a new one")
		} else {
			logger.Error(err, "[WARNING] Failed to get SkyApp")
			return ctrl.Result{}, err
		}
	}

	if !skyAppFound {
		// Create the object
		err = r.Create(ctx, skyApp)
		if err != nil {
			logger.Error(err, "[WARNING] Failed to create SkyApp")
			return ctrl.Result{}, err
		}
	} else {
		logger.Info("SkyApp Found", "Name", skyApp.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Dataflow{}).
		Complete(r)
}
