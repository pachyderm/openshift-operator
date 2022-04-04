module github.com/pachyderm/openshift-operator

go 1.16

require (
	github.com/creasty/defaults v1.5.1
	github.com/go-logr/logr v1.2.0
	github.com/imdario/mergo v0.3.12
	github.com/lib/pq v1.10.4
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/opdev/backup-handler v0.0.0-20220404164209-e68569966e09
	golang.org/x/mod v0.5.0
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	helm.sh/helm/v3 v3.8.0
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.23.3
	sigs.k8s.io/controller-runtime v0.11.1
)
