package generators

import (
	"fmt"
	"reflect"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// PachydermCluster is a structure that contains
// all the Kubernetes resources that make up a Pachyderm cluster
type PachydermCluster struct {
	pachyderm           *aimlv1beta1.Pachyderm
	etcdStatefulSet     *appsv1.StatefulSet
	postgreStatefulSet  *appsv1.StatefulSet
	Pod                 *corev1.Pod
	ClusterRoleBindings []*rbacv1.ClusterRoleBinding
	ClusterRoles        []*rbacv1.ClusterRole
	RoleBindings        []*rbacv1.RoleBinding
	Roles               []*rbacv1.Role
	ServiceAccounts     []*corev1.ServiceAccount
	Services            []*corev1.Service
	secrets             []*corev1.Secret
	configMaps          []*corev1.ConfigMap
	storageClasses      []*storagev1.StorageClass
	deployments         []*appsv1.Deployment
}

func getPachydermCluster(pd *aimlv1beta1.Pachyderm) (*PachydermCluster, error) {
	manifests, err := loadPachydermTemplates(pd)
	if err != nil {
		return nil, err
	}

	cluster := &PachydermCluster{
		pachyderm: pd,
	}
	for _, doc := range manifests {
		yamlDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, gvk, err := yamlDecoder.Decode([]byte(doc), nil, obj)
		if err != nil {
			fmt.Println("error converting to unstructured type.", err.Error())
		}

		// Convert from unstructured.Unstructured to kubernetes types
		switch gvk.Kind {
		case "Deployment":
			deployment := &appsv1.Deployment{}
			if err := toTypedResource(obj, &deployment); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
			cluster.deployments = append(cluster.deployments, deployment)
		case "StatefulSet":
			if err := cluster.parseStatefulSet(obj); err != nil {
				fmt.Println("error parsing statefulset.", err.Error())
			}
		case "Pod":
			pod := &corev1.Pod{}
			if err := toTypedResource(obj, cluster.Pod); err != nil {
				fmt.Println("error parsing pod.", err.Error())
			}
			cluster.Pod = pod
		case "ServiceAccount":
			sa := &corev1.ServiceAccount{}
			if err := toTypedResource(obj, sa); err != nil {
				fmt.Println("error converting to service account.", err.Error())
			}
			cluster.ServiceAccounts = append(cluster.ServiceAccounts, sa)
		case "Secret":
			secret := &corev1.Secret{}
			if err := toTypedResource(obj, secret); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			cluster.secrets = append(cluster.secrets, secret)
		case "ConfigMap":
			cm := &corev1.ConfigMap{}
			if err := toTypedResource(obj, cm); err != nil {
				fmt.Println("error converting to config map.", err.Error())
			}
			cluster.configMaps = append(cluster.configMaps, cm)
		case "StorageClass":
			sc := &storagev1.StorageClass{}
			if err := toTypedResource(obj, sc); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			cluster.storageClasses = append(cluster.storageClasses, sc)
		case "ClusterRole":
			clusterrole := &rbacv1.ClusterRole{}
			if err := toTypedResource(obj, clusterrole); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			cluster.ClusterRoles = append(cluster.ClusterRoles, clusterrole)
		case "ClusterRoleBinding":
			clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
			if err := toTypedResource(obj, clusterRoleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			cluster.ClusterRoleBindings = append(cluster.ClusterRoleBindings, clusterRoleBinding)
		case "Role":
			role := &rbacv1.Role{}
			if err := toTypedResource(obj, role); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			cluster.Roles = append(cluster.Roles, role)
		case "RoleBinding":
			roleBinding := &rbacv1.RoleBinding{}
			if err := toTypedResource(obj, roleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			cluster.RoleBindings = append(cluster.RoleBindings, roleBinding)
		case "Service":
			svc := &corev1.Service{}
			if err := toTypedResource(obj, svc); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			cluster.Services = append(cluster.Services, svc)
		}
	}

	return cluster, nil
}

func toTypedResource(unstructured *unstructured.Unstructured, object interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, object)
}

func (c *PachydermCluster) parseStatefulSet(obj *unstructured.Unstructured) error {
	var sts appsv1.StatefulSet
	if err := toTypedResource(obj, &sts); err != nil {
		fmt.Println("error converting to statefulset.", err.Error())
	}

	if !reflect.DeepEqual(sts, appsv1.StatefulSet{}) {
		switch sts.Name {
		case "etcd":
			c.etcdStatefulSet = &sts
		case "postgres":
			c.postgreStatefulSet = &sts
		}
	}

	return nil
}

// Pachyderm returns the pachyderm resource used to configure components
func (c *PachydermCluster) Pachyderm() *aimlv1beta1.Pachyderm {
	return c.pachyderm
}

// StorageClasses returns a new etcd storage class
// if an existing one is not used or provided
func (c *PachydermCluster) StorageClasses() []*storagev1.StorageClass {
	return c.storageClasses
}

// Secrets returns secrets used by the pachyderm resource
func (c *PachydermCluster) Secrets() []*corev1.Secret {
	return c.secrets
}

// ConfigMaps returns a slice of pachyderm cluster configmaps
func (c *PachydermCluster) ConfigMaps() []*corev1.ConfigMap {
	return c.configMaps
}

// EtcdStatefulSet returns the etcd statefulset resource
func (c *PachydermCluster) EtcdStatefulSet() *appsv1.StatefulSet {
	etcd := c.etcdStatefulSet
	catalog, _ := pachydermImagesCatalog(c.Pachyderm())
	for i, container := range etcd.Spec.Template.Spec.Containers {
		if container.Name == "etcd" {
			etcd.Spec.Template.Spec.Containers[i].Image = catalog.etcdImage().Name()
		}
	}

	// Remove security context deployed by the charts.
	// This conflicts with the random user ID range provided by Openshift
	if etcd.Spec.Template.Spec.SecurityContext != nil {
		etcd.Spec.Template.Spec.SecurityContext = nil
	}

	return etcd
}

// PostgreStatefulset returns the postgresql statefulset resource
func (c *PachydermCluster) PostgreStatefulset() *appsv1.StatefulSet {
	pg := c.postgreStatefulSet
	pd := c.Pachyderm()
	catalog, _ := pachydermImagesCatalog(pd)
	pgImage := catalog.postgresqlImage()

	pg.Spec.Template.Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "postgres",
				Image:           pgImage.Name(),
				ImagePullPolicy: corev1.PullPolicy(pgImage.PullPolicy),
				Env: []corev1.EnvVar{
					{
						Name:  "POSTGRESQL_USER",
						Value: pd.Spec.Pachd.Postgres.User,
					},
					{
						Name: "POSTGRESQL_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "postgres",
								},
								Key: "postgresql-password",
							},
						},
					},
					{
						Name: "POSTGRESQL_ADMIN_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "postgres",
								},
								Key: "postgresql-postgres-password",
							},
						},
					},
					{
						Name:  "POSTGRESQL_DATABASE",
						Value: pd.Spec.Pachd.Postgres.Database,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "tcp-postgresql",
						Protocol:      corev1.ProtocolTCP,
						ContainerPort: 5432,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/var/lib/pgsql",
						ReadOnly:  false,
					},
					{
						Name:      "custom-init-scripts",
						MountPath: "/docker-entrypoint-initdb.d/",
						ReadOnly:  true,
					},
					{
						Name:      "dshm",
						MountPath: "/dev/shm",
					},
				},
			},
		},
		Volumes: pg.Spec.Template.Spec.Volumes,
	}

	return pg
}

// PrepareCluster takes a pachyderm custom resource and returns
// child resources based on the pachyderm custom resource
// TODO: decode any input here
func PrepareCluster(pd *aimlv1beta1.Pachyderm) (*PachydermCluster, error) {
	cluster, err := getPachydermCluster(pd)
	if err != nil {
		return nil, err
	}

	for _, deployment := range cluster.deployments {
		if deployment.Name == "pg-bouncer" {
			setupPGBouncer(pd, deployment)
		}
		if deployment.Name == "pachd" {
			setupPachd(pd, deployment)
		}
	}

	return cluster, nil
}

// Deployments returns slice of deployments generated by the helm template command
func (c *PachydermCluster) Deployments() []*appsv1.Deployment {
	return c.deployments
}

func setupPGBouncer(pd *aimlv1beta1.Pachyderm, bouncer *appsv1.Deployment) {
	catalog, _ := pachydermImagesCatalog(pd)
	bouncerImage := catalog.pgBouncerImage()

	// pg-bouncer implementation
	for i, container := range bouncer.Spec.Template.Spec.Containers {
		if container.Name == "pg-bouncer" {
			container.Image = bouncerImage.Name()
			container.ImagePullPolicy = bouncerImage.ImagePullPolicy()
			container.VolumeMounts = []corev1.VolumeMount{
				{
					Name:      "config",
					MountPath: "/pgconf",
				},
			}
			bouncer.Spec.Template.Spec.Containers[i] = container
		}
	}
	if bouncer.Spec.Template.Spec.Volumes == nil {
		bouncer.Spec.Template.Spec.Volumes = []corev1.Volume{}
	}
	bouncer.Spec.Template.Spec.Volumes = append(
		bouncer.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	)
}

func setupPachd(pd *aimlv1beta1.Pachyderm, pachd *appsv1.Deployment) {
	catalog, _ := pachydermImagesCatalog(pd)
	pachdImage := catalog.pachdImage()
	pgImage := catalog.postgresqlImage()
	utilsImage := catalog.utilsImage()

	// Remove securityContext applied by helm chart in favour of
	// randomized user ID provided by Openshift
	if pachd.Spec.Template.Spec.SecurityContext != nil {
		pachd.Spec.Template.Spec.SecurityContext = nil
	}

	for i, initContainer := range pachd.Spec.Template.Spec.InitContainers {
		if initContainer.Name != "init-etcd" {
			pachd.Spec.Template.Spec.InitContainers[i].Image = pgImage.Name()
		}
		if initContainer.Name == "init-etcd" {
			pachd.Spec.Template.Spec.InitContainers[i].Image = utilsImage.Name()
		}
	}

	for i, container := range pachd.Spec.Template.Spec.Containers {
		if container.Name == "pachd" {
			var env []corev1.EnvVar
			for _, environment := range pachd.Spec.Template.Spec.Containers[i].Env {
				if environment.Name == "POSTGRES_HOST" {
					environment.Value = pd.Spec.Pachd.Postgres.Host
				}
				if environment.Name == "POSTGRES_USER" {
					environment.Value = pd.Spec.Pachd.Postgres.User
				}
				if environment.Name == "POSTGRES_DATABASE" {
					environment.Value = pd.Spec.Pachd.Postgres.Database
				}
				if environment.Name == "WORKER_IMAGE" {
					environment.Value = catalog.workerImage().Name()
				}
				if environment.Name == "WORKER_SIDECAR_IMAGE" {
					environment.Value = catalog.pachdImage().Name()
				}
				env = append(env, environment)
			}
			pachd.Spec.Template.Spec.Containers[i].Env = env
			pachd.Spec.Template.Spec.Containers[i].Image = pachdImage.Name()
			pachd.Spec.Template.Spec.Containers[i].ImagePullPolicy = pachdImage.ImagePullPolicy()
		}
	}
}
