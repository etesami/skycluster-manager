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

var pythonBaseScript string = `
import numpy as np
import collections
import pulp
class Task:
	def __init__(self, name):
		self.name = name
		self.vservices = []

	def add_vservice(self, vs):
		self.vservices.append(vs)
	
	def get_vservices(self):
		return self.vservices

	def contain_vservice(self, vs):
		return vs in self.vservices

class Dag:
	def __init__(self):
		self.tasks = []
		self.vservices = {}
		self.name = None
		import networkx as nx 
		self.graph = nx.DiGraph()

	def add(self, task):
		self.graph.add_node(task)
		self.tasks.append(task)

	def remove(self, task):
		self.tasks.remove(task)
		self.graph.remove_node(task)

	def add_edge(self, op1, op2):
		assert op1 in self.graph.nodes
		assert op2 in self.graph.nodes
		self.graph.add_edge(op1, op2)

	def get_graph(self):
		return self.graph

	def get_tasks(self):
		return self.tasks

	def get_edges(self):
		return self.graph.edges

class VService:
	def __init__(self, name, costs=None):
		self.name = name
		self.costs = {}
		if costs is not None:
			for p in costs:
				self.costs[p] = costs[p]
						
	def set_costs(self, costs):
		for p in costs:
			self.costs[p] = costs[p]
					
	def get_costs(self):
		return self.costs

class Provider:
	def __init__(self, name):
		self.name = name
		self.region = None
		self.zone = None

	def set_region(self, region):
		self.region = region
	
	def set_zone(self, zone):
		self.zone = zone

`

var pythonOptimizationScript string = `

prob = pulp.LpProblem('cost_optimization', pulp.LpMinimize)
        
# Prepare the constants.
V = dag.get_tasks()
E = dag.graph.edges()  

# Define the decision variables.
c = {
	v.name: pulp.LpVariable.matrix(v.name, range(len(providers)), cat='Binary') for v in dag.get_tasks()
}

# Formulate the constraints.
# 1. c[v] 
for v in V:
	prob += pulp.lpSum(c[v.name]) <= len(providers)
	prob += pulp.lpSum(c[v.name]) >= 1

# special constraints
# TODO: automatically add constraints
# c[v1][p0] = 1

# TODO: automatically add constraints
for v in V:
	if v.name == 'frontend':
		for ii, pp in enumerate(providers):
			if pp == 'aws-east1':
				prob += c[v.name][ii] == 1

objective = 0
for v in V:
	total_costs = {}
	for vs in v.get_vservices():
		# print(vs.name, vs.get_costs())
		for p in vs.get_costs():
			# p does not exist
			if p not in total_costs:
				total_costs[p] = float(vs.get_costs()[p])
			else: # p exists
				total_costs[p] += float(vs.get_costs()[p])
	objective += pulp.lpDot(c[v.name], list(total_costs.values()))

e = collections.defaultdict(lambda: collections.defaultdict(lambda: collections.defaultdict(dict)))
for u, v in E:
	for ii in range(0, len(providers)-1):
		for jj in range(ii+1, len(providers)):
			e[u.name][v.name][str(ii)+'_'+str(jj)] = pulp.LpVariable(
				u.name + v.name + str(ii)+'_'+str(jj) , cat='Binary')

# 2. e[u][v] 
for u, v in E:
	for ii in range(0, len(providers)-1):
		for jj in range(ii+1, len(providers)):
			prob += e[u.name][v.name][str(ii)+'_'+str(jj)] <= c[u.name][ii]
			prob += e[u.name][v.name][str(ii)+'_'+str(jj)] <= c[v.name][jj]
			prob += e[u.name][v.name][str(ii)+'_'+str(jj)] >= c[v.name][ii] + c[v.name][jj] - 1
			prob += e[u.name][v.name][str(ii)+'_'+str(jj)] >= 0

# Construct F
# C'ij=Cij+Cji
pp = list(egress_cost_dict.keys())
for u, v in E:
	for ii in range(0, len(providers)-1):
		for jj in range(ii+1, len(providers)):
			objective += e[u.name][v.name][str(ii)+'_'+str(jj)] * (float(egress_cost_dict[pp[ii]][pp[jj]]) + float(egress_cost_dict[pp[jj]][pp[ii]]))

prob += objective

# Last Step: Solve the problem
solver = pulp.PULP_CBC_CMD(msg=0)
prob.solve(solver)

# Step 8: Print the results
print("Status:", pulp.LpStatus[prob.status])

for v in V:
	print(v.name, [l.varValue for l in c[v.name]])
`

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
			taskNames := ""
			for i, thisTask := range skyapp.Spec.AppConfig {
				taskNames += thisTask.Name + ":"
				for j, thisVService := range thisTask.Constraints.VirtualServiceConstraints {
					taskNames += thisVService.VirtualServiceName
					if j < len(thisTask.Constraints.VirtualServiceConstraints)-1 {
						taskNames += ","
					}
				}
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

			// get providers data
			providers := &corev1alpha1.ProviderList{}
			listOps := &client.ListOptions{
				Namespace: ilptask.Namespace,
			}
			if err := r.List(ctx, providers, listOps); err != nil {
				log.Error(err, "Unable to get providers")
				return ctrl.Result{}, err
			} else {
				// iterate over the providers and create a configmap for each
				providersData := ""
				for i, thisProvider := range providers.Items {
					providersData += thisProvider.Spec.Name + ","
					providersData += thisProvider.Spec.Region + ","
					providersData += thisProvider.Spec.Zone
					if i < len(providers.Items)-1 {
						providersData += "\n"
					}
				}
				if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-providers", providersData, ilptask); err != nil {
					log.Error(err, "Unable to create ConfigMap for providers")
					return ctrl.Result{}, err
				}
			}

			// Fetch the ProviderAttribute instance
			// Assuming there is one instance of this type (part of TODO list)
			providerAttrs := &corev1alpha1.ProviderAttributeList{}
			listOps = &client.ListOptions{
				Namespace: ilptask.Namespace,
			}
			if err := r.List(ctx, providerAttrs, listOps); err != nil {
				log.Error(err, "Unable to get providerAttrs")
				return ctrl.Result{}, err
			}
			// Assume only one providerAttribute exists:
			providerAttr := &providerAttrs.Items[0]
			providerAttrData := ""
			for i, currentLink := range providerAttr.Spec.ProviderMetrics {
				for j, dest := range currentLink.DstProviderMetrics {
					providerAttrData += currentLink.SrcProviderName + ":"
					providerAttrData += dest.DstProviderName
					providerAttrData += "," + dest.Latency
					providerAttrData += "," + dest.EgressDataCost
					if j < len(currentLink.DstProviderMetrics)-1 {
						providerAttrData += "\n"
					}
				}
				if i < len(providerAttr.Spec.ProviderMetrics)-1 {
					providerAttrData += "\n"
				}
			}
			if err = r.createConfigMap(ctx, ilptask.Spec.AppName+"-providerattr", providerAttrData, ilptask); err != nil {
				log.Error(err, "Unable to create ConfigMap for providerAttr")
				return ctrl.Result{}, err
			}

			// get virtual service data
			vservices := &corev1alpha1.VirtualServiceList{}
			listOps = &client.ListOptions{
				Namespace: ilptask.Namespace,
			}
			if err := r.List(ctx, vservices, listOps); err != nil {
				log.Error(err, "Unable to get virtual services")
				return ctrl.Result{}, err
			} else {
				// iterate over the vservices and create a configmap for each
				vservicesData := ""
				for i, thisProvider := range vservices.Items {
					for j, thisVServiceProvider := range thisProvider.Spec.VServiceCosts {
						vservicesData += thisProvider.Spec.Name + ":"
						vservicesData += thisVServiceProvider.ProviderName
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

			// construct the command
			command := `
__PYTHON_BASE_SCRIPT__

providers = {}
filename = '/providers/__FILE_NAME_PROVIDERS__'
with open(filename, 'r') as file:
	for line in file:
		provider_name = line.strip().split(",")[0] + '-' + line.strip().split(",")[1]
		pp = Provider(provider_name)
		pp.set_region(line.strip().split(",")[1])
		pp.set_zone(line.strip().split(",")[2])
		providers[provider_name] = pp
print('providers [okay]')

vservices = {}
vs_costs = {}
filename = '/vservices/__FILE_NAME_VSERVICES__'
with open(filename, 'r') as file:
	for line in file:
		vservice_name = line.strip().split(":")[0]
		vservice_provider = line.strip().split(":")[1].split(",")[0]
		vservice_cost = line.strip().split(":")[1].split(",")[1]
		if vservice_name not in vservices:
			vservices[vservice_name] = VService(vservice_name)
		vs_costs[vservice_provider] = vservice_cost
		vservices[vservice_name].set_costs(vs_costs)
print('vservices [okay]')
print('\n'+'-'*5)

filename = '/tasks/__FILE_NAME_TASKS__'
dag = Dag()
tasks = {}
with open(filename, 'r') as file:
	for line in file:
		tt = line.strip().split(":")[0]
		tt_vservices = line.strip().split(":")[1].split(",")
		task = Task(tt)
		tasks[tt] = task
		for vs in tt_vservices:
			task.add_vservice(vservices[vs])
		dag.add(task)
print('tasks [okay]')
print('\n'+'-'*5)

filename = '/edges/__FILE_NAME_EDGES__'
with open(filename, 'r') as file:
	for line in file:
		u = line.strip().split(":")[0]
		v = line.strip().split(":")[1].split(',')[0]
		dag.add_edge(tasks[u],tasks[v])
print('edges [okay]')
print('\n'+'-'*5)

egress_cost_dict = {}
for pp in providers:
	egress_cost_dict[pp] = {}
	for dd in providers:
		egress_cost_dict[pp][dd] = 0
filename = '/providerattr/__FILE_NAME_PROVIDERATTR__'
with open(filename, 'r') as file:
	for line in file:
		pp = line.strip().split(':')[0]
		dd = line.strip().split(':')[1].split(',')[0]
		cc = line.strip().split(':')[1].split(',')[2]
		egress_cost_dict[pp][dd] = cc
print('providerAttr [okay]')
print('\n'+'-'*5)

# k[vs][provider] = cost
k = {}
for vs in vservices.values():
	k[vs.name] = vs.get_costs()
print('k [okay]')
print('\n'+'-'*5)

__PYTHON_OPTIMIZATION_SCRIPT__

`
			replacedCommand := strings.ReplaceAll(command, "__FILE_NAME_TASKS__", ilptask.Spec.AppName+"-tasks")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_PROVIDERS__", ilptask.Spec.AppName+"-providers")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_VSERVICES__", ilptask.Spec.AppName+"-vservices")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_EDGES__", ilptask.Spec.AppName+"-edges")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__FILE_NAME_PROVIDERATTR__", ilptask.Spec.AppName+"-providerattr")
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_BASE_SCRIPT__", pythonBaseScript)
			replacedCommand = strings.ReplaceAll(replacedCommand, "__PYTHON_OPTIMIZATION_SCRIPT__", pythonOptimizationScript)

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
