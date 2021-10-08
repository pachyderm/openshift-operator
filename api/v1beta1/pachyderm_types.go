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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Amazon storage backend for pachd
	AmazonStorageBackend string = "AMAZON"
	// Microsoft storage backend for pachd
	MicrosoftStorageBackend string = "MICROSOFT"
	// Google storage backend for pachd
	GoogleStorageBackend string = "GOOGLE"
	// Minio storage backend for pachd
	MinioStorageBackend string = "MINIO"
)

// PachydermSpec defines the desired state of Pachyderm
type PachydermSpec struct {
	// Allows user to change version of Pachyderm to deploy
	Version string `json:"version,omitempty"`
	// Allows the user to customize the etcd key-value store
	Etcd EtcdOptions `json:"etcd,omitempty"`
	// Allows the user to customize the pachd instance(s)
	Pachd PachdOptions `json:"pachd,omitempty"`
	// Allows the user to customize the console instance(s)
	Console ConsoleOptions `json:"console,omitempty"`
	// Allows user to customize worker instance(s)
	Worker WorkerOptions `json:"worker,omitempty"`
	// Allows user to customize Postgresql database
	Postgres PostgresOptions `json:"postgresql,omitempty"`
	// Allow user to provide an image pull secret
	ImagePullSecret *string `json:"imagePullSecret,omitempty"`
}

// WorkerOptions allows the user to configure workers
type WorkerOptions struct {
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// Name of worker service account.
	// Defaults to pachyderm-worker service account
	// +kubebuilder:default:=pachyderm-worker
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ConsoleOptions provides options to configure the Pachyderm console
type ConsoleOptions struct {
	// If true, this option disables the Pachyderm dashboard.
	Disable bool `json:"disable,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy console
	Image *ImageOverride `json:"image,omitempty"`
	// Optional resource requirements required to run the dash pods.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
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
	Tag string `json:"tag,omitempty"`
	// Determines when images should be pulled.
	// It accepts, "IfNotPresent","Never" or "Always"
	// +kubebuilder:validation:Enum:=IfNotPresent;Always;Never
	PullPolicy string `json:"pullPolicy,omitempty"`
}

// Name represents the full name of the image being override
func (o *ImageOverride) Name() string {
	if strings.Contains(o.Tag, "sha256:") {
		return strings.Join([]string{o.Repository, o.Tag}, "@")
	}
	if o.Tag == "" {
		o.Tag = "latest"
	}
	return strings.Join([]string{o.Repository, o.Tag}, ":")
}

// ImagePullPolicy returns the image pull policy
func (o *ImageOverride) ImagePullPolicy() corev1.PullPolicy {
	return corev1.PullPolicy(o.PullPolicy)
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
	// +kubebuilder:default:="10Gi"
	StorageSize string            `json:"storageSize,omitempty"`
	Service     *ServiceOverrides `json:"service,omitempty"`
}

// PachdOptions allows the user to customize pachd
type PachdOptions struct {
	// Disable pachd
	Disable bool `json:"disable,omitempty"`
	// Set an ID for the cluster deployment.
	// Defaults to a random value if none is provided
	ClusterID string `json:"clusterDeploymentID,omitempty"`
	// Resource requests and limits for Pachd
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
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
	// Pachyderm Pipeline System(PPS) worker GRPC port.
	// Defaults to port 1080
	// +kubebuilder:default=1080
	PPSWorkerGRPCPort int32 `json:"ppsWorkerGRPCPort,omitempty"`
	// Service overrides for pachd
	Service *ServiceOverrides `json:"service,omitempty"`
	// Allows user to customize metrics options
	Metrics            MetricsOptions `json:"metrics,omitempty"`
	ServiceAccountName string         `json:"serviceAccountName,omitempty"`
	// Postgresql server connection credentials
	Postgres PachdPostgresConfig `json:"postgresql,omitempty"`
}

// PostgresOptions allows user to customize Postgresql
type PostgresOptions struct {
	// Set disabled to true if you choose to provide
	// an external postgresql database
	Disable bool `json:"disable,omitempty"`
	// Storage class for the postgresql persistent storage
	StorageClass string                       `json:"storageClass,omitempty"`
	Service      ServiceOverrides             `json:"service,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PachdPostgresConfig
type PachdPostgresConfig struct {
	// Hostname opr address  of the postgresql host
	// +kubebuilder:default:=postgres
	Host string `json:"host,omitempty"`
	// Postgresql port number.
	// Defaults to 5432 when not set
	// +kubebuilder:default:=5432
	Port int32 `json:"port,omitempty"`
	// +kubebuilder:default:=disable
	SSL string `json:"ssl,omitempty"`
	// Username to use to connect to the database.
	// Defaults to pachyderm
	// +kubebuilder:default:=pachyderm
	User string `json:"user,omitempty"`
	// Will be autogenerated if left empty
	Password string `json:"-"`
	// Name of the kubernetes secret containing the database password.
	// Password should be in a secret with key postgres-password
	PasswordSecretName string `json:"passwordSecret,omitempty"`
	// Name of the database into which the table schemas will live
	// +kubebuilder:default:=pachyderm
	Database string `json:"database,omitempty"`
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
	// Should be one of "GOOGLE", "AMAZON", "MINIO" or "MICROSOFT"
	// +kubebuilder:validation:Enum:=AMAZON;MINIO;MICROSOFT;GOOGLE
	Backend string `json:"backend"`
	// Configures the Amazon storage backend
	Amazon *AmazonStorageOptions `json:"amazon,omitempty"`
	// Configures the Google storage backend
	Google *GoogleStorageOptions `json:"google,omitempty"`
	// Configures Microsoft storage backend
	Microsoft *MicrosoftStorageOptions `json:"microsoft,omitempty"`
	// Configures Minio object store
	Minio *MinioStorageOptions `json:"minio,omitempty"`
}

// GoogleStorageOptions exposes options to configure Google Cloud Storage
type GoogleStorageOptions struct {
	// Name of GCS bucket to hold objects
	Bucket string `json:"bucket,omitempty"`
	// Credentials json file
	CredentialSecret string `json:"credentialSecret,omitempty"`
	// Contents of the "credentials.json" key from the CredentialSecret
	CredentialsData []byte `json:"-"`
	// ServiceAccount used for Google container storage
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// AmazonStorageOptions exposes options to
// configure Amazon s3 storage
type AmazonStorageOptions struct {
	// Name of the S3 bucket to hold objects
	Bucket string `json:"-"`
	// CloudFrontDistribution sets the CloudFront distribution in the storage secrets.
	// It is analogous to the --cloudfront-distribution argument to pachctl deploy.
	CloudFrontDistribution string `json:"cloudFrontDistribution,omitempty"`
	// Custom endpoint for connecting to S3 object store
	CustomEndpoint string `json:"-"`
	// DisableSSL disables SSL.  It is analogous to the --disable-ssl
	DisableSSL bool `json:"disableSSL,omitempty"`
	// IAM identity with the desired permissions
	IAMRole string `json:"iamRole,omitempty"`
	// The access ID for the AWS S3 storage solution
	ID string `json:"-"`
	// LogOptions sets various log options in Pachyderm’s internal S3 client.
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
	// Region for the object storage cluster
	Region string `json:"-"`
	// Set a custom number of retries for object storage requests.
	// Default: 10
	Retries int `json:"retries,omitempty" default:"10"`
	// Reverse object storage paths.
	Reverse *bool `json:"reverse,omitempty" default:"true"`
	// The secret access key for the S3 bucket
	Secret string `json:"-"`
	// Set a custom timeout for object storage requests.
	// Default: 5m
	Timeout string `json:"timeout,omitempty" default:"5m"`
	// Token optionally sets the Amazon token to use.  Together with
	// ID and Secret, it implements the functionality of the
	// --credentials argument to pachctl deploy.
	Token string `json:"-"`
	// Sets a custom upload ACL for object store uploads.
	// Default: "bucket-owner-full-control"
	UploadACL string `json:"uploadACL,omitempty" default:"bucket-owner-full-control"`
	// Container for storing archives
	Vault *AmazonStorageVault `json:"vault,omitempty"`
	// The name of the secret containing the credentials to the S3 storage
	CredentialSecretName string `json:"credentialSecretName,omitempty" default:"pachyderm-aws-secret"`
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
	Phase        PachydermPhase `json:"phase"`
	PachdAddress string         `json:"pachdAddress,omitempty"`
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
