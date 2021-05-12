package generators

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	aimlv1beta1 "github.com/OchiengEd/pachyderm-operator/api/v1beta1"
	goyaml "github.com/go-yaml/yaml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

const (
	version string = "0.5.0"
)

// PachydermError defines custom error
// type used by the operator
type PachydermError string

func (e PachydermError) Error() string {
	return string(e)
}

func getManifestPath() string {
	manifestDir := filepath.Join("/", "manifests", version, "manifests.yaml")

	// Check operator is not running in Openshift
	if !isKubernetes() {
		wd, err := os.Getwd()
		if err != nil {
			return manifestDir
		}
		manifestDir = filepath.Join(wd, "hack", "manifests", version, "manifests.yaml")
	}
	return manifestDir
}

// isKubernetes() function checks if the pod is
// runnign in a Kubernetes environment
func isKubernetes() bool {
	const serviceAccountMount string = "/run/secrets/kubernetes.io/serviceaccount"

	fileInfo, err := os.Stat(serviceAccountMount)
	if err != nil {
		// if file is not found
		return false
	}

	// check if Kubernetes port environment variable exists
	_, ok := os.LookupEnv("KUBERNETES_PORT")

	return fileInfo.IsDir() && ok
}

func loadManifests() ([][]byte, error) {
	var objects [][]byte

	// Read manifests from file
	manifestPath := getManifestPath()
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	// Parse the manifest into individual Kubernetes object
	decodr := goyaml.NewDecoder(bytes.NewReader(data))
	for {
		var value interface{}
		if err := decodr.Decode(&value); err == io.EOF {
			break
		}
		valueBytes, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		objects = append(objects, valueBytes)
	}

	return objects, nil
}

// PachydermComponents is a structure that contains a slice of
// all the Kubernetes resources that make up a Pachyderm deployment
type PachydermComponents struct {
	parent              *aimlv1beta1.Pachyderm
	dashDeploy          *appsv1.Deployment
	pachdDeploy         *appsv1.Deployment
	etcdStatefulSet     *appsv1.StatefulSet
	Pod                 *corev1.Pod
	ClusterRoleBindings []rbacv1.ClusterRoleBinding
	ClusterRoles        []rbacv1.ClusterRole
	RoleBindings        []rbacv1.RoleBinding
	Roles               []rbacv1.Role
	ServiceAccounts     []corev1.ServiceAccount
	Services            []corev1.Service
	secrets             []corev1.Secret
	storageClass        storagev1.StorageClass
}

func getPachydermComponents() PachydermComponents {
	components := PachydermComponents{}

	manifests, err := loadManifests()
	if err != nil {
		fmt.Println("error reading manifests.", err.Error())
	}

	for _, doc := range manifests {
		yamlDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		_, gvk, err := yamlDecoder.Decode(doc, nil, obj)
		if err != nil {
			fmt.Println("error converting to unstructured type.", err.Error())
		}

		// Convert from unstructured.Unstructured to kubernetes types
		switch gvk.Kind {
		case "Deployment":
			if err := components.parseDeployment(obj); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
		case "StatefulSet":
			if err := components.parseStatefulSet(obj); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
		case "Pod":
			if err := components.parsePod(obj); err != nil {
				fmt.Println("error parsing pod.", err.Error())
			}
		case "ServiceAccount":
			var sa corev1.ServiceAccount
			if err := toTypedResource(obj, &sa); err != nil {
				fmt.Println("error converting to service account.", err.Error())
			}
			components.ServiceAccounts = append(components.ServiceAccounts, sa)
		case "Secret":
			var secret corev1.Secret
			if err := toTypedResource(obj, &secret); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			components.secrets = append(components.secrets, secret)
		case "StorageClass":
			var sc storagev1.StorageClass
			if err := toTypedResource(obj, &sc); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			components.storageClass = sc
		case "ClusterRole":
			var clusterrole rbacv1.ClusterRole
			if err := toTypedResource(obj, &clusterrole); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.ClusterRoles = append(components.ClusterRoles, clusterrole)
		case "ClusterRoleBinding":
			var clusterRoleBinding rbacv1.ClusterRoleBinding
			if err := toTypedResource(obj, &clusterRoleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, clusterRoleBinding)
		case "Role":
			var role rbacv1.Role
			if err := toTypedResource(obj, &role); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.Roles = append(components.Roles, role)
		case "RoleBinding":
			var roleBinding rbacv1.RoleBinding
			if err := toTypedResource(obj, &roleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.RoleBindings = append(components.RoleBindings, roleBinding)
		case "Service":
			var svc corev1.Service
			if err := toTypedResource(obj, &svc); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.Services = append(components.Services, svc)
		}
	}

	return components
}

func toTypedResource(unstructured *unstructured.Unstructured, object interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, object)
}

func (c *PachydermComponents) parseDeployment(obj *unstructured.Unstructured) error {
	var deployment *appsv1.Deployment
	if err := toTypedResource(obj, &deployment); err != nil {
		return err
	}

	switch deployment.Name {
	case "dash":
		c.dashDeploy = deployment
	case "pachd":
		c.pachdDeploy = deployment
	}

	return nil
}

func (c *PachydermComponents) parseStatefulSet(obj *unstructured.Unstructured) error {
	var sts *appsv1.StatefulSet
	if err := toTypedResource(obj, &sts); err != nil {
		fmt.Println("error converting to statefulset.", err.Error())
	}

	if sts != nil && sts.Name == "etcd" {
		c.etcdStatefulSet = sts
	}

	return nil
}

func (c *PachydermComponents) parsePod(obj *unstructured.Unstructured) error {
	var pod *corev1.Pod
	if err := toTypedResource(obj, &pod); err != nil {
		return err
	}

	if c.Pod != nil {
		c.Pod = pod
	}

	return nil
}

// Parent returns the pachyderm resource used to configure components
func (c *PachydermComponents) Parent() *aimlv1beta1.Pachyderm {
	return c.parent
}

func (c *PachydermComponents) StorageClass() *storagev1.StorageClass {
	// 	pd := c.parent
	// 	var allowExpansion bool = true

	// 	if pd.Spec.Etcd != nil && pd.Spec.Etcd.StorageClass == "" {
	// 		return nil
	// 	}

	// 	if pd.Spec.Etcd != nil {
	// 		sc := &c.storageClass
	// 		sc.Namespace = pd.Namespace
	// 		sc.AllowVolumeExpansion = &allowExpansion

	// 		switch pd.Spec.Etcd.StorageProvider {
	// 		case "google":
	// 			sc.Provisioner = "kubernetes.io/gce-pd"
	// 			sc.Parameters = map[string]string{
	// 				"type": "pd-ssd",
	// 			}
	// 		case "amazon":
	// 			sc.Provisioner = "kubernetes.io/aws-ebs"
	// 			sc.Parameters = map[string]string{
	// 				"type": "gp2",
	// 			}
	// 		}

	// 		return &c.storageClass
	// 	}

	return nil
}

func (c *PachydermComponents) Secrets() []corev1.Secret {
	pd := c.parent

	for _, secret := range c.secrets {
		secret.Namespace = pd.Namespace

		if secret.Name == "pachyderm-storage-secret" {
			setupStorageSecret(&secret, pd)
		}

		if secret.Name == "pachd-tls-cert" {
			setupPachdTLSSecret(&secret, pd)
		}
	}

	return c.secrets
}

func setupStorageSecret(secret *corev1.Secret, pd *aimlv1beta1.Pachyderm) {
	// data := secret.Data

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreBucket) != 0 {
	// 	data["amazon-bucket"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreBucket
	// }

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreSecret) != 0 {
	// 	data["amazon-secret"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreSecret
	// }

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreEndpoint) != 0 {
	// 	data["custom-endpoint"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreEndpoint
	// }

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreRegion) != 0 {
	// 	data["amazon-region"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreRegion
	// }

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreToken) != 0 {
	// 	data["amazon-token"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreToken
	// }

	// if len(pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreID) != 0 {
	// 	data["amazon-id"] = pd.Spec.Pachd.Storage.AmazonS3.ObjectStoreID
	// }
}

// TODO: generate self-signed TLS secret
func setupPachdTLSSecret(secret *corev1.Secret, pd *aimlv1beta1.Pachyderm) {
	data := secret.Data

	data["tls.crt"] = []byte{}
	data["tls.key"] = []byte{}
}

// EtcdStatefulSet returns the etcd statefulset resource
func (c *PachydermComponents) EtcdStatefulSet() *appsv1.StatefulSet {
	pd := c.parent

	if pd.Spec.Etcd != nil {
		for _, container := range c.etcdStatefulSet.Spec.Template.Spec.Containers {
			if container.Name == "etcd" {
				container.Resources = pd.Spec.Etcd.Resources
			}
		}
	}

	return c.etcdStatefulSet
}

// PachdDeployment returns the pachd deployment resource
func (c *PachydermComponents) PachdDeployment() *appsv1.Deployment {
	deploy := c.pachdDeploy

	for _, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == "pachd" {
			container.Env = pachdEnvVarirables(c.parent)
		}
	}

	return c.pachdDeploy
}

// DashDeployment returns the dash deployment resource
func (c *PachydermComponents) DashDeployment() *appsv1.Deployment {
	return c.dashDeploy
}

// StorageClass returns the etcd storage class resource
// func (c *PachydermComponents) StorageClasses() []storagev1.StorageClass {
// 	return c.storageClasses
// }

func Prepare(pd *aimlv1beta1.Pachyderm) PachydermComponents {
	components := getPachydermComponents()
	// set pachyderm resource as parent
	components.parent = pd
	return components
}
