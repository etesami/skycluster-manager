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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=providers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=providers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=providers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Provider object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Provider [" + req.Name + "] Reconciler started")

	provider := &corev1alpha1.Provider{}
	err := r.Get(ctx, req.NamespacedName, provider)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Provider [" + req.Name + "] not found. Why?")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch ["+req.Name+"]")
	}

	providersData := ""
	providersData += provider.Spec.Name + ","
	providersData += provider.Spec.Region + ","
	providersData += provider.Spec.Zone + ","
	providersData += provider.Spec.Type

	if err = r.createConfigMap(ctx, provider.Spec.AppName+"-providers", providersData, provider); err != nil {
		log.Error(err, "Unable to create ConfigMap for providers")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// write a function to create a configmap given the content of the file
func (r *ProviderReconciler) createConfigMap(ctx context.Context, name string, content string, provider *corev1alpha1.Provider) error {
	log := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: provider.Namespace}, cm)
	if err == nil {
		log.Info("Provider: ConfigMap already exists, not expected.")
		return nil
	}

	// Define a new ConfigMap object
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: provider.Namespace,
		},
		Data: map[string]string{
			name: content,
		},
	}

	// Set MyResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(provider, cm, r.Scheme); err != nil {
		log.Error(err, "Provider: Failed to set owner reference on ConfigMap")
		return err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Provider: Creating a new ConfigMap")
		err = r.Create(ctx, cm)
		if err != nil {
			log.Error(err, "Provider: Failed to create new ConfigMap")
			return err
		}
		// ConfigMap created successfully - return and requeue
		return nil
	}
	log.Info("Provider: Should not be here. CM should not exist at this point.")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.Provider{}).
		Complete(r)
}
