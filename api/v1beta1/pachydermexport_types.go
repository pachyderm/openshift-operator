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

// PachydermExportSpec defines the desired state of PachydermExport
type PachydermExportSpec struct {
	// Backup options allow the user to provide options
	// when performing a backup
	Backup *BackupOptions `json:"backup,omitempty"`

	// Restore allows a user to restore a backup to a new Pachyderm cluster
	Restore *RestoreOptions `json:"restore,omitempty"`
	// Storage Secret containing credentials to
	// upload the backup to an S3-compatible object store
	//+kubebuilder:validation:required=true
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="S3 Upload Secret",xDescriptors={"urn:alm:descriptor:io.kubernetes:Secret"}
	StorageSecret string `json:"storageSecret,omitempty"`
}

// BackupOptions exposes values to
type BackupOptions struct {
	// Name of Pachyderm instance to backup.
	Target string `json:"target"`
}

type RestoreOptions struct {
	// Name of the pachyderm instance to restore the backup to
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Restore Destination",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Destination RestoreDestination `json:"destination"`
	// Name of backup to restore
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Name",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	BackupName string `json:"backup,omitempty"`
}

// RestoreDestination name of pachyderm instance to restore to
type RestoreDestination struct {
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Restore Target",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Name string `json:"name,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pachyderm Namespace",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Namespace string `json:"namespace,omitempty"`
}

// PachydermExportStatus defines the observed state of PachydermExport
type PachydermExportStatus struct {
	// Time the backup process commenced
	StartedAt string `json:"startedAt,omitempty"`
	// Time the backup process completed
	CompletedAt string `json:"completedAt,omitempty"`
	// Name and location of backup resource created
	Backup string `json:"backupName,omitempty"`
	// ID of the backup
	BackupID string `json:"backupID,omitempty"`
	// Phase of the export status
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PachydermExport is the Schema for the pachydermexports API
type PachydermExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PachydermExportSpec   `json:"spec,omitempty"`
	Status PachydermExportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PachydermExportList contains a list of PachydermExport
type PachydermExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PachydermExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PachydermExport{}, &PachydermExportList{})
}
