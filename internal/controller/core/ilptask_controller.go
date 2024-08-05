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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ILPTask object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ILPTaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// print the name and namespace of the ILPTask
	log.Info("ILPTask [" + req.Name + "] Reconciler started")

	// // Fetch the DataflowAttribute instance
	// dataflowattribute := &corev1alpha1.DataflowAttribute{}
	// err = r.Get(ctx, req.NamespacedName, dataflowattribute)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		return ctrl.Result{}, nil
	// 	}
	// 	log.Error(err, "Unable to fetch DataflowAttribute, something is wrong.")
	// 	return ctrl.Result{}, err
	// }

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

	// Check if both SkyApp and DataflowAttribute references are set
	if ilptask.Spec.DataflowAttributeRef == (corev1alpha1.DataflowAttributeRef{}) ||
		ilptask.Spec.SkyAppRef == (corev1alpha1.SkyAppRef{}) {
		log.Info("ILPTask [" + req.Name + "] SkyApp or DataflowAttribute references are not set")
		return ctrl.Result{}, nil
	} else {
		log.Info("ILPTask [" + req.Name + "] SkyApp and DataflowAttribute references are set")
	}

	// Fetch the SkyApp instance
	// SkyApp may or may not exist
	skyapp := &corev1alpha1.SkyApp{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      ilptask.Spec.SkyAppRef.Name,
		Namespace: ilptask.Spec.SkyAppRef.Namespace,
	}, skyapp)
	if err == nil {
		log.Info("ILPTask [" + req.Name + "] SkyApp exists and was retrived")
	}

	// Logic to run the optimizer
	// Check if the optimization is already completed
	// if not, check if any pod is running
	// if not, create a pod to run the optimizer

	// Check if the optimization is already completed
	// if it is completed, the Status is not nil
	if ilptask.Status.Result != "" {
		log.Info("ILPTask [" + req.Name + "] task already completed or has a result")
		return ctrl.Result{}, nil
	}

	// Define the Pod name
	// podName := fmt.Sprintf("%s-ilptaskr", ilptask.Name)
	podName := ilptask.Spec.AppName

	// Check if the Pod exists
	pod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: ilptask.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Pod doesn't exist, create it
			log.Info("ILPTask [" + req.Name + "] Creating optimizer Pod")

			// Build application graph from SkyApp and DataflowAttribute
			nodeNames := ""
			for i, thisNode := range skyapp.Spec.AppConfig {
				nodeNames += thisNode.Name
				if i < len(skyapp.Spec.AppConfig)-1 {
					nodeNames += ","
				}
			}
			if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-nodes", nodeNames, ilptask); err != nil {
				log.Error(err, "Unable to create ConfigMap for nodes")
				return ctrl.Result{}, err
			}

			// construct the command
			command := `filename = '/scripts/__FILE_NAME__'
with open(filename, 'r') as file:
	for line in file:
		print(line, end='')`

			replacedCommand := strings.ReplaceAll(command, "__FILE_NAME__", ilptask.Spec.AppName+"-nodes")

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
							Image: "python:3.9",
							Command: []string{
								"python",
								"-c",
								replacedCommand,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: "nodes-cm",
									// This is the mount point inside the container
									MountPath: "/scripts",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nodes-cm",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										// This is the name of configMap
										Name: ilptask.Spec.AppName + "-nodes",
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
		log.Error(err, "Unable to get optimizer Pod")
		return ctrl.Result{}, err
	}

	// if Pod exists, check its status
	if pod.Status.Phase == corev1.PodSucceeded {
		log.Info("ILPTask [" + req.Name + "] Optimizer Pod exists, checking status...")
		// Pod completed successfully, get its logs

		// Update the ILPTask status
		ilptask.Status.Result = "Completed"
		if ilptask.Status.Solution, err = r.getPodLogs(ctx, pod); err != nil {
			log.Error(err, "Unable to get Pod logs")
			// return ctrl.Result{}, err
		}
		err = r.Status().Update(ctx, ilptask)
		if err != nil {
			log.Error(err, "Unable to update ILPTask status")
			return ctrl.Result{}, err
		}
		log.Info("ILPTask [" + req.Name + "] Optimizer Pod completed")

		// Delete the completed Pod
		err = r.Delete(ctx, pod)
		log.Info("ILPTask ["+req.Name+"] Deleting completed Pod", "name", pod.Name)
		if err != nil {
			log.Error(err, "Unable to delete completed Pod")
			// Don't return an error, as the main task is done
		}
		return ctrl.Result{}, nil

	} else if pod.Status.Phase == corev1.PodFailed {
		// Pod failed, update the ILPTask status with the failure
		log.Info("ILPTask [" + req.Name + "] Optimizer Pod failed")
		ilptask.Status.Result = "Failed"
		if ilptask.Status.Solution, err = r.getPodLogs(ctx, pod); err != nil {
			log.Error(err, "Unable to get Pod logs")
		}
		err = r.Status().Update(ctx, ilptask)
		if err != nil {
			log.Error(err, "Unable to update ILPTask status")
			// return ctrl.Result{}, err
		}
		// Delete the failed Pod
		log.Info("Deleting failed Pod", "name", pod.Name)
		err = r.Delete(ctx, pod)
		if err != nil {
			log.Error(err, "Unable to delete failed Pod")
			// Don't return an error, as the main task is done
		}
		return ctrl.Result{}, nil
	}

	// Pod is still running, requeue
	return ctrl.Result{RequeueAfter: time.Second * 6}, nil
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
