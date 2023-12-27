//go:build !ignore_autogenerated

/*
Copyright  The Kubernetes Authors.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterReferenceConsumer) DeepCopyInto(out *ClusterReferenceConsumer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Subject = in.Subject
	if in.PatternNames != nil {
		in, out := &in.PatternNames, &out.PatternNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterReferenceConsumer.
func (in *ClusterReferenceConsumer) DeepCopy() *ClusterReferenceConsumer {
	if in == nil {
		return nil
	}
	out := new(ClusterReferenceConsumer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterReferenceConsumer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterReferencePattern) DeepCopyInto(out *ClusterReferencePattern) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Paths != nil {
		in, out := &in.Paths, &out.Paths
		*out = make([]VersionedPath, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterReferencePattern.
func (in *ClusterReferencePattern) DeepCopy() *ClusterReferencePattern {
	if in == nil {
		return nil
	}
	out := new(ClusterReferencePattern)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterReferencePattern) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReferenceGrant) DeepCopyInto(out *ReferenceGrant) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = make([]ReferenceGrantFrom, len(*in))
		copy(*out, *in)
	}
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = make([]ReferenceGrantTo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReferenceGrant.
func (in *ReferenceGrant) DeepCopy() *ReferenceGrant {
	if in == nil {
		return nil
	}
	out := new(ReferenceGrant)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReferenceGrant) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReferenceGrantFrom) DeepCopyInto(out *ReferenceGrantFrom) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReferenceGrantFrom.
func (in *ReferenceGrantFrom) DeepCopy() *ReferenceGrantFrom {
	if in == nil {
		return nil
	}
	out := new(ReferenceGrantFrom)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReferenceGrantTo) DeepCopyInto(out *ReferenceGrantTo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReferenceGrantTo.
func (in *ReferenceGrantTo) DeepCopy() *ReferenceGrantTo {
	if in == nil {
		return nil
	}
	out := new(ReferenceGrantTo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionedPath) DeepCopyInto(out *VersionedPath) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionedPath.
func (in *VersionedPath) DeepCopy() *VersionedPath {
	if in == nil {
		return nil
	}
	out := new(VersionedPath)
	in.DeepCopyInto(out)
	return out
}