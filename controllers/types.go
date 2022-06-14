package controllers

type backup struct {
	CreatedAt *string `form:"created_at,omitempty" json:"created_at,omitempty" xml:"created_at,omitempty"`
	UpdatedAt *string `form:"updated_at,omitempty" json:"updated_at,omitempty" xml:"updated_at,omitempty"`
	DeletedAt *string `form:"deleted_at,omitempty" json:"deleted_at,omitempty" xml:"deleted_at,omitempty"`
	ID        *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Current state of the job
	State *string `form:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	// Name of pachyderm instance backed up
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Namespace of resource backed up
	Namespace *string `form:"namespace,omitempty" json:"namespace,omitempty" xml:"namespace,omitempty"`
	// Name of target pod
	Pod *string `form:"pod,omitempty" json:"pod,omitempty" xml:"pod,omitempty"`
	// Name of container in pod
	Container *string `form:"container,omitempty" json:"container,omitempty" xml:"container,omitempty"`
	// base64 encoded command to run in pod
	Command *string `form:"command,omitempty" json:"command,omitempty" xml:"command,omitempty"`
	// Kubernetes secret containing S3 storage credentials
	StorageSecret *string `form:"storage_secret,omitempty" json:"storage_secret,omitempty" xml:"storage_secret,omitempty"`
	// base64 encoded json representation of object
	KubernetesResource *string `form:"kubernetes_resource,omitempty" json:"kubernetes_resource,omitempty" xml:"kubernetes_resource,omitempty"`
	// URL of the uploaded backup tarball
	Location *string `form:"location,omitempty" json:"location,omitempty" xml:"location,omitempty"`
}

type restore struct {
	CreatedAt *string `form:"created_at,omitempty" json:"created_at,omitempty" xml:"created_at,omitempty"`
	UpdatedAt *string `form:"updated_at,omitempty" json:"updated_at,omitempty" xml:"updated_at,omitempty"`
	DeletedAt *string `form:"deleted_at,omitempty" json:"deleted_at,omitempty" xml:"deleted_at,omitempty"`
	ID        *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Name of pachyderm instance to restore to
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Namespace to restore to
	Namespace *string `form:"namespace,omitempty" json:"namespace,omitempty" xml:"namespace,omitempty"`
	// Key of backup tarball
	BackupLocation *string `form:"backup_location,omitempty" json:"backup_location,omitempty" xml:"backup_location,omitempty"`
	// name of pachyderm instance to restore to
	DestinationName *string `form:"destination_name,omitempty" json:"destination_name,omitempty" xml:"destination_name,omitempty"`
	// namespace to restore pachyderm to
	DestinationNamespace *string `form:"destination_namespace,omitempty" json:"destination_namespace,omitempty" xml:"destination_namespace,omitempty"`
	// Kubernetes secret containing S3 storage credentials
	StorageSecret *string `form:"storage_secret,omitempty" json:"storage_secret,omitempty" xml:"storage_secret,omitempty"`
	// base64 encoded kubernetes object
	KubernetesResource *string `form:"kubernetes_resource,omitempty" json:"kubernetes_resource,omitempty" xml:"kubernetes_resource,omitempty"`
	// base64 encoded database dump
	Database *string `form:"database,omitempty" json:"database,omitempty" xml:"database,omitempty"`
}
