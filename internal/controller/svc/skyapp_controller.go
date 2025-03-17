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

package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
	svcv1alpha1 "github.com/etesami/skycluster-manager/api/svc/v1alpha1"
	"github.com/pkg/errors"
)

// SkyAppReconciler reconciles a SkyApp object
type SkyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=svc.skycluster.io,resources=skyapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=svc.skycluster.io,resources=skyapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=svc.skycluster.io,resources=skyapps/finalizers,verbs=update

func (r *SkyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	loggerName := "SkyApp"
	logger.Info(fmt.Sprintf("[%s]\t Reconciling SkyApp for [%s]", loggerName, req.Name))

	skyApp := &svcv1alpha1.SkyApp{}
	if err := r.Get(ctx, req.NamespacedName, skyApp); err != nil {
		logger.Info(fmt.Sprintf("[%s]\t SkyApp not found.", loggerName))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// We need to create deployment and services for the remote cluster.
	// We have the base manifest in spec.manifest and we need to create
	// the Kubernetes Provider "objects" for the remote cluster.
	// Additionally istio configuration should be generated.

	// Creating Kubernetes Provider Objects for deployments and services
	// We need to watch for providerCfgName and when it is available,
	// proceed with the creation of the deployment and services and etc.
	providerCfgName := "providerCfgName"
	manifests := generateDeployObjectManifests(skyApp.Spec.Manifests, providerCfgName)

	// Creating Istio configuration
	manifestCfg := generateIstioConfig(skyApp.Spec.Manifests, providerCfgName)

	// Update the status with the objects that will be created
	for _, obj := range manifests {
		yamlObj, err := createYAMLManifest(obj.Object, obj.GetName(), obj.GetAPIVersion(), obj.GetKind())
		if err != nil {
			logger.Error(err, fmt.Sprintf("[%s]\t Error creating YAML manifest.", loggerName))
			return ctrl.Result{}, err
		}
		skyApp.Status.Objects = append(skyApp.Status.Objects, corev1alpha1.SkyService{
			Name:       yamlObj.Name,
			Kind:       yamlObj.Kind,
			APIVersion: yamlObj.APIVersion,
			Manifest:   yamlObj.Manifest,
		})
	}

	// Update the status with the objects that will be created
	for _, obj := range manifestCfg {
		yamlObj, err := createYAMLManifest(obj.Object, obj.GetName(), obj.GetAPIVersion(), obj.GetKind())
		if err != nil {
			logger.Error(err, fmt.Sprintf("[%s]\t Error creating YAML manifest.", loggerName))
			return ctrl.Result{}, err
		}
		skyApp.Status.Objects = append(skyApp.Status.Objects, corev1alpha1.SkyService{
			Name:       yamlObj.Name,
			Kind:       yamlObj.Kind,
			APIVersion: yamlObj.APIVersion,
			Manifest:   yamlObj.Manifest,
		})
	}

	if err := r.Status().Update(ctx, skyApp); err != nil {
		logger.Error(err, fmt.Sprintf("[%s]\t Error updating status.", loggerName))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createYAMLManifest creates a SkyService object from the given object
// and marshals it into a YAML manifest
func createYAMLManifest(obj any, name, kind, apiVersion string) (*corev1alpha1.SkyService, error) {
	// It appears that if we proceed with the obj itself, the YAML
	// is verbose and does not follow the format of the object (e.g. Deployment)
	// But we can use json.Marshal and then yaml.Marshal as a workaround
	var inInterface map[string]interface{}
	inrec, _ := json.Marshal(obj)
	json.Unmarshal(inrec, &inInterface)
	objYAML, err := yaml.Marshal(&inInterface)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshalling obj manifests.")
	}
	return &corev1alpha1.SkyService{
		Name:       name,
		Kind:       kind,
		APIVersion: apiVersion,
		Manifest:   string(objYAML),
	}, nil
}

// getUniqueProviders returns a list of unique providers from the given manifests
func getUniqueProviders(manifests []corev1alpha1.SkyService) []corev1alpha1.ProviderRefSpec {
	pExists := map[string]any{}
	providers := []corev1alpha1.ProviderRefSpec{}
	for _, manifest := range manifests {
		pID := fmt.Sprintf("%s-%s-%s",
			manifest.ProviderRef.ProviderName,
			manifest.ProviderRef.ProviderRegion,
			manifest.ProviderRef.ProviderZone)
		if _, ok := pExists[pID]; !ok {
			providers = append(providers, manifest.ProviderRef)
			pExists[pID] = struct{}{}
		}
	}
	return providers
}

// sameProviders returns true if the two providers are the same
// based on the provider name, region and zone
func sameProviders(p1, p2 corev1alpha1.ProviderRefSpec) bool {
	return p1.ProviderName == p2.ProviderName &&
		p1.ProviderRegion == p2.ProviderRegion &&
		p1.ProviderZone == p2.ProviderZone
}

// getProviderId returns a unique identifier for the provider
// based on the provider name, region, zone and type
func getProviderId(p corev1alpha1.ProviderRefSpec) string {
	return fmt.Sprintf("%s-%s-%s-%s", p.ProviderName, p.ProviderRegion, p.ProviderZone, p.ProviderType)
}

func generateDeployObjectManifests(manifests []corev1alpha1.SkyService, providerCfgName string) map[string]unstructured.Unstructured {
	objs := map[string]unstructured.Unstructured{}
	// Generate deployment and service manifests
	for _, manifest := range manifests {
		// for deployment and services we only wrap them in "Object"
		if strings.ToLower(manifest.Kind) == "deployment" || strings.ToLower(manifest.Kind) == "service" {
			obj := &unstructured.Unstructured{}
			obj.SetAPIVersion("kubernetes.crossplane.io/v1alpha2")
			obj.SetKind("Object")
			obj.SetName(manifest.Name)
			obj.SetLabels(map[string]string{
				corev1alpha1.SKYCLUSTER_MANAGEDBY_LABEL: corev1alpha1.SKYCLUSTER_MANAGEDBY_VALUE,
			})
			obj.Object["spec"] = map[string]interface{}{
				"forProvider": map[string]interface{}{
					"manifest": manifest.Manifest,
				},
				"providerConfigRef": map[string]string{
					"name": providerCfgName,
				},
			}
			objs[manifest.Name] = *obj
		}
	}
	return objs
}

func generateIstioConfig(manifests []corev1alpha1.SkyService, providerCfgName string) map[string]unstructured.Unstructured {
	allProviders := getUniqueProviders(manifests)
	objs := map[string]unstructured.Unstructured{}
	// Generate Istio configuration
	// For each service we should create a DestinationRule
	// and set the localityLbSetting to enable
	// Then we set the failover attributes to all other
	// providers we are allowed to send traffic
	for _, manifest := range manifests {
		if strings.ToLower(manifest.Kind) == "service" {

			failovers := []map[string]string{}
			// In default settings, we allow traffic to all other providers,
			for _, p1 := range allProviders {
				for _, p2 := range allProviders {
					if !sameProviders(p1, p2) {
						failovers = append(failovers, map[string]string{
							"from": getProviderId(p1),
							"to":   getProviderId(p2),
						})
						failovers = append(failovers, map[string]string{
							"to":   getProviderId(p1),
							"from": getProviderId(p2),
						})
					}
				}
			}

			istioObj := map[string]interface{}{
				"apiVersion": "networking.istio.io/v1",
				"kind":       "DestinationRule",
				"metadata": map[string]interface{}{
					"name": manifest.Name,
				},
				"spec": map[string]interface{}{
					"host": manifest.Name,
					"trafficPolicy": map[string]interface{}{
						"loadBalancer": map[string]interface{}{
							"simple": "LEAST_REQUEST",
							"localityLbSetting": map[string]interface{}{
								"enabled":  true,
								"failover": failovers,
							},
						},
					},
				},
			}

			obj := &unstructured.Unstructured{}
			obj.SetAPIVersion("kubernetes.crossplane.io/v1alpha2")
			obj.SetKind("Object")
			obj.SetName(manifest.Name)
			obj.SetLabels(map[string]string{
				corev1alpha1.SKYCLUSTER_MANAGEDBY_LABEL: corev1alpha1.SKYCLUSTER_MANAGEDBY_VALUE,
			})
			obj.Object["spec"] = map[string]interface{}{
				"forProvider": map[string]interface{}{
					"manifest": istioObj,
				},
				"providerConfigRef": map[string]string{
					"name": providerCfgName,
				},
			}

			objs[manifest.Name] = *obj
		}
	}
	return objs
}

// SetupWithManager sets up the controller with the Manager.
func (r *SkyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&svcv1alpha1.SkyApp{}).
		Named("svc-skyapp").
		Complete(r)
}
