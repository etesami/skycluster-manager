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

package v1alpha1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var skyapplog = logf.Log.WithName("skyapp-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *SkyApp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		// For(r).
		For(&appsv1.Deployment{}).
		WithDefaulter(&SkyApp{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-apps-v1-deployment,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update;delete,versions=v1,name=mdeployment.kb.io,admissionReviewVersions=v1

// var _ webhook.Defaulter = &SkyApp{}
var _ webhook.CustomDefaulter = &SkyApp{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SkyApp) Default(ctx context.Context, obj runtime.Object) error {
	// func (r *SkyApp) Default() {

	deploy, ok := obj.(*appsv1.Deployment)

	if !ok {
		skyapplog.Info("Deployment received", "Deployment", obj)
		return nil
	}

	skyAnnotation, exists := deploy.Annotations["skycluster-manager.savitestbed.ca/managed-by"]
	if exists && skyAnnotation == "skycluster" {
		skyapplog.Info("Contains annotation", "Annotation", skyAnnotation)
		// set scheduler to skycluster scheduler
		deploy.Spec.Template.Spec.SchedulerName = "dummy-scheduler"
	}

	return nil
}
