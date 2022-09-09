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

// RestoreDestination name of pachyderm instance to restore to
type RestoreDestination struct {
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Restore Target",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Name string `json:"name,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pachyderm Namespace",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Namespace string `json:"namespace,omitempty"`
}

// PachydermImportSpec defines the desired state of PachydermImport
type PachydermImportSpec struct {
	// Name of the pachyderm instance to restore the backup to
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Restore Destination",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Destination RestoreDestination `json:"destination,omitempty"`
	// Name of backup resource in S3 to restore
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Name",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	BackupName string `json:"backup,omitempty"`
	// Storage Secret containing credentials to
	// upload the backup to an S3-compatible object store
	//+kubebuilder:validation:required=true
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="S3 Upload Secret",xDescriptors={"urn:alm:descriptor:io.kubernetes:Secret"}
	StorageSecret string `json:"storageSecret,omitempty"`
}

// PachydermImportStatus defines the observed state of PachydermImport
type PachydermImportStatus struct {
	// Phase reports the status of the restore
	Phase string `json:"phase,omitempty"`
	// Unique ID of the backup
	ID string `json:"id,omitempty"`
	// Time the restore process commenced
	StartedAt string `json:"startedAt,omitempty"`
	// Time the restore process completed
	CompletedAt string `json:"completedAt,omitempty"`
	// Status reports the state of the restore request
	Status string `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PachydermImport is the Schema for the pachydermimports API
type PachydermImport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PachydermImportSpec   `json:"spec,omitempty"`
	Status PachydermImportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PachydermImportList contains a list of PachydermImport
type PachydermImportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PachydermImport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PachydermImport{}, &PachydermImportList{})
}
