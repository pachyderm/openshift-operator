//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021 Pachyderm.

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

package v1beta1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AmazonStorageOptions) DeepCopyInto(out *AmazonStorageOptions) {
	*out = *in
	if in.Reverse != nil {
		in, out := &in.Reverse, &out.Reverse
		*out = new(bool)
		**out = **in
	}
	if in.Vault != nil {
		in, out := &in.Vault, &out.Vault
		*out = new(AmazonStorageVault)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AmazonStorageOptions.
func (in *AmazonStorageOptions) DeepCopy() *AmazonStorageOptions {
	if in == nil {
		return nil
	}
	out := new(AmazonStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AmazonStorageVault) DeepCopyInto(out *AmazonStorageVault) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AmazonStorageVault.
func (in *AmazonStorageVault) DeepCopy() *AmazonStorageVault {
	if in == nil {
		return nil
	}
	out := new(AmazonStorageVault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackupOptions) DeepCopyInto(out *BackupOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackupOptions.
func (in *BackupOptions) DeepCopy() *BackupOptions {
	if in == nil {
		return nil
	}
	out := new(BackupOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsoleOptions) DeepCopyInto(out *ConsoleOptions) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(ImageOverride)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceOverrides)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsoleOptions.
func (in *ConsoleOptions) DeepCopy() *ConsoleOptions {
	if in == nil {
		return nil
	}
	out := new(ConsoleOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EtcdOptions) DeepCopyInto(out *EtcdOptions) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(ImageOverride)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceOverrides)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EtcdOptions.
func (in *EtcdOptions) DeepCopy() *EtcdOptions {
	if in == nil {
		return nil
	}
	out := new(EtcdOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GoogleStorageOptions) DeepCopyInto(out *GoogleStorageOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GoogleStorageOptions.
func (in *GoogleStorageOptions) DeepCopy() *GoogleStorageOptions {
	if in == nil {
		return nil
	}
	out := new(GoogleStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageOverride) DeepCopyInto(out *ImageOverride) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageOverride.
func (in *ImageOverride) DeepCopy() *ImageOverride {
	if in == nil {
		return nil
	}
	out := new(ImageOverride)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalStorageOptions) DeepCopyInto(out *LocalStorageOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalStorageOptions.
func (in *LocalStorageOptions) DeepCopy() *LocalStorageOptions {
	if in == nil {
		return nil
	}
	out := new(LocalStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricsOptions) DeepCopyInto(out *MetricsOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricsOptions.
func (in *MetricsOptions) DeepCopy() *MetricsOptions {
	if in == nil {
		return nil
	}
	out := new(MetricsOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicrosoftStorageOptions) DeepCopyInto(out *MicrosoftStorageOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicrosoftStorageOptions.
func (in *MicrosoftStorageOptions) DeepCopy() *MicrosoftStorageOptions {
	if in == nil {
		return nil
	}
	out := new(MicrosoftStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinioStorageOptions) DeepCopyInto(out *MinioStorageOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinioStorageOptions.
func (in *MinioStorageOptions) DeepCopy() *MinioStorageOptions {
	if in == nil {
		return nil
	}
	out := new(MinioStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectStorageOptions) DeepCopyInto(out *ObjectStorageOptions) {
	*out = *in
	if in.Amazon != nil {
		in, out := &in.Amazon, &out.Amazon
		*out = new(AmazonStorageOptions)
		(*in).DeepCopyInto(*out)
	}
	if in.Google != nil {
		in, out := &in.Google, &out.Google
		*out = new(GoogleStorageOptions)
		**out = **in
	}
	if in.Microsoft != nil {
		in, out := &in.Microsoft, &out.Microsoft
		*out = new(MicrosoftStorageOptions)
		**out = **in
	}
	if in.Minio != nil {
		in, out := &in.Minio, &out.Minio
		*out = new(MinioStorageOptions)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectStorageOptions.
func (in *ObjectStorageOptions) DeepCopy() *ObjectStorageOptions {
	if in == nil {
		return nil
	}
	out := new(ObjectStorageOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachdOptions) DeepCopyInto(out *PachdOptions) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	in.Storage.DeepCopyInto(&out.Storage)
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(ImageOverride)
		**out = **in
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceOverrides)
		(*in).DeepCopyInto(*out)
	}
	out.Metrics = in.Metrics
	out.Postgres = in.Postgres
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachdOptions.
func (in *PachdOptions) DeepCopy() *PachdOptions {
	if in == nil {
		return nil
	}
	out := new(PachdOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachdPostgresConfig) DeepCopyInto(out *PachdPostgresConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachdPostgresConfig.
func (in *PachdPostgresConfig) DeepCopy() *PachdPostgresConfig {
	if in == nil {
		return nil
	}
	out := new(PachdPostgresConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pachyderm) DeepCopyInto(out *Pachyderm) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pachyderm.
func (in *Pachyderm) DeepCopy() *Pachyderm {
	if in == nil {
		return nil
	}
	out := new(Pachyderm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Pachyderm) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermList) DeepCopyInto(out *PachydermList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Pachyderm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermList.
func (in *PachydermList) DeepCopy() *PachydermList {
	if in == nil {
		return nil
	}
	out := new(PachydermList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PachydermList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermRestore) DeepCopyInto(out *PachydermRestore) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermRestore.
func (in *PachydermRestore) DeepCopy() *PachydermRestore {
	if in == nil {
		return nil
	}
	out := new(PachydermRestore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermSpec) DeepCopyInto(out *PachydermSpec) {
	*out = *in
	in.Etcd.DeepCopyInto(&out.Etcd)
	in.Pachd.DeepCopyInto(&out.Pachd)
	in.Console.DeepCopyInto(&out.Console)
	in.Worker.DeepCopyInto(&out.Worker)
	in.Postgres.DeepCopyInto(&out.Postgres)
	if in.ImagePullSecret != nil {
		in, out := &in.ImagePullSecret, &out.ImagePullSecret
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermSpec.
func (in *PachydermSpec) DeepCopy() *PachydermSpec {
	if in == nil {
		return nil
	}
	out := new(PachydermSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermStatus) DeepCopyInto(out *PachydermStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermStatus.
func (in *PachydermStatus) DeepCopy() *PachydermStatus {
	if in == nil {
		return nil
	}
	out := new(PachydermStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermVault) DeepCopyInto(out *PachydermVault) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermVault.
func (in *PachydermVault) DeepCopy() *PachydermVault {
	if in == nil {
		return nil
	}
	out := new(PachydermVault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PachydermVault) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermVaultList) DeepCopyInto(out *PachydermVaultList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PachydermVault, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermVaultList.
func (in *PachydermVaultList) DeepCopy() *PachydermVaultList {
	if in == nil {
		return nil
	}
	out := new(PachydermVaultList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PachydermVaultList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermVaultSpec) DeepCopyInto(out *PachydermVaultSpec) {
	*out = *in
	if in.Backup != nil {
		in, out := &in.Backup, &out.Backup
		*out = new(BackupOptions)
		**out = **in
	}
	if in.Restore != nil {
		in, out := &in.Restore, &out.Restore
		*out = new(RestoreOptions)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermVaultSpec.
func (in *PachydermVaultSpec) DeepCopy() *PachydermVaultSpec {
	if in == nil {
		return nil
	}
	out := new(PachydermVaultSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PachydermVaultStatus) DeepCopyInto(out *PachydermVaultStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PachydermVaultStatus.
func (in *PachydermVaultStatus) DeepCopy() *PachydermVaultStatus {
	if in == nil {
		return nil
	}
	out := new(PachydermVaultStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PostgresOptions) DeepCopyInto(out *PostgresOptions) {
	*out = *in
	in.Service.DeepCopyInto(&out.Service)
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PostgresOptions.
func (in *PostgresOptions) DeepCopy() *PostgresOptions {
	if in == nil {
		return nil
	}
	out := new(PostgresOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestoreOptions) DeepCopyInto(out *RestoreOptions) {
	*out = *in
	out.Pachyderm = in.Pachyderm
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestoreOptions.
func (in *RestoreOptions) DeepCopy() *RestoreOptions {
	if in == nil {
		return nil
	}
	out := new(RestoreOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceOverrides) DeepCopyInto(out *ServiceOverrides) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceOverrides.
func (in *ServiceOverrides) DeepCopy() *ServiceOverrides {
	if in == nil {
		return nil
	}
	out := new(ServiceOverrides)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkerOptions) DeepCopyInto(out *WorkerOptions) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(ImageOverride)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkerOptions.
func (in *WorkerOptions) DeepCopy() *WorkerOptions {
	if in == nil {
		return nil
	}
	out := new(WorkerOptions)
	in.DeepCopyInto(out)
	return out
}
