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

func (c *PachydermCluster) SetGoogleCredentials(credentials []byte) {
	c.gcsCredentials = credentials
}

// PachydermError defines custom error
// type used by the operator
type PachydermError string

func (e PachydermError) Error() string {
	return string(e)
}

func getPachydermCluster(pd *aimlv1beta1.Pachyderm) (*PachydermCluster, error) {
	manifests, err := loadPachydermTemplates(pd)
	if err != nil {
		return nil, err
	}

	components := &PachydermCluster{}
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
			if err := components.parseDeployment(obj); err != nil {
				fmt.Println("error parsing deployment.", err.Error())
			}
		case "StatefulSet":
			if err := components.parseStatefulSet(obj); err != nil {
				fmt.Println("error parsing statefulset.", err.Error())
			}
		case "Pod":
			if err := components.parsePod(obj); err != nil {
				fmt.Println("error parsing pod.", err.Error())
			}
		case "ServiceAccount":
			sa := &corev1.ServiceAccount{}
			if err := toTypedResource(obj, sa); err != nil {
				fmt.Println("error converting to service account.", err.Error())
			}
			components.ServiceAccounts = append(components.ServiceAccounts, sa)
		case "Secret":
			secret := &corev1.Secret{}
			if err := toTypedResource(obj, secret); err != nil {
				fmt.Println("error converting to secret.", err.Error())
			}
			components.secrets = append(components.secrets, secret)
		case "ConfigMap":
			cm := &corev1.ConfigMap{}
			if err := toTypedResource(obj, cm); err != nil {
				fmt.Println("error converting to config map.", err.Error())
			}
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
			components.ClusterRoleBindings = append(components.ClusterRoleBindings, clusterRoleBinding)
		case "Role":
			role := &rbacv1.Role{}
			if err := toTypedResource(obj, role); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.Roles = append(components.Roles, role)
		case "RoleBinding":
			roleBinding := &rbacv1.RoleBinding{}
			if err := toTypedResource(obj, roleBinding); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.RoleBindings = append(components.RoleBindings, roleBinding)
		case "Service":
			svc := &corev1.Service{}
			if err := toTypedResource(obj, svc); err != nil {
				fmt.Println("error converting to cluster role.", err.Error())
			}
			components.Services = append(components.Services, svc)
		}
	}

	return components, nil
}

func toTypedResource(unstructured *unstructured.Unstructured, object interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, object)
}

func (c *PachydermCluster) parseDeployment(obj *unstructured.Unstructured) error {
	var deployment appsv1.Deployment
	if err := toTypedResource(obj, &deployment); err != nil {
		return err
	}

	if !reflect.DeepEqual(deployment, appsv1.Deployment{}) {
		switch deployment.Name {
		case "dash":
			c.dashDeploy = &deployment
		case "pachd":
			c.pachdDeploy = &deployment
		}
	}

	return nil
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

func (c *PachydermCluster) parsePod(obj *unstructured.Unstructured) error {
	var pod corev1.Pod
	if err := toTypedResource(obj, &pod); err != nil {
		return err
	}
	c.Pod = &pod

	return nil
}

// Parent returns the pachyderm resource used to configure components
func (c *PachydermCluster) Pachyderm() *aimlv1beta1.Pachyderm {
	return c.pachyderm
}

// StorageClass returns a new etcd storage class
// if an existing one is not used or provided
func (c *PachydermCluster) StorageClasses() []*storagev1.StorageClass {
	return c.storageClasses
}

// Secrets returns secrets used by the pachyderm resource
func (c *PachydermCluster) Secrets() []*corev1.Secret {
	return c.secrets
}

func (c *PachydermCluster) ConfgigMaps() []*corev1.ConfigMap {
	return c.configMaps
}

// EtcdStatefulSet returns the etcd statefulset resource
func (c *PachydermCluster) EtcdStatefulSet() *appsv1.StatefulSet {
	return c.etcdStatefulSet
}

// TODO: update labels to use recommended k8s labels
// PachdDeployment returns the pachd deployment resource
func (c *PachydermCluster) PachdDeployment() *appsv1.Deployment {
	return c.pachdDeploy
}

// DashDeployment returns the dash deployment resource
func (c *PachydermCluster) DashDeployment() *appsv1.Deployment {
	return c.dashDeploy
}

// PostgreStatefulset returns the postgresql statefulset resource
func (c *PachydermCluster) PostgreStatefulset() *appsv1.StatefulSet {
	return c.postgreStatefulSet
}

// PrepareCluster takes a pachyderm custom resource and returns
// child resources based on the pachyderm custom resource
func PrepareCluster(pd *aimlv1beta1.Pachyderm) (*PachydermCluster, error) {
	cluster, err := getPachydermCluster(pd)
	if err != nil {
		return nil, err
	}

	// set pachyderm resource as parent
	cluster.pachyderm = pd
	return cluster, nil
}
