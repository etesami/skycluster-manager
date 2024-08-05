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
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=skyapp,verbs=create;update;patch;delete
// +kubebuilder:rbac:groups=core.skycluster-manager.savitestbed.ca,resources=dataflowattribute,verbs=create;update;patch;delete

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
	log.Info("Reconciling ILPTask ["+req.Name+"]", "name", req.Name, "namespace", req.Namespace)

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

	// Check if the optimization is already completed
	// if it is completed, the Status is not nil
	if ilptask.Status.Result != "" {
		log.Info("ILPTask already completed or has a result")
		return ctrl.Result{}, nil
	}

	// Define the Pod name
	podName := fmt.Sprintf("%s-ilptaskr", ilptask.Name)

	// Check if the Pod exists
	pod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: podName, Namespace: ilptask.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Pod doesn't exist, create it
			log.Info("Creating optimizer Pod", "name", podName)
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
								"import time; print('Optimizer running...'); time.sleep(5); print('Optimizer completed')",
							},
						},
					},
				},
			}
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "Unable to create optimizer Pod")
				return ctrl.Result{}, err
			}
			// Requeue to check the Pod status later
			log.Info("Requeue to check optimizer Pod status")
			return ctrl.Result{RequeueAfter: time.Second * 6}, nil
		}
		log.Error(err, "Unable to get optimizer Pod")
		return ctrl.Result{}, err
	}

	// Check Pod status
	if pod.Status.Phase == corev1.PodSucceeded {
		log.Info("Optimizer Pod exists, checking status...")
		// Pod completed successfully, get its logs
		// logs, err := r.getPodLogs(ctx, pod)
		// if err != nil {
		// 	log.Error(err, "Unable to get Pod logs")
		// 	return ctrl.Result{}, err
		// }
		// logs := pod.Spec.NodeName
		log.Info("Optimizer Pod completed")

		// Update the ILPTask status
		ilptask.Status.Result = "Completed"
		ilptask.Status.Solution, err = r.getPodLogs(ctx, pod)
		err = r.Status().Update(ctx, ilptask)
		if err != nil {
			log.Error(err, "Unable to update ILPTask status")
			return ctrl.Result{}, err
		}

		// Delete the completed Pod
		err = r.Delete(ctx, pod)
		log.Info("Deleting completed Pod", "name", pod.Name)
		if err != nil {
			log.Error(err, "Unable to delete completed Pod")
			// Don't return an error, as the main task is done
		}

		return ctrl.Result{}, nil
	} else if pod.Status.Phase == corev1.PodFailed {
		// Pod failed, update the ILPTask status with the failure
		log.Info("Optimizer Pod failed")
		ilptask.Status.Result = "Failed"
		ilptask.Status.Solution, err = r.getPodLogs(ctx, pod)
		err = r.Status().Update(ctx, ilptask)
		if err != nil {
			log.Error(err, "Unable to update ILPTask status")
			return ctrl.Result{}, err
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
		Watches(
			&corev1alpha1.DataflowAttribute{},
			&handler.EnqueueRequestForObject{},
		).
		Watches(
			&corev1alpha1.SkyApp{},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}
