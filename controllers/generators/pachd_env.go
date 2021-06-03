package generators

import (
	"fmt"
	"reflect"
	"strings"

	aimlv1beta1 "github.com/OchiengEd/pachyderm-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func pachdEnvVarirables(pd *aimlv1beta1.Pachyderm) []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	pachdOpts := pd.Spec.Pachd

	if !reflect.DeepEqual(pachdOpts, aimlv1beta1.PachdOptions{}) {
		// enable loki logging
		envs = append(envs, corev1.EnvVar{
			Name:  "LOKI_LOGGING",
			Value: fmt.Sprintf("%t", pachdOpts.LokiLogging),
		})

		// pachd root
		envs = append(envs, corev1.EnvVar{
			Name:  "PACH_ROOT",
			Value: "/pach",
		})

		// etcd prefix
		envs = append(envs, corev1.EnvVar{
			Name:  "ETCD_PREFIX",
			Value: "",
		})

		// number of shards
		envs = append(envs, corev1.EnvVar{
			Name:  "NUM_SHARDS",
			Value: fmt.Sprintf("%d", pachdOpts.NumShards),
		})

		// storage backend
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE_BACKEND",
			Value: strings.ToUpper(pachdOpts.Storage.Backend),
		})

		// storage host path
		if pachdOpts.Storage.LocalStorage != nil {
			envs = append(envs, corev1.EnvVar{
				Name:  "STORAGE_HOST_PATH",
				Value: pachdOpts.Storage.LocalStorage.HostPath,
			})
		}

		if pd.Spec.Worker != nil {
			// worker image
			envs = append(envs, corev1.EnvVar{
				Name:  "WORKER_IMAGE",
				Value: getWorkerImage(pd),
			})

			// worker sidecar image
			envs = append(envs, corev1.EnvVar{
				Name:  "WORKER_SIDECAR_IMAGE",
				Value: "",
			})

			// worker image pull secret
			envs = append(envs, corev1.EnvVar{
				Name:  "WORKER_IMAGE_PULL_POLICY",
				Value: pd.Spec.Worker.Image.PullPolicy,
			})

			// worker service account
			envs = append(envs, corev1.EnvVar{
				Name:  "WORKER_SERVICE_ACCOUNT",
				Value: pd.Spec.Worker.ServiceAccountName,
			})
		}

		if pachdOpts.Image != nil {
			// image pull secret
			envs = append(envs, corev1.EnvVar{
				Name:  "IMAGE_PULL_SECRET",
				Value: pachdOpts.Image.PullPolicy,
			})
		}

		// pachd version
		envs = append(envs, corev1.EnvVar{
			Name:  "PACHD_VERSION",
			Value: "",
		})

		// metrics
		if pachdOpts.Metrics != nil {
			envs = append(envs, corev1.EnvVar{
				Name:  "METRICS",
				Value: fmt.Sprintf("%t", pachdOpts.Metrics.Enabled),
			})
		}

		// log level
		envs = append(envs, corev1.EnvVar{
			Name:  "LOG_LEVEL",
			Value: pachdOpts.LogLevel,
		})

		// block cache bytes
		envs = append(envs, corev1.EnvVar{
			Name:  "BLOCK_CACHE_BYTES",
			Value: pachdOpts.BlockCacheBytes,
		})

		// expose Docker socket
		envs = append(envs, corev1.EnvVar{
			Name:  "NO_EXPOSE_DOCKER_SOCKET",
			Value: fmt.Sprintf("%t", pachdOpts.ExposeDockerSocket),
		})

		// disable pachyderm auth for testing
		envs = append(envs, corev1.EnvVar{
			Name:  "PACHYDERM_AUTHENTICATION_DISABLED_FOR_TESTING",
			Value: fmt.Sprintf("%t", pachdOpts.AuthenticationDisabledForTesting),
		})

		// pachd namespace
		envs = append(envs, corev1.EnvVar{
			Name: "PACH_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		})

		// pachd memory request
		// TODO: handle error
		divisor, _ := resource.ParseQuantity("0")
		envs = append(envs, corev1.EnvVar{
			Name: "PACHD_MEMORY_REQUEST",
			ValueFrom: &corev1.EnvVarSource{
				ResourceFieldRef: &corev1.ResourceFieldSelector{
					ContainerName: "pachd",
					Divisor:       divisor,
					Resource:      "requests.memory",
				},
			},
		})

		// expose object API
		envs = append(envs, corev1.EnvVar{
			Name:  "EXPOSE_OBJECT_API",
			Value: fmt.Sprintf("%t", pachdOpts.ExposeObjectAPI),
		})

		// require critical servers only
		envs = append(envs, corev1.EnvVar{
			Name:  "REQUIRE_CRITICAL_SERVERS_ONLY",
			Value: fmt.Sprintf("%t", pachdOpts.RequireCriticalServers),
		})

		// pachd pod name
		envs = append(envs, corev1.EnvVar{
			Name: "PACHD_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		})

		// require critical servers only
		envs = append(envs, corev1.EnvVar{
			Name:  "PPS_WORKER_GRPC_PORT",
			Value: fmt.Sprintf("%d", pachdOpts.PPSWorkerGRPCPort),
		})

		// require critical servers only
		envs = append(envs, corev1.EnvVar{
			Name:  "STORAGE_V2",
			Value: "80",
		})
	} // .spec.pachd

	// setup pachd storage
	storageOpts := setupPachdStorage(pd)
	if len(storageOpts) > 0 {
		envs = append(envs, storageOpts...)
	}

	return envs
}

func getWorkerImage(pd *aimlv1beta1.Pachyderm) string {
	workerOpt := pd.Spec.Worker

	if workerOpt != nil {
		if workerOpt.Image != nil {
			workerImg := []string{
				workerOpt.Image.Repository,
				workerOpt.Image.ImageTag,
			}
			return strings.Join(workerImg, ":")
		}
	}

	// load default worker images
	return ""
}

func setupPachdStorage(pd *aimlv1beta1.Pachyderm) []corev1.EnvVar {
	pachdOpts := pd.Spec.Pachd
	storageEnv := []corev1.EnvVar{
		{
			Name:  "STORAGE_PUT_FILE_CONCURRENCY_LIMIT",
			Value: fmt.Sprintf("%d", pachdOpts.Storage.PutFileConcurrencyLimit),
		},
		{
			Name:  "STORAGE_UPLOAD_CONCURRENCY_LIMIT",
			Value: fmt.Sprintf("%d", pachdOpts.Storage.PutFileConcurrencyLimit),
		},
	}

	if !reflect.DeepEqual(pd.Spec.Pachd, aimlv1beta1.PachdOptions{}) {

		switch backend := strings.ToLower(pd.Spec.Pachd.Storage.Backend); backend {
		case "amazon":
			if pachdOpts.Storage.Amazon != nil {
				var optional bool = true
				// setup Amazon server configs
				amzn := []corev1.EnvVar{
					{
						Name: "AMAZON_REGION",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-region",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_BUCKET",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-bucket",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_ID",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-id",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_SECRET",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-secret",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-token",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_VAULT_ADDR",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-vault-addr",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_VAULT_ROLE",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-vault-role",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_VAULT_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-vault-token",
								Optional: &optional,
							},
						},
					},
					{
						Name: "AMAZON_DISTRIBUTION",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "amazon-distribution",
								Optional: &optional,
							},
						},
					},
					{
						Name: "CUSTOM_ENDPOINT",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "custom-endpoint",
								Optional: &optional,
							},
						},
					},
					{
						Name: "RETRIES",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "retries",
								Optional: &optional,
							},
						},
					},
					{
						Name: "TIMEOUT",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "timeout",
								Optional: &optional,
							},
						},
					},
					{
						Name: "UPLOAD_ACL",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "upload-acl",
								Optional: &optional,
							},
						},
					},
					{
						Name: "REVERSE",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "reverse",
								Optional: &optional,
							},
						},
					},
					{
						Name: "PART_SIZE",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "part-size",
								Optional: &optional,
							},
						},
					},
					{
						Name: "MAX_UPLOAD_PARTS",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "max-upload-parts",
								Optional: &optional,
							},
						},
					},
					{
						Name: "DISABLE_SSL",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "disable-ssl",
								Optional: &optional,
							},
						},
					},
					{
						Name: "NO_VERIFY_SSL",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "no-verify-ssl",
								Optional: &optional,
							},
						},
					},
					{
						Name: "OBJ_LOG_OPTS",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "pachyderm-storage-secret",
								},
								Key:      "log-options",
								Optional: &optional,
							},
						},
					},
				}

				// append amazon storage options
				storageEnv = append(storageEnv, amzn...)
			}
		case "google":
			google := []corev1.EnvVar{}
			storageEnv = append(storageEnv, google...)
		case "local":
			storageEnv = append(storageEnv, corev1.EnvVar{
				Name:  "STORAGE_HOST_PATH",
				Value: pd.Spec.Pachd.Storage.LocalStorage.HostPath,
			})
		case "minio":
			minio := []corev1.EnvVar{}
			storageEnv = append(storageEnv, minio...)
		}
	}

	return storageEnv
}
