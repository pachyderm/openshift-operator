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
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Version",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
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
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account Name",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=pachyderm-worker
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// ConsoleOptions provides options to configure the Pachyderm console
type ConsoleOptions struct {
	// If true, this option disables the Pachyderm dashboard.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Disable",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	Disable bool `json:"disable,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy console
	Image *ImageOverride `json:"image,omitempty"`
	// Optional resource requirements required to run the dash pods.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// The address to use as the host in the dash ingress.
	// Used as the host of a rule
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="URL",xDescriptors={"urn:alm:descriptor:text"}
	URL string `json:"url,omitempty"`
	// Service overrides
	Service *ServiceOverrides `json:"service,omitempty"`
}

// ImageOverride allows the user to override the default image
// and change the image pull policy
type ImageOverride struct {
	// This option dictates the particular image to pull e.g. quay.io/app-sre/ubi8-ubi-minimal
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Image Repo",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Repository string `json:"repository,omitempty"`
	// Used with the image registry to choose a specific
	// image in a cointainer registry to pull e.g. latest
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tag",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Tag string `json:"tag,omitempty"`
	// Determines when images should be pulled.
	// It accepts, "IfNotPresent","Never" or "Always"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Pull Policy",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
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
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Annotations",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Annotations []string `json:"annotations,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Type",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Type string `json:"type"`
}

// EtcdOptions allows users to change the etcd statefulset
type EtcdOptions struct {
	// Optional parameter to set the number of nodes in the Etcd statefulset.
	// Analogous --dynamic-etcd-nodes argument to 'pachctl deploy'
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Dynamic Nodes",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	DynamicNodes int32 `json:"dynamicNodes,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// Resource requests and limits for Etcd
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// If specified, etcd would use an existing storage class for its storage
	// Name of existing storage class to use for the Etcd persistent volume.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Class",xDescriptors={"urn:alm:descriptor:io.kubernetes:StorageClass","urn:alm:descriptor:io.kubernetes:custom"}
	StorageClass string `json:"storageClass,omitempty"`
	// The size of the storage to use for etcd.
	// For example: "100Gi"
	// +kubebuilder:default:="10Gi"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Size",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	StorageSize string `json:"storageSize,omitempty"`
	// Service overrides
	Service *ServiceOverrides `json:"service,omitempty"`
}

// PachdOptions allows the user to customize pachd
type PachdOptions struct {
	// Disable pachd
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Disable",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	Disable bool `json:"disable,omitempty"`
	// Set an ID for the cluster deployment.
	// Defaults to a random value if none is provided
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Cluster ID",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	ClusterID string `json:"clusterDeploymentID,omitempty"`
	// Resource requests and limits for Pachd
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Require only critical Pachd servers to startup and run without errors.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Require Critical Servers Only",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	RequireCriticalServers bool `json:"requireCriticalServersOnly,omitempty"`
	// Object storage options for Pachd
	Storage ObjectStorageOptions `json:"storage,omitempty"`
	// Optional image overrides.
	// Used to specify alternative images to use to deploy dash
	Image *ImageOverride `json:"image,omitempty"`
	// The log level option determines the severity of logs
	// that are of interest to the user
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Level",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=info
	LogLevel string `json:"logLevel,omitempty"`
	// Optional value to determine the format of the logs
	// Default: false
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Loki Logging",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	LokiLogging bool `json:"lokiLogging,omitempty"`
	// Pachyderm Pipeline System(PPS) worker GRPC port.
	// Defaults to port 1080
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="PPS Worker GRPC Port",xDescriptors={"urn:alm:descriptor:number","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default=1080
	PPSWorkerGRPCPort int32 `json:"ppsWorkerGRPCPort,omitempty"`
	// Service overrides for pachd
	Service *ServiceOverrides `json:"service,omitempty"`
	// Allows user to customize metrics options
	Metrics MetricsOptions `json:"metrics,omitempty"`
	// Service Account Name to use for the pachd pod
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account Name",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Postgresql server connection credentials
	Postgres PachdPostgresConfig `json:"postgresql,omitempty"`
}

// PostgresOptions allows user to customize Postgresql
type PostgresOptions struct {
	// Set disabled to true if you choose to provide
	// an external postgresql database
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Disable",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	Disable bool `json:"disable,omitempty"`
	// Storage class for the postgresql persistent storage
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage Class",xDescriptors={"urn:alm:descriptor:io.kubernetes:StorageClass","urn:alm:descriptor:io.kubernetes:custom"}
	StorageClass string `json:"storageClass,omitempty"`
	// Service overrides
	Service ServiceOverrides `json:"service,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resources",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:resourceRequirements"}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PachdPostgresConfig sets up storage for pachd
type PachdPostgresConfig struct {
	// Hostname opr address  of the postgresql host
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Host",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=postgres
	Host string `json:"host,omitempty"`
	// Postgresql port number.
	// Defaults to 5432 when not set
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Port",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=5432
	Port int32 `json:"port,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="SSL",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=disable
	SSL string `json:"ssl,omitempty"`
	// Username to use to connect to the database.
	// Defaults to pachyderm
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="User",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=pachyderm
	User string `json:"user,omitempty"`
	// The password field is used internally by the operator.
	// If the PasswordSecretName is set, it will contain the password
	// read from the secret.
	// Field is not visible to the user
	Password string `json:"-"`
	// Name of the kubernetes secret containing the database password.
	// Password should have the key postgres-password
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Password Secret",xDescriptors={"urn:alm:descriptor:io.kubernetes:Secret","urn:alm:descriptor:io.kubernetes:advanced"}
	PasswordSecretName string `json:"passwordSecret,omitempty"`
	// Name of the database into which the table schemas will live
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Database",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:advanced"}
	// +kubebuilder:default:=pachyderm
	Database string `json:"database,omitempty"`
}

// MetricsOptions allows the user to enable/disable pachyderm metrics
type MetricsOptions struct {
	// If true, this option allows user to disable metrics endpoint.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Disable",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	Disable bool `json:"disable,omitempty"`

	// Option to customize pachd metrics endpoint.
	// When not set, defaults to /metrics
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Endpoint",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Endpoint string `json:"endpoint,omitempty"`
}

// ObjectStorageOptions exposes options to configure
// object store backend for Pachyderm resource
type ObjectStorageOptions struct {
	// The maximum number of files to upload or fetch from remote sources (HTTP, blob storage) using PutFile concurrently.
	// Default: 100
	// +kubebuilder:default:=100
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Put File Concurrency Limit",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	PutFileConcurrencyLimit int32 `json:"putFileConcurrencyLimit,omitempty"`
	// The maximum number of concurrent object storage uploads per Pachd instance.
	// Default: 100
	// +kubebuilder:default:=100
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upload File Concurrency Limit",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	UploadFileConcurrencyLimit int32 `json:"uploadFileConcurrencyLimit,omitempty"`
	// Sets the type of storage backend.
	// Should be one of "GOOGLE", "AMAZON", "MINIO" or "MICROSOFT"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backend",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
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
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Bucket",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Bucket string `json:"bucket,omitempty"`
	// Credentials json file
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Credential Secret",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	CredentialSecret string `json:"credentialSecret,omitempty"`
	// Contents of the "credentials.json" key from the CredentialSecret
	CredentialsData []byte `json:"-"`
	// ServiceAccount used for Google container storage
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Service Account Name",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// AmazonStorageOptions exposes options to
// configure Amazon s3 storage
type AmazonStorageOptions struct {
	// Name of the S3 bucket to hold objects
	Bucket string `json:"-"`
	// CloudFrontDistribution sets the CloudFront distribution in the storage secrets.
	// It is analogous to the --cloudfront-distribution argument to pachctl deploy.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CloudFront Distribution",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	CloudFrontDistribution string `json:"cloudFrontDistribution,omitempty"`
	// Custom endpoint for connecting to S3 object store
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Custom Endpoint",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	CustomEndpoint string `json:"-"`
	// DisableSSL disables SSL.  It is analogous to the --disable-ssl
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Disable SSL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	DisableSSL bool `json:"disableSSL,omitempty"`
	// IAM identity with the desired permissions
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="IAM Role",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	IAMRole string `json:"iamRole,omitempty"`
	// The access ID for the AWS S3 storage solution
	ID string `json:"-"`
	// LogOptions sets various log options in Pachyderm’s internal S3 client.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Log Options",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	LogOptions string `json:"logOptions,omitempty"`
	// Set a custom maximum number of upload parts.
	// Default: 10000
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Max Upload Parts",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	MaxUploadParts int `json:"maxUploadParts,omitempty" default:"10000"`
	// Skip SSL certificate verification.
	// Typically used for enabling self-signed certificates
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Verify SSL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	VerifySSL bool `json:"verifySSL,omitempty"`
	// Set a custom part size for object storage uploads.
	// Default: 5242880
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Part Size",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	PartSize int64 `json:"partSize,omitempty" default:"5242880"`
	// Region for the object storage cluster
	Region string `json:"-"`
	// Set a custom number of retries for object storage requests.
	// Default: 10
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Retries",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:number"}
	Retries int `json:"retries,omitempty" default:"10"`
	// Reverse object storage paths.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Reverse",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:booleanSwitch","urn:alm:descriptor:com.tectonic.ui:custom"}
	Reverse *bool `json:"reverse,omitempty" default:"true"`
	// The secret access key for the S3 bucket
	Secret string `json:"-"`
	// Set a custom timeout for object storage requests.
	// Default: 5m
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Timeout",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	Timeout string `json:"timeout,omitempty" default:"5m"`
	// Token optionally sets the Amazon token to use.  Together with
	// ID and Secret, it implements the functionality of the
	// --credentials argument to pachctl deploy.
	Token string `json:"-"`
	// Sets a custom upload ACL for object store uploads.
	// Default: "bucket-owner-full-control"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Upload ACL",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text","urn:alm:descriptor:io.kubernetes:custom"}
	UploadACL string `json:"uploadACL,omitempty" default:"bucket-owner-full-control"`
	// Container for storing archives
	Vault *AmazonStorageVault `json:"vault,omitempty"`
	// The name of the secret containing the credentials to the S3 storage
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="S3 Credentials Secret",xDescriptors={"urn:alm:descriptor:io.kubernetes:Secret"}
	CredentialSecretName string `json:"credentialSecretName,omitempty" default:"pachyderm-aws-secret"`
}

// AmazonStorageVault exposes options to configure
// Amazon vault
type AmazonStorageVault struct {
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Address",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Address string `json:"address,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Role",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Role string `json:"role,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Token",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Token string `json:"token,omitempty"`
}

// MicrosoftStorageOptions exposes options to
// configure Microsoft storage
type MicrosoftStorageOptions struct {
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Container",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Container string `json:"container,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ID",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	ID string `json:"id,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Secret string `json:"secret,omitempty"`
}

// MinioStorageOptions exposes options to
// confugure Minio object store
type MinioStorageOptions struct {
	// Name of minio bucket to store pachd objects
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Bucket",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Bucket string `json:"bucket,omitempty"`
	// The hostname and port that are used to access the minio object store
	// Example: "minio-server:9000"
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Endpoint",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Endpoint string `json:"endpoint,omitempty"`
	// The user access ID that is used to access minio object store.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="ID",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	ID string `json:"id,omitempty"`
	// The associated password that is used with the user access ID
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secret",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Secret string `json:"secret,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secure",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
	Secure string `json:"secure,omitempty"`
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Signature",xDescriptors={"urn:alm:descriptor:text","urn:alm:descriptor:io.kubernetes:custom"}
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
	// PhaseUpgrading denotes the pachyderm resource is
	// updating from one version to another
	PhaseUpgrading PachydermPhase = "Upgrading"
)

// PachydermStatus defines the observed state of Pachyderm
type PachydermStatus struct {
	// Deployment phase of the pachyderm cluster
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:phase"}
	Phase PachydermPhase `json:"phase,omitempty"`
	// Address of the pachyderm cluster
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Pachd Address",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:pachdAddress"}
	PachdAddress string `json:"pachdAddress,omitempty"`
	// Version of the deployed pachyderm cluster
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Version",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:version"}
	CurrentVersion string `json:"currentVersion,omitempty"`
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
