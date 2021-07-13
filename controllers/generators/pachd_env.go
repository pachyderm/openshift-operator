package generators

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (c *PachydermComponents) pachdEnvVarirables() []corev1.EnvVar {
	pd := c.Pachyderm()

	divisor, err := resource.ParseQuantity("0")
	if err != nil {
		fmt.Println("error getting divisor:", err.Error())
	}

	envs := []corev1.EnvVar{
		{
			Name:  "POSTGRES_HOST",
			Value: pd.Spec.Pachd.Postgres.Host,
		},
		{
			Name:  "POSTGRES_PORT",
			Value: fmt.Sprintf("%d", pd.Spec.Pachd.Postgres.Port),
		},
		{
			Name:  "POSTGRES_SERVICE_SSL",
			Value: pd.Spec.Pachd.Postgres.SSL,
		},
		{ // enable loki logging
			Name:  "LOKI_LOGGING",
			Value: fmt.Sprintf("%t", pd.Spec.Pachd.LokiLogging),
		},
		{ // pachd root
			Name:  "PACH_ROOT",
			Value: "/pach",
		},
		{ // etcd prefix
			Name:  "ETCD_PREFIX",
			Value: "",
		},
		{ // storage backend
			Name:  "STORAGE_BACKEND",
			Value: strings.ToUpper(pd.Spec.Pachd.Storage.Backend),
		},
		{
			Name:  "STORAGE_PUT_FILE_CONCURRENCY_LIMIT",
			Value: fmt.Sprintf("%d", pd.Spec.Pachd.Storage.PutFileConcurrencyLimit),
		},
		{
			Name:  "STORAGE_UPLOAD_CONCURRENCY_LIMIT",
			Value: fmt.Sprintf("%d", pd.Spec.Pachd.Storage.PutFileConcurrencyLimit),
		},
		{
			Name:  "WORKER_USES_ROOT",
			Value: "false",
		},
		{
			Name:  "WORKER_IMAGE",
			Value: c.workerImage(),
		},
		{
			Name:  "WORKER_SIDECAR_IMAGE",
			Value: c.workerSidecarImage(),
		},
		{
			Name:  "WORKER_IMAGE_PULL_POLICY",
			Value: c.workerImagePullPolicy(),
		},
		{
			Name:  "WORKER_SERVICE_ACCOUNT",
			Value: pd.Spec.Worker.ServiceAccountName,
		},
		{
			Name:  "IMAGE_PULL_SECRET",
			Value: c.imagePullSecret(),
		},
		{
			Name:  "LOG_LEVEL",
			Value: pd.Spec.Pachd.LogLevel,
		},
		{
			Name:  "PPS_WORKER_GRPC_PORT",
			Value: fmt.Sprintf("%d", pd.Spec.Pachd.PPSWorkerGRPCPort),
		},
		{
			Name:  "REQUIRE_CRITICAL_SERVERS_ONLY",
			Value: fmt.Sprintf("%t", pd.Spec.Pachd.RequireCriticalServers),
		},
		{
			Name:  "METRICS",
			Value: fmt.Sprintf("%t", !pd.Spec.Pachd.Metrics.Disable),
		},
		{
			Name: "PACH_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "PACHD_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "PACHD_MEMORY_REQUEST",
			ValueFrom: &corev1.EnvVarSource{
				ResourceFieldRef: &corev1.ResourceFieldSelector{
					ContainerName: "pachd",
					Divisor:       divisor,
					Resource:      "requests.memory",
				},
			},
		},
	}

	// metrics endpoint
	if pd.Spec.Pachd.Metrics.Endpoint != "" {
		// TODO: check if this is still supported
		envs = append(envs, corev1.EnvVar{
			Name:  "METRICS_ENDPOINT",
			Value: pd.Spec.Pachd.Metrics.Endpoint,
		})
	}

	return envs
}

func (c *PachydermComponents) workerImage() string {
	pd := c.Pachyderm()

	workerImg := strings.Split(c.workerImageName, ":")

	if pd.Spec.Worker.Image != nil {
		if pd.Spec.Worker.Image.Repository != "" {
			workerImg[0] = pd.Spec.Worker.Image.Repository
		}
		if pd.Spec.Worker.Image.ImageTag != "" {
			workerImg[1] = pd.Spec.Worker.Image.ImageTag
		}
	}

	return strings.Join(workerImg, ":")
}

func (c *PachydermComponents) workerSidecarImage() string {
	// load default worker sidecar images
	return c.workerSidecarName
}

func (c *PachydermComponents) workerImagePullPolicy() string {
	pd := c.Pachyderm()
	if pd.Spec.Worker.Image != nil {
		if pd.Spec.Worker.Image.PullPolicy != "" {
			return pd.Spec.Worker.Image.PullPolicy
		}
	}
	return "IfNotPresent"
}

func (c *PachydermComponents) imagePullSecret() string {
	pd := c.Pachyderm()
	if pd.Spec.ImagePullSecret != nil {
		return *pd.Spec.ImagePullSecret
	}
	return ""
}
