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
	"strings"

	goyaml "github.com/go-yaml/yaml"
	aimlv1beta1 "github.com/opdev/pachyderm-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// PachydermComponents is a structure that contains a slice of
// all the Kubernetes resources that make up a Pachyderm deployment
type PachydermComponents struct {
	// auto-detected worker image name
	workerImageName string
	// auto-detected worker sidecar image name
	workerSidecarName   string
	gcsCredentials      []byte
	pachyderm           *aimlv1beta1.Pachyderm
	dashDeploy          *appsv1.Deployment
	pachdDeploy         *appsv1.Deployment
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
}

func (c *PachydermComponents) SetGoogleCredentials(credentials []byte) {
	c.gcsCredentials = credentials
}

func (c *PachydermComponents) getGCSCredentials() []byte {
	return c.gcsCredentials
}

// PachydermError defines custom error
// type used by the operator
type PachydermError string

func (e PachydermError) Error() string {
	return string(e)
}

func getManifestPath(version string) string {
	manifestPath := filepath.Join("/", "manifests", version[1:], "manifests.yaml")

	// Check operator is not running in Openshift
	if !isKubernetes() {
		wd, err := os.Getwd()
		if err != nil {
			return manifestPath
		}
		manifestPath = filepath.Join(wd, "hack", "manifests", version, "manifests.yaml")
	}
	return manifestPath
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

func loadManifests(version string) ([][]byte, error) {
	var objects [][]byte

	// Read manifests from file
	manifestPath := getManifestPath(version)
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

func getPachydermComponents(pd *aimlv1beta1.Pachyderm) *PachydermComponents {
	components := &PachydermComponents{}

	manifests, err := loadManifests(pd.Spec.Version)
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
				fmt.Println("error parsing statefulset.", err.Error())
			}
		case "Pod":
			if err := components.parsePod(obj, pd.Namespace); err != nil {
				fmt.Println("error parsing pod.", err.Error())
			}
		case "ServiceAccount":
			sa := &corev1.ServiceAccount{}
			if err := toTypedResource(obj, sa); err != nil {
				fmt.Println("error converting to service account.", err.Error())
			}
			sa.Namespace = pd.Namespace
			components.ServiceAccounts = append(components.ServiceAccounts, sa)
		case "Secret":
			secret := &corev1.Secret{}
			if err := toTypedResource(obj, secret); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			secret.Namespace = pd.Namespace
			components.secrets = append(components.secrets, secret)
		case "ConfigMap":
			cm := &corev1.ConfigMap{}
			if err := toTypedResource(obj, cm); err != nil {
				fmt.Println("error converting to config map.", err.Error())
			}
			cm.Namespace = pd.Namespace
			components.configMaps = append(components.configMaps, cm)
		case "StorageClass":
			sc := &storagev1.StorageClass{}
			if err := toTypedResource(obj, sc); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			components.storageClasses = append(components.storageClasses, sc)
		case "ClusterRole":
			clusterrole := &rbacv1.ClusterRole{}
			if err := toTypedResource(obj, clusterrole); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.ClusterRoles = append(components.ClusterRoles, clusterrole)
		case "ClusterRoleBinding":
			clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
			if err := toTypedResource(obj, clusterRoleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			for i := range clusterRoleBinding.Subjects {
				clusterRoleBinding.Subjects[i].Namespace = pd.Namespace
			}
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, clusterRoleBinding)
		case "Role":
			role := &rbacv1.Role{}
			if err := toTypedResource(obj, role); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			role.Namespace = pd.Namespace
			components.Roles = append(components.Roles, role)
		case "RoleBinding":
			roleBinding := &rbacv1.RoleBinding{}
			if err := toTypedResource(obj, roleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			roleBinding.Namespace = pd.Namespace
			components.RoleBindings = append(components.RoleBindings, roleBinding)
		case "Service":
			svc := &corev1.Service{}
			if err := toTypedResource(obj, svc); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			svc.Namespace = pd.Namespace
			components.Services = append(components.Services, svc)
		}
	}

	components.readDefaultImages()

	return components
}

func (c *PachydermComponents) readDefaultImages() {
	var pachdEnvs []corev1.EnvVar
	for _, container := range c.pachdDeploy.Spec.Template.Spec.Containers {
		if container.Name == "pachd" {
			pachdEnvs = container.Env
		}
	}

	if len(pachdEnvs) > 0 {
		for _, env := range pachdEnvs {
			if env.Name == "WORKER_IMAGE" {
				c.workerImageName = env.Value
			}

			if env.Name == "WORKER_SIDECAR_IMAGE" {
				c.workerSidecarName = env.Value
			}
		}
	}
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
		switch sts.Name {
		case "etcd":
			c.etcdStatefulSet = &sts
		case "postgres":
			c.postgreStatefulSet = &sts
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
func (c *PachydermComponents) Pachyderm() *aimlv1beta1.Pachyderm {
	return c.pachyderm
}

// StorageClass returns a new etcd storage class
// if an existing one is not used or provided
func (c *PachydermComponents) StorageClasses() []*storagev1.StorageClass {
	var storageClasses []*storagev1.StorageClass
	pd := c.pachyderm
	var allowExpansion bool = true

	// if storage class is not provided,
	// create a new storage class
	for _, sc := range c.storageClasses {
		sc.AllowVolumeExpansion = &allowExpansion

		if !reflect.DeepEqual(pd.Spec.Pachd, aimlv1beta1.PachdOptions{}) {
			// check if the etcd-storage-class and postgresql-storage-class
			// need to be created
			if pd.Spec.Etcd.StorageClass == "" ||
				pd.Spec.Postgres.StorageClass == "" {
				storageClassProvisioner(pd, sc)
				storageClasses = append(storageClasses, sc)
			}

		}
	}

	return storageClasses
}

func storageClassProvisioner(pd *aimlv1beta1.Pachyderm, sc *storagev1.StorageClass) {
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

// Secrets returns secrets used by the pachyderm resource
func (c *PachydermComponents) Secrets() []*corev1.Secret {
	pd := c.pachyderm

	for _, secret := range c.secrets {
		if secret.Name == "pachyderm-storage-secret" {
			c.setupStorageSecret(secret)
		}

		if secret.Name == "pachd-tls-cert" {
			setupPachdTLSSecret(secret, pd)
		}
	}

	return c.secrets
}

func (c *PachydermComponents) ConfgigMaps() []*corev1.ConfigMap {
	for _, cm := range c.configMaps {
		cm.Namespace = c.pachyderm.Namespace
	}
	return c.configMaps
}

func (c *PachydermComponents) setupStorageSecret(secret *corev1.Secret) {
	pd := c.pachyderm
	if pd.Spec.Pachd.Storage.Backend == "local" {
		secret.Data = map[string][]byte{}
	}

	if pd.Spec.Pachd.Storage.Backend == "amazon" {
		secret.Data = map[string][]byte{
			"AMAZON_BUCKET":       toBytes(pd.Spec.Pachd.Storage.Amazon.Bucket),
			"AMAZON_SECRET":       toBytes(pd.Spec.Pachd.Storage.Amazon.Secret),
			"AMAZON_REGION":       toBytes(pd.Spec.Pachd.Storage.Amazon.Region),
			"AMAZON_TOKEN":        toBytes(pd.Spec.Pachd.Storage.Amazon.Token),
			"AMAZON_ID":           toBytes(pd.Spec.Pachd.Storage.Amazon.ID),
			"AMAZON_DISTRIBUTION": toBytes(pd.Spec.Pachd.Storage.Amazon.CloudFrontDistribution),
			"CUSTOM_ENDPOINT":     toBytes(pd.Spec.Pachd.Storage.Amazon.CustomEndpoint),
			"DISABLE_SSL":         toBytes(fmt.Sprintf("%t", pd.Spec.Pachd.Storage.Amazon.DisableSSL)),
			"OBJ_LOG_OPTS":        toBytes(pd.Spec.Pachd.Storage.Amazon.LogOptions),
			"MAX_UPLOAD_PARTS":    toBytes(fmt.Sprintf("%d", pd.Spec.Pachd.Storage.Amazon.MaxUploadParts)),
			"NO_VERIFY_SSL":       toBytes(fmt.Sprintf("%t", pd.Spec.Pachd.Storage.Amazon.VerifySSL)),
			"PART_SIZE":           toBytes(fmt.Sprintf("%d", pd.Spec.Pachd.Storage.Amazon.PartSize)),
			"RETRIES":             toBytes(fmt.Sprintf("%d", pd.Spec.Pachd.Storage.Amazon.Retries)),
			"REVERSE":             toBytes(fmt.Sprintf("%t", *pd.Spec.Pachd.Storage.Amazon.Reverse)),
			"TIMEOUT":             toBytes(pd.Spec.Pachd.Storage.Amazon.Timeout),
			"UPLOAD_ACL":          toBytes(pd.Spec.Pachd.Storage.Amazon.UploadACL),
		}
	}

	if pd.Spec.Pachd.Storage.Backend == "minio" {
		secret.Data = map[string][]byte{
			"MINIO_BUCKET":    toBytes(pd.Spec.Pachd.Storage.Minio.Bucket),
			"MINIO_ENDPOINT":  toBytes(pd.Spec.Pachd.Storage.Minio.Endpoint),
			"MINIO_ID":        toBytes(pd.Spec.Pachd.Storage.Minio.ID),
			"MINIO_SECRET":    toBytes(pd.Spec.Pachd.Storage.Minio.Secret),
			"MINIO_SECURE":    toBytes(pd.Spec.Pachd.Storage.Minio.Secure),
			"MINIO_SIGNATURE": toBytes(pd.Spec.Pachd.Storage.Minio.Signature),
		}
	}

	if pd.Spec.Pachd.Storage.Backend == "google" {
		secret.Data = map[string][]byte{
			"GOOGLE_BUCKET": toBytes(pd.Spec.Pachd.Storage.Google.Bucket),
			"GOOGLE_CRED":   c.getGCSCredentials(),
		}
	}

	if pd.Spec.Pachd.Storage.Backend == "microsoft" {
		secret.Data = map[string][]byte{
			"MICROSOFT_CONTAINER": toBytes(pd.Spec.Pachd.Storage.Microsoft.Container),
			"MICROSOFT_SECRET":    toBytes(pd.Spec.Pachd.Storage.Microsoft.Secret),
			"MICROSOFT_ID":        toBytes(pd.Spec.Pachd.Storage.Microsoft.ID),
		}
	}
}

// accepts string and returns a slice of type bytes
func toBytes(value string) []byte {
	if aimlv1beta1.IsBase64Encoded(value) {
		out, _ := base64.StdEncoding.DecodeString(value)
		return out
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
	for i := range c.etcdStatefulSet.Spec.VolumeClaimTemplates {
		c.etcdStatefulSet.Spec.VolumeClaimTemplates[i].Namespace = pd.Namespace
	}

	return c.etcdStatefulSet
}

// PachdDeployment returns the pachd deployment resource
func (c *PachydermComponents) PachdDeployment() *appsv1.Deployment {
	deploy := c.pachdDeploy
	pachyderm := c.Pachyderm()

	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == "pachd" {
			deploy.Spec.Template.Spec.Containers[i].Env = c.pachdEnvVarirables()

			if pachyderm.Spec.Pachd.Image != nil {
				pachdImage := strings.Split(container.Image, ":")
				if pachyderm.Spec.Pachd.Image.Repository != "" {
					pachdImage[0] = pachyderm.Spec.Pachd.Image.Repository
				}
				if pachyderm.Spec.Pachd.Image.ImageTag != "" {
					pachdImage[1] = pachyderm.Spec.Pachd.Image.ImageTag
				}
				deploy.Spec.Template.Spec.Containers[i].Image = strings.Join(pachdImage, ":")
			}
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

	for i, container := range c.dashDeploy.Spec.Template.Spec.Containers {
		if container.Name == "dash" {
			c.dashDeploy.Spec.Template.Spec.Containers[i].Env = c.dashEnvironmentVars()
		}
	}
	return c.dashDeploy
}

// PostgreStatefulset returns the postgresql statefulset resource
func (c *PachydermComponents) PostgreStatefulset() *appsv1.StatefulSet {
	return c.postgreStatefulSet
}

// Prepare takes a pachyderm custom resource and returns
// child resources based on the pachyderm custom resource
func Prepare(pd *aimlv1beta1.Pachyderm) *PachydermComponents {
	components := getPachydermComponents(pd)
	// set pachyderm resource as parent
	components.pachyderm = pd
	return components
}
