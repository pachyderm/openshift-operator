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
	// Allows user to change version of Pachyderm to deploy
	Version string `json:"version,omitempty"`
	// Allows the user to customize the etcd key-value store
	Etcd EtcdOptions `json:"etcd,omitempty"`
	// Allows the user to customize the pachd instance(s)
	Pachd PachdOptions `json:"pachd,omitempty"`
	// Allows the user to customize the dashd instance(s)
	Dashd DashOptions `json:"dash,omitempty"`
	// Allows user to customize worker instance(s)
	Worker *WorkerOptions `json:"worker,omitempty"`
	// Allows user to customize Postgresql database
	Postgres PostgresOptions `json:"postgresql,omitempty"`
}

// WorkerOptions allows the user to configure workers
type WorkerOptions struct {
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// Name of worker service account
	// +kubebuilder:default:=pachyderm-worker
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// DashOptions provides options to configure the dashd component
type DashOptions struct {
	// If true, this option disables the Pachyderm dashboard.
	Disable bool `json:"disable,omitempty"`
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
	// Determines when images should be pulled.
	// It accepts, "IfNotPresent","Never" or "Always"
	// +kubebuilder:validation:Enum:=IfNotPresent;Always;Never
	PullPolicy string `json:"pullPolicy,omitempty"`
}

// ServiceOverrides allows user to customize k8s
// service type and annotations
type ServiceOverrides struct {
	Annotations []string `json:"annotations,omitempty"`
	Type        string   `json:"type"`
}

// EtcdOptions allows users to change the etcd statefulset
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
	// Increasing this number blindly could lead to degraded performance.
	// Default: 16
	// +kubebuilder:default:=16
	NumShards int32 `json:"numShards,omitempty"`
	// Size of Pachd's in-memory cache for PFS file.
	// Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc)
	BlockCacheBytes string `json:"blockCacheBytes,omitempty"`
	// Resource requests and limits for Pachd
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Require only critical Pachd servers to startup and run without errors.
	RequireCriticalServers bool `json:"requireCriticalServersOnly,omitempty"`
	// Object storage options for Pachd
	Storage ObjectStorageOptions `json:"storage,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// The log level option determines the severity of logs
	// that are of interest to the user
	// +kubebuilder:default:=info
	LogLevel string `json:"logLevel,omitempty"`
	// Optional value to determine the format of the logs
	// Default: false
	LokiLogging bool `json:"lokiLogging,omitempty"`
	// When true, allows user to disable authentication during testing
	AuthenticationDisabledForTesting bool `json:"authenticationDisabledForTesting,omitempty"`
	// Pachyderm Pipeline System(PPS) worker GRPC port
	// +kubebuilder:default:=1080
	PPSWorkerGRPCPort int `json:"ppsWorkerGRPCPort,omitempty"`
	// Expose the Docker socket to worker containers.
	// When false, limits the worker container privileges preventing them from
	// automatically setting the container's working dir and user
	ExposeDockerSocket bool `json:"exposeDockerSocket,omitempty"`
	// If set, instructs pachd to serve its object/block API on its public port.
	// Do not  use in production
	ExposeObjectAPI bool              `json:"exposeObjectAPI,omitempty"`
	Service         *ServiceOverrides `json:"service,omitempty"`
	// Allows user to customize metrics options
	Metrics            *MetricsOptions `json:"metrics,omitempty"`
	ServiceAccountName string          `json:"serviceAccountName,omitempty"`
	// Postgresql server connection credentials
	Postgres PachdPostgresConfig `json:"postgresql,omitempty"`
}

// PostgresOptions allows user to customize Postgresql
type PostgresOptions struct {
	Disabled     bool                        `json:"disabled,omitempty"`
	StorageClass string                      `json:"storageClass,omitempty"`
	Service      ServiceOverrides            `json:"service,omitempty"`
	Resources    corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PachdPostgresConfig
type PachdPostgresConfig struct {
	// +kubebuilder:default:=postgres
	Host string `json:"host,omitempty"`
	// +kubebuilder:default:=5432
	Port int32 `json:"port,omitempty"`
	// +kubebuilder:default:=disable
	SSL      string `json:"ssl,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// MetricsOptions allows the user to enable/disable pachyderm metrics
type MetricsOptions struct {
	// If true, this option allows user to disable metrics endpoint.
	Disable bool `json:"disable,omitempty"`

	// Option to customize pachd metrics endpoint.
	// When not set, defaults to /metrics
	Endpoint string `json:"endpoint,omitempty"`
}

// ObjectStorageOptions exposes options to configure
// object store backend for Pachyderm resource
type ObjectStorageOptions struct {
	// The maximum number of files to upload or fetch from remote sources (HTTP, blob storage) using PutFile concurrently.
	// Default: 100
	// +kubebuilder:default:=100
	PutFileConcurrencyLimit int32 `json:"putFileConcurrencyLimit,omitempty"`
	// The maximum number of concurrent object storage uploads per Pachd instance.
	// Default: 100
	// +kubebuilder:default:=100
	UploadFileConcurrencyLimit int32 `json:"uploadFileConcurrencyLimit,omitempty"`
	// Sets the type of storage backend.
	// Should be one of "google", "amazon", "minio", "microsoft" or "local"
	// +kubebuilder:validation:Enum:=amazon;minio;microsoft;local;google
	Backend string `json:"backend"`
	// Configures the Amazon storage backend
	Amazon *AmazonStorageOptions `json:"amazon,omitempty"`
	// Configures the Google storage backend
	Google *GoogleStorageOptions `json:"google,omitempty"`
	// Configures Microsoft storage backend
	Microsoft *MicrosoftStorageOptions `json:"microsoft,omitempty"`
	// Configures Minio object store
	Minio *MinioStorageOptions `json:"minio,omitempty"`
	// Kubernetes hostPath
	Local *LocalStorageOptions `json:"local,omitempty"`
}

// GoogleStorageOptions exposes options to configure Google Cloud Storage
type GoogleStorageOptions struct {
	// Name of GCS bucket to hold objects
	Bucket string `json:"bucket,omitempty"`
	// Credentials json file
	CredentialSecret   string `json:"credentialSecret,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// AmazonStorageOptions exposes options to
// configure Amazon s3 storage
type AmazonStorageOptions struct {
	// Name of the S3 bucket to hold objects
	Bucket string `json:"bucket,omitempty"`
	// AWS cloudfront distribution
	CloudFrontDistribution string `json:"cloudFrontDistribution,omitempty"`
	// Custom endpoint for connecting to S3 object store
	CustomEndpoint string `json:"customEndpoint,omitempty"`
	// Disable SSL.
	DisableSSL bool `json:"disableSSL,omitempty"`
	// IAM identity with the desired permissions
	IAMRole string `json:"iamRole,omitempty"`
	// Set an ID for the cluster deployment.
	// Defaults to a random value.
	ID string `json:"id,omitempty"`
	// Enable verbose logging in Pachyderm's internal S3 client for debugging.
	LogOptions string `json:"logOptions,omitempty"`
	// Set a custom maximum number of upload parts.
	// Default: 10000
	MaxUploadParts int `json:"maxUploadParts,omitempty" default:"10000"`
	// Skip SSL certificate verification.
	// Typically used for enabling self-signed certificates
	VerifySSL bool `json:"verifySSL,omitempty"`
	// Set a custom part size for object storage uploads.
	// Default: 5242880
	PartSize int64 `json:"partSize,omitempty" default:"5242880"`
	// Region for the object storqge cluster
	Region string `json:"region,omitempty"`
	// Set a custom number of retries for object storage requests.
	// Default: 10
	Retries int `json:"retries,omitempty" default:"10"`
	// Reverse object storage paths.
	// +kubebuilder:default:=true
	Reverse *bool `json:"reverse,omitempty"`
	// The secret access key for the S3 bucket
	Secret string `json:"secret,omitempty"`
	// Set a custom timeout for object storage requests.
	// Default: 5m
	Timeout string `json:"timeout,omitempty" default:"5m"`
	Token   string `json:"token,omitempty"`
	// Sets a custom upload ACL for object store uploads.
	// Default: "bucket-owner-full-control"
	UploadACL string `json:"uploadACL,omitempty" default:"bucket-owner-full-control"`
	// Container for storing archives
	Vault *AmazonStorageVault `json:"vault,omitempty"`
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
	// Name of minio bucket to store pachd objects
	Bucket string `json:"bucket,omitempty"`
	// The hostname and port that are used to access the minio object store
	// Example: "minio-server:9000"
	Endpoint string `json:"endpoint,omitempty"`
	// The user access ID that is used to access minio object store.
	ID string `json:"id,omitempty"`
	// The associated password that is used with the user access ID
	Secret    string `json:"secret,omitempty"`
	Secure    string `json:"secure,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// LocalStorageOptions exposes options to
// confifure local storage
type LocalStorageOptions struct {
	// Location on the worker node to be
	// mounted into the pod.
	// Default: "/var/pachyderm/"
	HostPath string `json:"hostPath,omitempty" default:"/var/pachyderm/"`
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
