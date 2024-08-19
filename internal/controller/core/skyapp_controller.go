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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// SkyAppReconciler reconciles a SkyApp object
type SkyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyapps/finalizers,verbs=update

// Reconcile reconciles the SkyApp object
func (r *SkyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("SkyApp [" + req.Name + "] Reconciler started")

	skyapp := &corev1alpha1.SkyApp{}
	err := r.Get(ctx, req.NamespacedName, skyapp)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyApp [" + req.Name + "] not found. Why?")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch ["+req.Name+"]")
	}

	// Check if ILPTask exists, if not create it
	// Update the object with reference to skyapp object

	ilptask := &corev1alpha1.ILPTask{}
	err = r.Get(ctx, client.ObjectKey{
		Namespace: skyapp.Namespace,
		Name:      skyapp.Spec.AppName,
	}, ilptask)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("SkyApp [" + skyapp.Spec.AppName + "] ilptask not found. Creating it...")
			// set annotations
			ilptask.Annotations = make(map[string]string)
			ilptask.Annotations[SkyClusterAnnotationManagedBy] = "skycluster-manager"
			ilptask.Annotations[SkyClusterAnnotationConfigType] = "ilp-task"
			ilptask.Annotations[SkyClusterAnnotationCreationTime] = time.Now().Format(time.RFC3339)
			ilptask.Spec.AppName = skyapp.Spec.AppName
			ilptask.ObjectMeta.Name = skyapp.Spec.AppName
			ilptask.ObjectMeta.Namespace = skyapp.Namespace
			ilptask.Spec.ProblemDefinition = "import time; print('Optimizer running...'); time.sleep(5); print('Optimizer completed')"
			ilptask.Spec.SkyAppRef.Name = skyapp.Name
			ilptask.Spec.SkyAppRef.Namespace = skyapp.Namespace
			if err = controllerutil.SetControllerReference(skyapp, ilptask, r.Scheme); err != nil {
				log.Error(err, "Failed to set controller reference for ILPTask ["+skyapp.Spec.AppName+"]")
				return ctrl.Result{}, err
			}
			err = r.Create(ctx, ilptask)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					log.Info("ILPTask [" + skyapp.Spec.AppName + "] already exists")
					return ctrl.Result{}, nil
				}
				log.Error(err, "Failed to create ILPTask ["+skyapp.Spec.AppName+"]")
				return ctrl.Result{}, err
			}
			// log.Info("SkyApp [" + skyapp.Spec.AppName + "] created successfully")
		} else {
			log.Error(err, "Failed to fetch ILPTask ["+skyapp.Spec.AppName+"]")
			return ctrl.Result{}, err
		}
	} else {
		// Update the object with reference to skyapp object
		if ilptask.Annotations == nil {
			ilptask.Annotations = make(map[string]string)
		}
		ilptask.Annotations[SkyClusterAnnotationManagedBy] = "skycluster-manager"
		ilptask.Annotations[SkyClusterAnnotationConfigType] = "ilp-task"
		ilptask.Annotations[SkyClusterAnnotationCreationTime] = time.Now().Format(time.RFC3339)
		ilptask.Spec.SkyAppRef.Name = skyapp.Name
		ilptask.Spec.SkyAppRef.Namespace = skyapp.Namespace
		if err = controllerutil.SetControllerReference(skyapp, ilptask, r.Scheme); err != nil {
			log.Error(err, "Failed to set controller reference for ILPTask ["+skyapp.Spec.AppName+"]")
			return ctrl.Result{}, err
		}
		err = r.Update(ctx, ilptask)
		if err != nil {
			log.Error(err, "Failed to update ILPTask ["+skyapp.Spec.AppName+"]")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.SkyApp{}).
		Complete(r)
}
