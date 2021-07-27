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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PachydermVaultSpec defines the desired state of PachydermVault
type PachydermVaultSpec struct {
	// Backup options allow the user to provide options
	// when performing a backup
	Backup *BackupOptions `json:"backup,omitempty"`

	// Restore allows a user to restore a backup to a new Pachyderm cluster
	Restore *RestoreOptions `json:"restore,omitempty"`
}

type BackupOptions struct {
	// Name of Pachyderm instance to backup.
	Pachyderm string `json:"pachyderm"`
}

type RestoreOptions struct {
	// Name of the pachyderm instance to be
	// deployed from a specific backup
	Pachyderm PachydermRestore `json:"pachyderm"`
	// Name of backup to restore
	BackupName string `json:"backup,omitempty"`
}

type PachydermRestore struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// PachydermVaultStatus defines the observed state of PachydermVault
type PachydermVaultStatus struct {
	// Time the backup process commenced
	StartedAt string `json:"startedAt,omitempty"`
	// Time the backup process completed
	CompletedAt string `json:"completedAt,omitempty"`
	// Name and location of backup resource created
	Backup string `json:"backupName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PachydermVault is the Schema for the pachydermvaults API
type PachydermVault struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PachydermVaultSpec   `json:"spec,omitempty"`
	Status PachydermVaultStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PachydermVaultList contains a list of PachydermVault
type PachydermVaultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PachydermVault `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PachydermVault{}, &PachydermVaultList{})
}
