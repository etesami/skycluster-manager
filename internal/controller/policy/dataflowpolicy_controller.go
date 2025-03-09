/*
Copyright 2025.

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

package policy

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
	policyv1alpha1 "github.com/etesami/skycluster-manager/api/policy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataflowPolicyReconciler reconciles a DataflowPolicy object
type DataflowPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=policy.skycluster.io,resources=dataflowpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy.skycluster.io,resources=dataflowpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy.skycluster.io,resources=dataflowpolicies/finalizers,verbs=update

func (r *DataflowPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	loggerName := "DFPolicy"
	logger.Info(fmt.Sprintf("[%s]\t Reconciling DataflowPolicy for [%s]", loggerName, req.Name))

	dfPolicy := &policyv1alpha1.DataflowPolicy{}
	if err := r.Get(ctx, req.NamespacedName, dfPolicy); err != nil {
		logger.Info(fmt.Sprintf("[%s]\t DataflowPolicy not found.", loggerName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Append the name of the DataflowPolicy to the skycluster.core.skycluster.io/v1alpha1 object
	// If the SkyCluster object does not exist, create it and then append the name

	skyCluster := &corev1alpha1.SkyCluster{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: dfPolicy.Namespace,
		Name:      dfPolicy.Name,
	}, skyCluster); err != nil {
		logger.Info(fmt.Sprintf("[%s]\t SkyCluster not found. Creating it.", loggerName))
		skyCluster = &corev1alpha1.SkyCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dfPolicy.Name,
				Namespace: dfPolicy.Namespace,
			},
			Spec: corev1alpha1.SkyClusterSpec{
				DataflowPolicyRef: corev1alpha1.ResourceSpec{
					Name: dfPolicy.Name,
				},
			},
		}
		if err := r.Create(ctx, skyCluster); err != nil {
			logger.Error(err, fmt.Sprintf("[%s]\t Failed to create SkyCluster.", loggerName))
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// The SkyCluster object exists, update it by appending the name of the DataflowPolicy
	logger.Info(fmt.Sprintf("[%s]\t Updating SkyCluster.", loggerName))
	skyCluster.Spec.DataflowPolicyRef.Name = dfPolicy.Name
	if err := r.Update(ctx, skyCluster); err != nil {
		logger.Error(err, fmt.Sprintf("[%s]\t Failed to update SkyCluster.", loggerName))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataflowPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policyv1alpha1.DataflowPolicy{}).
		Named("policy-dataflowpolicy").
		Complete(r)
}
