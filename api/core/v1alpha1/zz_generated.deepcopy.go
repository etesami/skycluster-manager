//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeployMap) DeepCopyInto(out *DeployMap) {
	*out = *in
	if in.Component != nil {
		in, out := &in.Component, &out.Component
		*out = make([]SkyComponent, len(*in))
		copy(*out, *in)
	}
	if in.Edges != nil {
		in, out := &in.Edges, &out.Edges
		*out = make([]SkyComponent, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeployMap.
func (in *DeployMap) DeepCopy() *DeployMap {
	if in == nil {
		return nil
	}
	out := new(DeployMap)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Deployment) DeepCopyInto(out *Deployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Deployment.
func (in *Deployment) DeepCopy() *Deployment {
	if in == nil {
		return nil
	}
	out := new(Deployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Deployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentList) DeepCopyInto(out *DeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Deployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentList.
func (in *DeploymentList) DeepCopy() *DeploymentList {
	if in == nil {
		return nil
	}
	out := new(DeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentSpec) DeepCopyInto(out *DeploymentSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentSpec.
func (in *DeploymentSpec) DeepCopy() *DeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(DeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentStatus) DeepCopyInto(out *DeploymentStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentStatus.
func (in *DeploymentStatus) DeepCopy() *DeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(DeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ILPTask) DeepCopyInto(out *ILPTask) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ILPTask.
func (in *ILPTask) DeepCopy() *ILPTask {
	if in == nil {
		return nil
	}
	out := new(ILPTask)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ILPTask) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ILPTaskList) DeepCopyInto(out *ILPTaskList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ILPTask, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ILPTaskList.
func (in *ILPTaskList) DeepCopy() *ILPTaskList {
	if in == nil {
		return nil
	}
	out := new(ILPTaskList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ILPTaskList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ILPTaskSpec) DeepCopyInto(out *ILPTaskSpec) {
	*out = *in
	if in.SkyComponents != nil {
		in, out := &in.SkyComponents, &out.SkyComponents
		*out = make([]SkyComponent, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ILPTaskSpec.
func (in *ILPTaskSpec) DeepCopy() *ILPTaskSpec {
	if in == nil {
		return nil
	}
	out := new(ILPTaskSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ILPTaskStatus) DeepCopyInto(out *ILPTaskStatus) {
	*out = *in
	in.Optimization.DeepCopyInto(&out.Optimization)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ILPTaskStatus.
func (in *ILPTaskStatus) DeepCopy() *ILPTaskStatus {
	if in == nil {
		return nil
	}
	out := new(ILPTaskStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricSpec) DeepCopyInto(out *MetricSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricSpec.
func (in *MetricSpec) DeepCopy() *MetricSpec {
	if in == nil {
		return nil
	}
	out := new(MetricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitoringMetricSpec) DeepCopyInto(out *MonitoringMetricSpec) {
	*out = *in
	out.Resource = in.Resource
	out.Metric = in.Metric
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitoringMetricSpec.
func (in *MonitoringMetricSpec) DeepCopy() *MonitoringMetricSpec {
	if in == nil {
		return nil
	}
	out := new(MonitoringMetricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OptimizationSpec) DeepCopyInto(out *OptimizationSpec) {
	*out = *in
	in.DeployMap.DeepCopyInto(&out.DeployMap)
	out.ConfigMapRef = in.ConfigMapRef
	out.PodRef = in.PodRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OptimizationSpec.
func (in *OptimizationSpec) DeepCopy() *OptimizationSpec {
	if in == nil {
		return nil
	}
	out := new(OptimizationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProviderRefSpec) DeepCopyInto(out *ProviderRefSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProviderRefSpec.
func (in *ProviderRefSpec) DeepCopy() *ProviderRefSpec {
	if in == nil {
		return nil
	}
	out := new(ProviderRefSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyCluster) DeepCopyInto(out *SkyCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyCluster.
func (in *SkyCluster) DeepCopy() *SkyCluster {
	if in == nil {
		return nil
	}
	out := new(SkyCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyClusterList) DeepCopyInto(out *SkyClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SkyCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyClusterList.
func (in *SkyClusterList) DeepCopy() *SkyClusterList {
	if in == nil {
		return nil
	}
	out := new(SkyClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyClusterSpec) DeepCopyInto(out *SkyClusterSpec) {
	*out = *in
	out.DataflowPolicyRef = in.DataflowPolicyRef
	out.DeploymentPolciyRef = in.DeploymentPolciyRef
	if in.SkyProviders != nil {
		in, out := &in.SkyProviders, &out.SkyProviders
		*out = make([]ProviderRefSpec, len(*in))
		copy(*out, *in)
	}
	if in.SkyComponents != nil {
		in, out := &in.SkyComponents, &out.SkyComponents
		*out = make([]SkyComponent, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyClusterSpec.
func (in *SkyClusterSpec) DeepCopy() *SkyClusterSpec {
	if in == nil {
		return nil
	}
	out := new(SkyClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyClusterStatus) DeepCopyInto(out *SkyClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyClusterStatus.
func (in *SkyClusterStatus) DeepCopy() *SkyClusterStatus {
	if in == nil {
		return nil
	}
	out := new(SkyClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyComponent) DeepCopyInto(out *SkyComponent) {
	*out = *in
	out.Component = in.Component
	out.Provider = in.Provider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyComponent.
func (in *SkyComponent) DeepCopy() *SkyComponent {
	if in == nil {
		return nil
	}
	out := new(SkyComponent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyProvider) DeepCopyInto(out *SkyProvider) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyProvider.
func (in *SkyProvider) DeepCopy() *SkyProvider {
	if in == nil {
		return nil
	}
	out := new(SkyProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyProvider) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyProviderList) DeepCopyInto(out *SkyProviderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SkyProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyProviderList.
func (in *SkyProviderList) DeepCopy() *SkyProviderList {
	if in == nil {
		return nil
	}
	out := new(SkyProviderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyProviderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyProviderSpec) DeepCopyInto(out *SkyProviderSpec) {
	*out = *in
	out.ProviderRef = in.ProviderRef
	if in.DependsOn != nil {
		in, out := &in.DependsOn, &out.DependsOn
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.DependedBy != nil {
		in, out := &in.DependedBy, &out.DependedBy
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyProviderSpec.
func (in *SkyProviderSpec) DeepCopy() *SkyProviderSpec {
	if in == nil {
		return nil
	}
	out := new(SkyProviderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyProviderStatus) DeepCopyInto(out *SkyProviderStatus) {
	*out = *in
	in.Conditions.DeepCopyInto(&out.Conditions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyProviderStatus.
func (in *SkyProviderStatus) DeepCopy() *SkyProviderStatus {
	if in == nil {
		return nil
	}
	out := new(SkyProviderStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyVM) DeepCopyInto(out *SkyVM) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyVM.
func (in *SkyVM) DeepCopy() *SkyVM {
	if in == nil {
		return nil
	}
	out := new(SkyVM)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyVM) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyVMList) DeepCopyInto(out *SkyVMList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SkyVM, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyVMList.
func (in *SkyVMList) DeepCopy() *SkyVMList {
	if in == nil {
		return nil
	}
	out := new(SkyVMList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SkyVMList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyVMSpec) DeepCopyInto(out *SkyVMSpec) {
	*out = *in
	if in.MonitoringMetricList != nil {
		in, out := &in.MonitoringMetricList, &out.MonitoringMetricList
		*out = make([]MonitoringMetricSpec, len(*in))
		copy(*out, *in)
	}
	out.ProviderRef = in.ProviderRef
	if in.DependsOn != nil {
		in, out := &in.DependsOn, &out.DependsOn
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.DependedBy != nil {
		in, out := &in.DependedBy, &out.DependedBy
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyVMSpec.
func (in *SkyVMSpec) DeepCopy() *SkyVMSpec {
	if in == nil {
		return nil
	}
	out := new(SkyVMSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkyVMStatus) DeepCopyInto(out *SkyVMStatus) {
	*out = *in
	in.Conditions.DeepCopyInto(&out.Conditions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyVMStatus.
func (in *SkyVMStatus) DeepCopy() *SkyVMStatus {
	if in == nil {
		return nil
	}
	out := new(SkyVMStatus)
	in.DeepCopyInto(out)
	return out
}
