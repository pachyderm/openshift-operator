package generators

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

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
	pachyderm           *aimlv1beta1.Pachyderm
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
	secrets             []*corev1.Secret
	storageClass        storagev1.StorageClass
}

func getPachydermComponents(pd *aimlv1beta1.Pachyderm) PachydermComponents {
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
			if err := components.parseDeployment(obj, pd.Namespace); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
		case "StatefulSet":
			if err := components.parseStatefulSet(obj, pd.Namespace); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
		case "Pod":
			if err := components.parsePod(obj, pd.Namespace); err != nil {
				fmt.Println("error parsing pod.", err.Error())
			}
		case "ServiceAccount":
			var sa corev1.ServiceAccount
			if err := toTypedResource(obj, &sa); err != nil {
				fmt.Println("error converting to service account.", err.Error())
			}
			sa.Namespace = pd.Namespace
			components.ServiceAccounts = append(components.ServiceAccounts, sa)
		case "Secret":
			var secret corev1.Secret
			if err := toTypedResource(obj, &secret); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			secret.Namespace = pd.Namespace
			components.secrets = append(components.secrets, &secret)
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
			role.Namespace = pd.Namespace
			components.Roles = append(components.Roles, role)
		case "RoleBinding":
			var roleBinding rbacv1.RoleBinding
			if err := toTypedResource(obj, &roleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			roleBinding.Namespace = pd.Namespace
			components.RoleBindings = append(components.RoleBindings, roleBinding)
		case "Service":
			var svc corev1.Service
			if err := toTypedResource(obj, &svc); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			svc.Namespace = pd.Namespace
			components.Services = append(components.Services, svc)
		}
	}

	return components
}

func toTypedResource(unstructured *unstructured.Unstructured, object interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, object)
}

func (c *PachydermComponents) parseDeployment(obj *unstructured.Unstructured, namespace string) error {
	var deployment appsv1.Deployment
	if err := toTypedResource(obj, &deployment); err != nil {
		return err
	}

	if !reflect.DeepEqual(deployment, appsv1.Deployment{}) {
		deployment.Namespace = namespace

		switch deployment.Name {
		case "dash":
			c.dashDeploy = &deployment
		case "pachd":
			c.pachdDeploy = &deployment
		}
	}

	return nil
}

func (c *PachydermComponents) parseStatefulSet(obj *unstructured.Unstructured, namespace string) error {
	var sts appsv1.StatefulSet
	if err := toTypedResource(obj, &sts); err != nil {
		fmt.Println("error converting to statefulset.", err.Error())
	}

	if !reflect.DeepEqual(sts, appsv1.StatefulSet{}) {
		sts.Namespace = namespace

		if sts.Name == "etcd" {
			c.etcdStatefulSet = &sts
		}
	}

	return nil
}

func (c *PachydermComponents) parsePod(obj *unstructured.Unstructured, namespace string) error {
	var pod corev1.Pod
	if err := toTypedResource(obj, &pod); err != nil {
		return err
	}
	pod.Namespace = namespace
	c.Pod = &pod

	return nil
}

// Parent returns the pachyderm resource used to configure components
func (c *PachydermComponents) Parent() *aimlv1beta1.Pachyderm {
	return c.pachyderm
}

// EtcdStorageClassName return aname of storage class to be used by etcd
func EtcdStorageClassName(pd *aimlv1beta1.Pachyderm) string {
	var storageClass string = "etcd-storage-class"
	if pd.Spec.Etcd.StorageClass != "" {
		storageClass = pd.Spec.Etcd.StorageClass
	}
	return storageClass
}

// StorageClass returns a new etcd storage class
// if an existing one is not used or provided
func (c *PachydermComponents) StorageClass() *storagev1.StorageClass {
	pd := c.pachyderm
	var allowExpansion bool = true

	if !reflect.DeepEqual(pd.Spec.Etcd, aimlv1beta1.EtcdOptions{}) &&
		pd.Spec.Etcd.StorageClass != "" {
		return nil
	}

	// if storage class is not provided,
	// create a new storage class
	sc := &c.storageClass
	sc.AllowVolumeExpansion = &allowExpansion

	if !reflect.DeepEqual(pd.Spec.Pachd, aimlv1beta1.PachdOptions{}) {
		switch pd.Spec.Pachd.Storage.Backend {
		case "google":
			sc.Provisioner = "kubernetes.io/gce-pd"
			sc.Parameters = map[string]string{
				"type": "pd-ssd",
			}
		case "amazon":
			// https://docs.aws.amazon.com/eks/latest/userguide/ebs-csi.html
			sc.Provisioner = "ebs.csi.aws.com"
			sc.Parameters = map[string]string{
				"type": "gp3",
			}
		case "microsoft":
			sc.Provisioner = ""
			sc.Parameters = map[string]string{
				"type": "",
			}
		case "minio":
			sc.Provisioner = ""
			sc.Parameters = map[string]string{
				"type": "",
			}
		default:
			sc.Provisioner = "ebs.csi.aws.com"
			sc.Parameters = map[string]string{
				"type": "gp3",
			}
		}
	}

	return &c.storageClass
}

// Secrets returns secrets used by the pachyderm resource
func (c *PachydermComponents) Secrets() []*corev1.Secret {
	pd := c.pachyderm

	for _, secret := range c.secrets {
		if secret.Name == "pachyderm-storage-secret" {
			setupStorageSecret(secret, pd)
		}

		if secret.Name == "pachd-tls-cert" {
			setupPachdTLSSecret(secret, pd)
		}
	}

	return c.secrets
}

func setupStorageSecret(secret *corev1.Secret, pd *aimlv1beta1.Pachyderm) {
	data := secret.Data

	if pd.Spec.Pachd.Storage.Backend == "amazon" {
		if pd.Spec.Pachd.Storage.Amazon.Bucket != "" {
			data["amazon-bucket"] = toBytes(pd.Spec.Pachd.Storage.Amazon.Bucket)
		}

		if pd.Spec.Pachd.Storage.Amazon.Secret != "" {
			data["amazon-secret"] = toBytes(pd.Spec.Pachd.Storage.Amazon.Secret)
		}

		if pd.Spec.Pachd.Storage.Amazon.CustomEndpoint != "" {
			data["custom-endpoint"] = toBytes(pd.Spec.Pachd.Storage.Amazon.CustomEndpoint)
		}

		if pd.Spec.Pachd.Storage.Amazon.Region != "" {
			data["amazon-region"] = toBytes(pd.Spec.Pachd.Storage.Amazon.Region)
		}

		if pd.Spec.Pachd.Storage.Amazon.Token != "" {
			data["amazon-token"] = toBytes(pd.Spec.Pachd.Storage.Amazon.Token)
		}

		if pd.Spec.Pachd.Storage.Amazon.ID != "" {
			data["amazon-id"] = toBytes(pd.Spec.Pachd.Storage.Amazon.ID)
		}
	}

	if pd.Spec.Pachd.Storage.Backend == "local" {
		secret.Data = map[string][]byte{}
	}

}

// accepts string and returns a slice of type bytes
func toBytes(value string) []byte {
	if aimlv1beta1.IsBase64Encoded(value) {
		if out, err := base64.StdEncoding.DecodeString(value); err == nil {
			return out
		}
	}
	return []byte(value)
}

// generate self-signed TLS secret
func setupPachdTLSSecret(secret *corev1.Secret, pd *aimlv1beta1.Pachyderm) {
	rsaKey, err := newPrivateKeyRSA(keyBitSize)
	if err != nil {
		fmt.Println("error:", err.Error())
	}

	x509Cert, err := newClientCertificate(rsaKey, []string{"example.pachyderm.com"})
	if err != nil {
		fmt.Println("error:", err.Error())
	}

	secret.Data = map[string][]byte{
		"tls.crt": encodeCertificateToPEM(x509Cert),
		"tls.key": encodePrivateKeyToPEM(rsaKey),
	}
}

// EtcdStatefulSet returns the etcd statefulset resource
func (c *PachydermComponents) EtcdStatefulSet() *appsv1.StatefulSet {
	pd := c.pachyderm

	// set resource requests and limits
	if !reflect.DeepEqual(pd.Spec.Etcd, aimlv1beta1.EtcdOptions{}) {
		for _, container := range c.etcdStatefulSet.Spec.Template.Spec.Containers {
			if container.Name == "etcd" {
				if pd.Spec.Etcd.Resources != nil {
					container.Resources.Limits = pd.Spec.Etcd.Resources.Limits
					container.Resources.Requests = pd.Spec.Etcd.Resources.Requests
				}
			}
		}
	}

	// set etcd storage class
	for _, volumeClaim := range c.etcdStatefulSet.Spec.VolumeClaimTemplates {
		if volumeClaim.Name == "etcd-storage" {
			volumeClaim.Annotations["volume.beta.kubernetes.io/storage-class"] = EtcdStorageClassName(pd)
		}
	}

	return c.etcdStatefulSet
}

// PachdDeployment returns the pachd deployment resource
func (c *PachydermComponents) PachdDeployment() *appsv1.Deployment {
	deploy := c.pachdDeploy
	pachyderm := c.Parent()

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == "pachd" {
			deploy.Spec.Template.Spec.Containers[i].Env = pachdEnvVarirables(c.pachyderm)
		}
	}

	if pachyderm.Spec.Pachd.Storage.Backend == "local" {
		for _, volume := range deploy.Spec.Template.Spec.Volumes {
			if volume.Name == "pach-disk" {
				dirOrCreate := corev1.HostPathDirectoryOrCreate
				volume = corev1.Volume{
					Name: "pach-disk",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: filepath.Join(pachyderm.Spec.Pachd.Storage.Local.HostPath, "pachd"),
							Type: &dirOrCreate,
						},
					},
				}
			}
		}
	}

	return c.pachdDeploy
}

// DashDeployment returns the dash deployment resource
func (c *PachydermComponents) DashDeployment() *appsv1.Deployment {
	return c.dashDeploy
}

// Prepare takes a pachyderm custom resource and returns
// child resources based on the pachyderm custom resource
func Prepare(pd *aimlv1beta1.Pachyderm) PachydermComponents {
	components := getPachydermComponents(pd)
	// set pachyderm resource as parent
	components.pachyderm = pd
	return components
}
