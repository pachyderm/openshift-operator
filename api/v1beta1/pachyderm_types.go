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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PachydermSpec defines the desired state of Pachyderm
type PachydermSpec struct {
	// Allows the user to customize the etcd key-value store
	Etcd EtcdOptions `json:"etcd,omitempty"`
	// Allows the user to customize the pachd instance(s)
	Pachd PachdOptions `json:"pachd,omitempty"`
	// Allows the user to customize the dashd instance(s)
	Dashd  DashOptions    `json:"dash,omitempty"`
	Worker *WorkerOptions `json:"worker,omitempty"`
}

// WorkerOptions allows the user to configure workers
type WorkerOptions struct {
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image              *ImageOverride `json:"image,omitempty"`
	ServiceAccountName string         `json:"serviceAccountName,omitempty"`
}

// DashOptions provides options to configure the dashd component
type DashOptions struct {
	// Option to disable dash
	// Default: true
	Enabled bool `json:"enabled,omitempty" default:"true"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// Optional resource requirements required to run the dash pods.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// The address to use as the host in the dash ingress.
	// Used as the host of a rule
	URL     string            `json:"url,omitempty"`
	Service *ServiceOverrides `json:"service,omitempty"`
}

// ImageOverride allows the user to override the default image
// and change the image pull policy
type ImageOverride struct {
	// This option dictates the particular image to pull
	Repository string `json:"repository,omitempty"`
	// Used with the image registry to choose a specific
	// image in a cointainer registry to pull
	ImageTag string `json:"tag,omitempty"`
	// Determines when images should be pulled
	// Either, "IfNotPresent" or "Always"
	PullPolicy string `json:"pullPolicy,omitempty"`
}

type ServiceOverrides struct {
	Annotations []string `json:"annotations,omitempty"`
	Type        string   `json:"type"`
}

// EtcdOptions allows users to change the etcd statefulset
// TODO: potentially remove StorageProvider
type EtcdOptions struct {
	// Optional parameter to set the number of nodes in the Etcd statefulset.
	// Analogous --dynamic-etcd-nodes argument to 'pachctl deploy'
	DynamicNodes int32 `json:"dynamicNodes,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// Resource requests and limits for Etcd
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// If specified, etcd would use an existing storage class for its storage
	// Name of existing storage class to use for the Etcd persistent volume.
	StorageClass string `json:"storageClass,omitempty"`
	// The size of the storage to use for etcd.
	// For example: "100Gi"
	StorageSize string            `json:"storageSize,omitempty"`
	Service     *ServiceOverrides `json:"service,omitempty"`
}

// PachdOptions allows the user to customize pachd
type PachdOptions struct {
	// Set an ID for the cluster deployment.
	// Defaults to a random value if none is provided
	ClusterID string `json:"clusterDeploymentID,omitempty"`
	// Sets the maximum number of pachd nodes allowed in the cluster.
	// Increasing this number blindly could lead to degraded performance
	// Default: 16
	NumShards int32 `json:"numShards,omitempty" default:"16"`
	// Size of Pachd's in-memory cache for PFS file.
	// Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc)
	BlockCacheBytes string `json:"blockCacheBytes,omitempty"`
	// Resource requests and limits for Pachd
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Require only critical Pachd servers to startup and run without errors.
	// Default: false
	RequireCriticalServers bool `json:"requireCriticalServersOnly,omitempty"`
	// Object storage options for Pachd
	Storage *ObjectStorageOptions `json:"storage,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// The log level option determines the severity of logs
	// that are of interest to the user
	LogLevel string `json:"logLevel,omitempty"`
	// Optional value to determine the format of the logs
	// Default: false
	LokiLogging                      bool              `json:"lokiLogging,omitempty"`
	AuthenticationDisabledForTesting bool              `json:"authenticationDisabledForTesting,omitempty"`
	PPSWorkerGRPCPort                int               `json:"ppsWorkerGRPCPort,omitempty"`
	ExposeDockerSocket               bool              `json:"exposeDockerSocket,omitempty"`
	ExposeObjectAPI                  bool              `json:"exposeObjectAPI,omitempty"`
	Service                          *ServiceOverrides `json:"service,omitempty"`
	Metrics                          *MetricsOptions   `json:"metrics,omitempty"`
	ServiceAccountName               string            `json:"serviceAccountName,omitempty"`
}

// MetricsOptions allows the user to enable/disable pachyderm metrics
type MetricsOptions struct {
	// Default: true
	Enabled  bool   `json:"enabled,omitempty" default:"true"`
	Endpoint string `json:"endpoint,omitempty"`
}

// ObjectStorageOptions exposes options to configure
// object store backend for Pachyderm resource
type ObjectStorageOptions struct {
	// The maximum number of files to upload or fetch from remote sources (HTTP, blob storage) using PutFile concurrently.
	// Default: 100
	PutFileConcurrencyLimit int32 `json:"putFileConcurrencyLimit,omitempty" default:"100"`
	// The maximum number of concurrent object storage uploads per Pachd instance.
	// Default: 100
	UploadFileConcurrencyLimit int32 `json:"uploadFileConcurrencyLimit,omitempty" default:"100"`
	// Sets the type of storage backend.
	// Should be one of "google", "amazon", "minio", "microsoft" of "local"
	Backend string `json:"backend,omitempty"`
	// Configures the Amazon storage backend
	AmazonStorage *AmazonStorageOptions `json:"amazon,omitempty"`
	// Configures the Google storage backend
	GoogleStorage    *GoogleStorageOptions    `json:"google,omitempty"`
	MicrosoftStorage *MicrosoftStorageOptions `json:"microsoft,omitempty"`
	MinioStorage     *MinioStorageOptions     `json:"minio,omitempty"`
	// Kubernetes hostPath
	LocalStorage *LocalStorageOptions `json:"local,omitempty"`
}

// GoogleStorageOptions exposes options to configure Google Cloud Storage
type GoogleStorageOptions struct {
	Bucket             string `json:"bucket,omitempty"`
	CredentialSecret   string `json:"credentialSecret,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// AmazonStorageOptions exposes options to
// configure Amazon s3 storage
type AmazonStorageOptions struct {
	Bucket                 string              `json:"bucket,omitempty"`
	CloudFrontDistribution string              `json:"cloudFrontDistribution,omitempty"`
	CustomEndpoint         string              `json:"customEndpoint,omitempty"`
	DisableSSL             bool                `json:"disableSSL,omitempty"`
	IAMRole                string              `json:"iamRole,omitempty"`
	ID                     string              `json:"id,omitempty"`
	LogOptions             string              `json:"logOptions,omitempty"`
	MaxUploadParts         int                 `json:"maxUploadParts,omitempty"`
	VerifySSL              bool                `json:"verifySSL,omitempty"`
	PartSize               string              `json:"partSize,omitempty"`
	Region                 string              `json:"region,omitempty"`
	Retries                int                 `json:"retries,omitempty"`
	Reverse                bool                `json:"reverse,omitempty"`
	Secret                 string              `json:"secret,omitempty"`
	Timeout                string              `json:"timeout,omitempty"`
	Token                  string              `json:"token,omitempty"`
	UploadACL              string              `json:"uploadACL,omitempty"`
	Vault                  *AmazonStorageVault `json:"vault,omitempty"`
}

// AmazonStorageVault exposes options to configure
// Amazon vault
type AmazonStorageVault struct {
	Address string `json:"address,omitempty"`
	Role    string `json:"role,omitempty"`
	Token   string `json:"token,omitempty"`
}

// MicrosoftStorageOptions exposes options to
// configure Microsoft storage
type MicrosoftStorageOptions struct {
	Container string `json:"container,omitempty"`
	ID        string `json:"id,omitempty"`
	Secret    string `json:"secret,omitempty"`
}

// MinioStorageOptions exposes options to
// confugure Minio object store
type MinioStorageOptions struct {
	Bucket    string `json:"bucket,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	ID        string `json:"id,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Secure    string `json:"secure,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// LocalStorageOptions exposes options to
// confifure local storage
type LocalStorageOptions struct {
	// Location on the worker node to be mounted
	// into the pod
	HostPath string `json:"hostPath,omitempty"`
}

// PachydermPhase defines the data type used
// to report the status of a Pachyderm resource
type PachydermPhase string

const (
	// PhaseInitializing sets the Pachyderm status to initilizing
	PhaseInitializing PachydermPhase = "Initializing"
	// PhaseRunning sets the resource status to running
	PhaseRunning PachydermPhase = "Running"
	// PhaseDeleting reports the resource status to deleting
	PhaseDeleting PachydermPhase = "Deleting"
)

// PachydermStatus defines the observed state of Pachyderm
type PachydermStatus struct {
	Phase PachydermPhase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Pachyderm is the Schema for the pachyderms API
type Pachyderm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PachydermSpec   `json:"spec,omitempty"`
	Status PachydermStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PachydermList contains a list of Pachyderm
type PachydermList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pachyderm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pachyderm{}, &PachydermList{})
}
