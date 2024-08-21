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
	"encoding/json"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/etesami/skycluster-manager/api/core/v1alpha1"
)

// ILPTaskReconciler reconciles a ILPTask object
type ILPTaskReconciler struct {
	client.Client
	Clientset *corev1client.CoreV1Client
	Scheme    *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=ilptasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=ilptasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=ilptasks/finalizers,verbs=update
//// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyapp,verbs=create;update;patch;delete
//// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=dataflowattribute,verbs=create;update;patch;delete

// Reconcile reconciles the ILPTask object
func (r *ILPTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// print the name and namespace of the ILPTask
	log.Info("ILPTask [" + req.Name + "] Reconciler started")

	// Fetch the ILPTask instance
	ilptask := &corev1alpha1.ILPTask{}
	err := r.Get(ctx, req.NamespacedName, ilptask)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch ILPTask, something is wrong.")
		return ctrl.Result{}, err
	}

	// TODO: The ILPTask may be created by SkyApp or DataflowAttribute
	// We need to check it any other references exist for this object
	// Scenario: SkyApp and Dataflow are created, then SkyApp is deleted
	// Then SkyApp is created again, therefor, the ILPTask should check
	// if DataflowAttribute exists and if it does, it should add a reference

	// Check if both SkyApp and DataflowAttribute references are set
	if ilptask.Spec.DataflowAttributeRef == (corev1alpha1.DataflowAttributeRef{}) ||
		ilptask.Spec.SkyAppRef == (corev1alpha1.SkyAppRef{}) {
		log.Info("ILPTask [" + req.Name + "] SkyApp or DataflowAttribute references are not set")
		return ctrl.Result{}, nil
	} else {
		log.Info("ILPTask [" + req.Name + "] SkyApp and DataflowAttribute references are set")
	}

	// At this point we have both SkyApp and DataflowAttribute references

	// Fetch the SkyApp instance
	// SkyApp may or may not exist
	skyapp := &corev1alpha1.SkyApp{}
	if err = r.Get(ctx, types.NamespacedName{
		Name:      ilptask.Spec.SkyAppRef.Name,
		Namespace: ilptask.Spec.SkyAppRef.Namespace,
	}, skyapp); err == nil {
		log.Info("ILPTask [" + req.Name + "]: SkyApp exists and was retrived")
	} else {
		log.Error(err, "ILPTask ["+req.Name+"]: Unable to fetch SkyApp, TODO: INVESTIGATE this case.")
	}

	// Fetch the DataflowAttribute instance
	dataflowattr := &corev1alpha1.DataflowAttribute{}
	if err = r.Get(ctx, types.NamespacedName{
		Name:      ilptask.Spec.DataflowAttributeRef.Name,
		Namespace: ilptask.Spec.DataflowAttributeRef.Namespace,
	}, dataflowattr); err == nil {
		log.Info("ILPTask [" + req.Name + "]: DataflowAttribute exists and was retrived")
	} else {
		log.Error(err, "ILPTask ["+req.Name+"]: Unable to fetch DataflowAttribute, TODO: INVESTIGATE this case.")
	}

	// Logic to run the optimizer
	// Check if the optimization is already completed
	// if not, check if any pod is running
	// if not, create a pod to run the optimizer

	// Check if the optimization is already completed
	// if it is completed, the Status is not nil
	if ilptask.Status.Result != "" {
		// TODO: If the ilptask is changed, we may need to recalulate the optimization
		// For now, we are assuming that the optimization is done once and the result is final
		// Cases to consider:
		// 1. SkyApp or DataflowAttribute are updated: the optimization should be recalculated
		// 2. Changes within underlay resources
		//    - VirtualService/Region is not available
		//      In this case, a vistual service (and the provider) should reflect this failure and we should remove
		//      the virtual service from the optimization.
		//      Someone flags the virtual service (it is a configmap now)
		//              The ilptask shoould be flagged as incompleted
		//              The reconcile function and the optimizer should be re-run
		log.Info("ILPTask [" + req.Name + "] task already completed or has a result")

		skyxrd := &corev1alpha1.SkyXRD{}
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: skyapp.Namespace,
			Name:      skyapp.Spec.AppName,
		}, skyxrd); err != nil {
			if errors.IsNotFound(err) {
				// SkyXRD doesn't exist, create it
				log.Info("ILPTask [" + skyapp.Spec.AppName + "] SkyXRD doesn't exist, creating it")
				if err := r.createSkyXRD(ctx, skyapp, ilptask); err != nil {
					log.Error(err, "Unable to create SkyXRD")
					return ctrl.Result{}, err
				}
			} else {
				log.Error(err, "Failed to fetch SkyXRD, Unknown error occurred")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Define the Pod name
	podName := ilptask.Spec.AppName

	// Check if the Pod exists
	pod := &corev1.Pod{}
	if err := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: ilptask.Namespace}, pod); err != nil {
		if errors.IsNotFound(err) {
			// Pod doesn't exist, create it
			log.Info("ILPTask [" + req.Name + "] Creating optimizer Pod")

			// Build application graph from SkyApp and DataflowAttribute
			taskNames := ""
			for i, thisTask := range skyapp.Spec.AppConfig {
				taskNames += thisTask.Name + ":"
				vservices := ""
				locConstraints := ""
				for j, thisVService := range thisTask.Constraints.VirtualServiceConstraints {
					vservices += thisVService.VirtualServiceName
					if j < len(thisTask.Constraints.VirtualServiceConstraints)-1 {
						vservices += ","
					}
				}
				for j, thisLoc := range thisTask.Constraints.LocationConstraints {
					locName := "" + thisLoc.ProviderName + ","
					locType := "" + thisLoc.ProviderType + ","
					locRegion := "" + thisLoc.Region + ""
					locConstraints += locName + locType + locRegion
					if j < len(thisTask.Constraints.LocationConstraints)-1 {
						locConstraints += "__"
					}
				}
				taskNames += vservices + "__" + locConstraints
				if i < len(skyapp.Spec.AppConfig)-1 {
					taskNames += "\n"
				}
			}
			if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-tasks", taskNames, ilptask); err != nil {
				log.Error(err, "Unable to create ConfigMap for tasks")
				return ctrl.Result{}, err
			}

			edges := ""
			for i, currentEdge := range dataflowattr.Spec.Connections {
				for j, dest := range currentEdge.Destinations {
					edges += currentEdge.Source + ":"
					edges += dest.Name
					edges += "," + dest.Constraints.Latency
					if j < len(currentEdge.Destinations)-1 {
						edges += "\n"
					}
				}
				if i < len(dataflowattr.Spec.Connections)-1 {
					edges += "\n"
				}
			}
			if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-edges", edges, ilptask); err != nil {
				log.Error(err, "Unable to create ConfigMap for edges")
				return ctrl.Result{}, err
			}

			// Get providers data
			// TODO: The providers should be accessible and available
			//    Currently a provider is a configmap and it should contain a field
			//    that indicates if the provider is available or not, we should only select
			//    the available providers.
			// Filter based on the labels
			if providers, err := r.Clientset.ConfigMaps("").List(ctx, metav1.ListOptions{
				LabelSelector: SkyClusterAnnotationManagedBy + "=skycluster," +
					SkyClusterAnnotationConfigType + "=provider-vars," +
					SkyClusterAnnotationProvierZone + "=default",
			}); err != nil {
				log.Error(err, "Unable to retrieve Providers configmaps")
				return ctrl.Result{}, err
			} else {
				// iterate over the providers and create a configmap for each
				providersData := ""
				for i, thisProvider := range providers.Items {
					providerName := thisProvider.GetAnnotations()[SkyClusterAnnotationProvierName]
					providerNameCombined := thisProvider.GetAnnotations()[SkyClusterAnnotationProvierName] +
						thisProvider.GetAnnotations()[SkyClusterAnnotationProvierRegion] +
						thisProvider.GetAnnotations()[SkyClusterAnnotationProvierType]
					providersData += providerNameCombined + "," + providerName + ","
					providersData += thisProvider.GetAnnotations()[SkyClusterAnnotationProvierRegion] + ","
					providersData += thisProvider.GetAnnotations()[SkyClusterAnnotationSkyClusterRegion] + ","
					providersData += thisProvider.GetAnnotations()[SkyClusterAnnotationProvierZone] + ","
					providersData += thisProvider.GetAnnotations()[SkyClusterAnnotationProvierType]
					if i < len(providers.Items)-1 {
						providersData += "\n"
					}
				}
				if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-providers", providersData, ilptask); err != nil {
					log.Error(err, "Unable to create ConfigMap for providers")
					return ctrl.Result{}, err
				}
			}

			// TODO: we may need to discard providers attribute that are not available
			// Fetch the ProviderAttribute instance
			// Assuming there is one instance of this type (part of TODO list)
			providerAttrs := &corev1alpha1.ProviderAttributeList{}
			if err := r.List(ctx, providerAttrs, &client.ListOptions{
				Namespace: ilptask.Namespace,
			}); err != nil {
				log.Error(err, "Unable to get providerAttrs")
				return ctrl.Result{}, err
			}

			// iterate over the providerAttrs and create a configmap for each
			providerAttrData := ""
			for i, providerAttr := range providerAttrs.Items {
				currentNodeName := providerAttr.Spec.ProviderReference.Name +
					providerAttr.Spec.ProviderReference.Region +
					providerAttr.Spec.ProviderReference.Type
				for j, destnationNode := range providerAttr.Spec.ProviderMetrics {
					providerAttrData += currentNodeName + ":"
					providerAttrData += destnationNode.DstProviderRef.Name +
						destnationNode.DstProviderRef.Region +
						destnationNode.DstProviderRef.Type
					providerAttrData += "," + destnationNode.Latency
					providerAttrData += "," + destnationNode.EgressDataCost
					if j < len(providerAttr.Spec.ProviderMetrics)-1 {
						providerAttrData += "\n"
					}
				}
				if i < len(providerAttrs.Items)-1 {
					providerAttrData += "\n"
				}
			}

			if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-providerattr", providerAttrData, ilptask); err != nil {
				log.Error(err, "Unable to create ConfigMap for providerAttr")
				return ctrl.Result{}, err
			}

			// get virtual service data
			// TODO: The virtual services should be accessible and available
			//    Similar to providers we may need to filter the virtual services
			vservices := &corev1alpha1.VirtualServiceList{}
			if err := r.List(ctx, vservices, &client.ListOptions{
				Namespace: ilptask.Namespace,
			}); err != nil {
				log.Error(err, "Unable to get virtual services")
				return ctrl.Result{}, err
			} else {
				// iterate over the vservices and create a configmap for each
				vservicesData := ""
				for i, thisProvider := range vservices.Items {
					for j, thisVServiceProvider := range thisProvider.Spec.VServiceCosts {
						vservicesData += thisProvider.Spec.Name + ":"
						vservicesData += thisVServiceProvider.ProviderRef.Name +
							thisVServiceProvider.ProviderRef.Region +
							thisVServiceProvider.ProviderRef.Type
						vservicesData += "," + thisVServiceProvider.Cost
						if j < len(thisProvider.Spec.VServiceCosts)-1 {
							vservicesData += "\n"
						}
					}
					if i < len(vservices.Items)-1 {
						vservicesData += "\n"
					}
				}
				if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-vservices", vservicesData, ilptask); err != nil {
					log.Error(err, "Unable to create ConfigMap for vservices")
					return ctrl.Result{}, err
				}
			}

			replacedCommand := strings.ReplaceAll(pythonOptimizationCommand, "__FILE_NAME_TASKS__", ilptask.Spec.AppName+"-tasks")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_PROVIDERS__", ilptask.Spec.AppName+"-providers")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_VSERVICES__", ilptask.Spec.AppName+"-vservices")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_EDGES__", ilptask.Spec.AppName+"-edges")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_PROVIDERATTR__", ilptask.Spec.AppName+"-providerattr")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_BASE_SCRIPT__", pythonBaseScript)
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_OPTIMIZATION_BASE__", pythonOptimizationBase)
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_OPTIMIZATION_PROBLEM__", pythonOptimizationProblem)
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_OPTIMIZATION_CONSTRAINTS__", pythonOptimizationConstraints)

			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podName,
					Namespace: ilptask.Namespace,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "optimizer",
							Image: "etesami/optimizer:v1.2",
							Command: []string{
								"python",
								"-c",
								replacedCommand,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: "tasks-cm",
									// This is the mount point inside the container
									MountPath: "/tasks",
								},
								{
									Name:      "providers-cm",
									MountPath: "/providers",
								},
								{
									Name:      "vservices-cm",
									MountPath: "/vservices",
								},
								{
									Name:      "edges-cm",
									MountPath: "/edges",
								},
								{
									Name:      "providerattr-cm",
									MountPath: "/providerattr",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tasks-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										// This is the name of configMap
										Name: ilptask.Spec.AppName + "-tasks",
									},
								},
							},
						},
						{
							Name: "providers-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ilptask.Spec.AppName + "-providers",
									},
								},
							},
						},
						{
							Name: "vservices-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ilptask.Spec.AppName + "-vservices",
									},
								},
							},
						},
						{
							Name: "edges-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ilptask.Spec.AppName + "-edges",
									},
								},
							},
						},
						{
							Name: "providerattr-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ilptask.Spec.AppName + "-providerattr",
									},
								},
							},
						},
					},
				},
			}
			if err = controllerutil.SetControllerReference(skyapp, pod, r.Scheme); err != nil {
				log.Error(err, "Failed to set controller reference for pod")
				return ctrl.Result{}, err
			}
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "Unable to create optimizer Pod")
				return ctrl.Result{}, err
			}
			// Requeue to check the Pod status later
			log.Info("ILPTask [" + req.Name + "] Requeue to check optimizer Pod status")
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
		log.Error(err, "Pod exists but I am unable to get optimizer Pod")
		return ctrl.Result{}, err
	}

	// at this point we may have the results
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		log.Info("ILPTask [" + req.Name + "] Updating ILPTask annotations")
		if ilptask.Annotations == nil {
			ilptask.Annotations = make(map[string]string)
		}
		ilptask.Annotations[SkyClusterAnnotationCompletionTime] = time.Now().Format(time.RFC3339)
		if err = r.Update(ctx, ilptask); err != nil {
			log.Error(err, "Unable to update ILPTask")
			return ctrl.Result{}, err
		}

		if pod.Status.Phase == corev1.PodSucceeded {
			log.Info("ILPTask [" + req.Name + "] Optimizer Pod succeeded!")
			ilptask.Status.Result = "Completed"
		} else {
			log.Info("ILPTask [" + req.Name + "] Optimizer Pod failed")
			ilptask.Status.Result = "Failed"
		}

		if ilptask.Status.Solution, err = r.getPodLogs(ctx, pod); err != nil {
			log.Error(err, "Unable to get Pod logs")
			// return ctrl.Result{}, err
		}
		if err := r.Status().Update(ctx, ilptask); err != nil {
			log.Error(err, "Unable to update ILPTask status")
			return ctrl.Result{}, err
		}
		log.Info("ILPTask [" + req.Name + "] Status updated!")
		log.Info("ILPTask [" + skyapp.Spec.AppName + "] Status: " + ilptask.Status.Result)

		// TODO: When there is a optimization result, we should propagate the result to the SkyXRD object
		// Below there are some comments on how to do this.

		// 1. We should call XProviderSetup XRD for each distinct provider (provider, region, type)
		//    To do so, first we need to update SkyApp with results of the optimization
		//    and then the SkyApp controller take care of creating other XRD objects.
		//    (or maybe ILPTask controller can directly create the an object containing the
		//   	optimization results. This may be better as the tasks are kept separate from the
		//   	SkyApp controller. Will investigate this later.)

		//    For simplicity, I can create a new controller: SkyXRD controller
		//    and a new object of this type is created when the optimization results are ready.

		// TODO: at this point I assume the SkyXRD object does not exist
		// If it exists, it implies that the optimization is already done once
		// and this is either a change in the ILPTask. Need to consider this case later.
		if err := r.createSkyXRD(ctx, skyapp, ilptask); err != nil {
			log.Error(err, "Unable to create SkyXRD")
			return ctrl.Result{}, err
		}

		// Delete the completed Pod
		log.Info("ILPTask [" + req.Name + "] Deleting completed Pod")
		if err := r.Delete(ctx, pod); err != nil {
			log.Error(err, "Unable to delete completed Pod")
			// Don't return an error, as the main task is done
		}
		return ctrl.Result{}, nil
	}

	// Pod is still running, requeue
	return ctrl.Result{RequeueAfter: time.Second * 2}, nil
}

func (r *ILPTaskReconciler) createSkyXRD(ctx context.Context, skyapp *corev1alpha1.SkyApp, ilptask *corev1alpha1.ILPTask) error {
	log := log.FromContext(ctx)
	var result corev1alpha1.TaskPlacement
	// Assuming the pod log is a JSON string (it should be a json string unless something is wrong)
	// Parse the JSON string
	if err := json.Unmarshal([]byte(ilptask.Status.Solution), &result); err != nil {
		log.Error(err, "Unable to parse the JSON string")
		return err
	}

	skyxrd := &corev1alpha1.SkyXRD{
		ObjectMeta: metav1.ObjectMeta{
			Name:      skyapp.Spec.AppName,
			Namespace: skyapp.Namespace,
		},
		Spec: corev1alpha1.SkyXRDSpec{
			AppName:       skyapp.Spec.AppName,
			SkyAppRefName: skyapp.Name,
			TaskPlacement: result,
		},
	}
	if err := controllerutil.SetControllerReference(skyapp, skyxrd, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference on SkyXRD")
		return err
	}
	if err := r.Create(ctx, skyxrd); err != nil {
		log.Error(err, "Failed to create SkyXRD")
		return err
	}
	return nil
}

// write a function to create a configmap given the content of the file
func (r *ILPTaskReconciler) createConfigMap(ctx context.Context, name string, content string, ilptask *corev1alpha1.ILPTask) error {
	log := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: ilptask.Namespace}, cm)
	if err == nil {
		log.Info("ILPTask: ConfigMap already exists, not expected.")
		return nil
	}

	// Define a new ConfigMap object
	cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ilptask.Namespace,
			Annotations: map[string]string{
				SkyClusterAnnotationManagedBy:  "skycluster",
				SkyClusterAnnotationConfigType: "optimizer",
			},
		},
		Data: map[string]string{
			name: content,
		},
	}

	// Set MyResource instance as the owner and controller
	if err := controllerutil.SetControllerReference(ilptask, cm, r.Scheme); err != nil {
		log.Error(err, "ILPTask: Failed to set owner reference on ConfigMap")
		return err
	}

	// Check if this ConfigMap already exists
	found := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("ILPTask: Creating a new ConfigMap")
		err = r.Create(ctx, cm)
		if err != nil {
			log.Error(err, "ILPTask: Failed to create new ConfigMap")
			return err
		}
		// ConfigMap created successfully - return and requeue
		return nil
	}
	log.Info("ILPTask: Should not be here. CM should not exist at this point.")
	return nil
}

func (r *ILPTaskReconciler) getPodLogs(ctx context.Context, pod *corev1.Pod) (string, error) {
	req := r.Clientset.Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ILPTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.ILPTask{}).
		// Watches(
		// 	&corev1alpha1.DataflowAttribute{},
		// 	&handler.EnqueueRequestForObject{},
		// ).
		// Watches(
		// 	&corev1alpha1.SkyApp{},
		// 	&handler.EnqueueRequestForObject{},
		// ).
		Complete(r)
}
