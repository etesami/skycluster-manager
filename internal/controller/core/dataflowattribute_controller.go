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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// DataflowAttributeReconciler reconciles a DataflowAttribute object
type DataflowAttributeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=policy.skycluster-manager.savitestbed.ca,resources=dataflowattributes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy.skycluster-manager.savitestbed.ca,resources=dataflowattributes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy.skycluster-manager.savitestbed.ca,resources=dataflowattributes/finalizers,verbs=update

// Reconcile reconciles the DataflowAttribute object
func (r *DataflowAttributeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("DataflowAttr [" + req.Name + "] Reconciler started")

	dataflowattr := &corev1alpha1.DataflowAttribute{}
	err := r.Get(ctx, req.NamespacedName, dataflowattr)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("DataflowAttr [" + req.Name + "] not found. Why?")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch ["+req.Name+"]")
	}

	// Check if ILPTask exists, if not create it
	// Update the object with reference to dataflowattr object

	ilptask := &corev1alpha1.ILPTask{}
	err = r.Get(ctx, client.ObjectKey{
		Namespace: dataflowattr.Namespace,
		Name:      dataflowattr.Spec.AppName,
	}, ilptask)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("DataflowAttr [" + dataflowattr.Spec.AppName + "] not found. Creating it...")
			ilptask.ObjectMeta.Name = dataflowattr.Spec.AppName
			ilptask.ObjectMeta.Namespace = dataflowattr.Namespace
			ilptask.Spec.AppName = dataflowattr.Spec.AppName
			ilptask.Spec.ProblemDefinition = "import time; print('Optimizer running...'); time.sleep(5); print('Optimizer completed')"
			ilptask.Spec.DataflowAttributeRef.Name = dataflowattr.Name
			ilptask.Spec.DataflowAttributeRef.Namespace = dataflowattr.Namespace
			err = r.Create(ctx, ilptask)
			if err != nil {
				// check if error is already exists error
				if errors.IsAlreadyExists(err) {
					log.Info("DataflowAttr [" + dataflowattr.Spec.AppName + "] already exists, checking the references")
					// Update the object with reference to dataflowattr object
					if err := r.Get(ctx, client.ObjectKey{
						Namespace: dataflowattr.Namespace,
						Name:      dataflowattr.Spec.AppName,
					}, ilptask); err != nil {
						log.Error(err, "Failed to fetch ILPTask ["+dataflowattr.Spec.AppName+"]")
						return ctrl.Result{}, err
					}
					if ilptask.Spec.DataflowAttributeRef.Name != dataflowattr.Name ||
						ilptask.Spec.DataflowAttributeRef.Namespace != dataflowattr.Namespace {
						log.Info("DataflowAttr [" + dataflowattr.Spec.AppName + "] references are not correct, updating it...")
						ilptask.Spec.DataflowAttributeRef.Name = dataflowattr.Name
						ilptask.Spec.DataflowAttributeRef.Namespace = dataflowattr.Namespace
						if err := r.Update(ctx, ilptask); err != nil {
							log.Error(err, "Failed to update ILPTask ["+dataflowattr.Spec.AppName+"]")
							return ctrl.Result{}, err
						}
						return ctrl.Result{}, nil
					}
					return ctrl.Result{}, nil
				}
				log.Error(err, "Failed to create ILPTask ["+dataflowattr.Spec.AppName+"]")
				return ctrl.Result{}, err
			}
			log.Info("DataflowAttr [" + dataflowattr.Spec.AppName + "] created successfully")
		} else {
			log.Error(err, "Failed to fetch ILPTask ["+dataflowattr.Spec.AppName+"]")
			return ctrl.Result{}, err
		}
	} else {
		// Update the object with reference to dataflowattr object
		ilptask.Spec.DataflowAttributeRef.Name = dataflowattr.Name
		ilptask.Spec.DataflowAttributeRef.Namespace = dataflowattr.Namespace
		if err := r.Update(ctx, ilptask); err != nil {
			log.Error(err, "Failed to update ILPTask ["+dataflowattr.Spec.AppName+"]")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataflowAttributeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.DataflowAttribute{}).
		Complete(r)
}
