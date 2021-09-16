module github.com/pachyderm/openshift-operator

go 1.16

require (
	github.com/creasty/defaults v1.5.1
	github.com/go-logr/logr v0.4.0
	github.com/imdario/mergo v0.3.12
	github.com/lib/pq v1.10.3 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	golang.org/x/mod v0.4.2
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.2
)
