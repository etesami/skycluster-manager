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
func (in *PortSpec) DeepCopyInto(out *PortSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortSpec.
func (in *PortSpec) DeepCopy() *PortSpec {
	if in == nil {
		return nil
	}
	out := new(PortSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecGroup) DeepCopyInto(out *SecGroup) {
	*out = *in
	if in.TCPPorts != nil {
		in, out := &in.TCPPorts, &out.TCPPorts
		*out = make([]PortSpec, len(*in))
		copy(*out, *in)
	}
	if in.UDPPorts != nil {
		in, out := &in.UDPPorts, &out.UDPPorts
		*out = make([]PortSpec, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecGroup.
func (in *SecGroup) DeepCopy() *SecGroup {
	if in == nil {
		return nil
	}
	out := new(SecGroup)
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
	if in.SecGroup != nil {
		in, out := &in.SecGroup, &out.SecGroup
		*out = make([]SecGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.ProviderRef = in.ProviderRef
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkyVMStatus.
func (in *SkyVMStatus) DeepCopy() *SkyVMStatus {
	if in == nil {
		return nil
	}
	out := new(SkyVMStatus)
	in.DeepCopyInto(out)
	return out
}
