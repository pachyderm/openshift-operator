module github.com/pachyderm/openshift-operator

go 1.16

require (
	github.com/Azure/go-autorest/autorest/adal v0.9.16 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/aws/aws-sdk-go v1.44.26 // indirect
	github.com/creasty/defaults v1.5.1
	github.com/go-logr/logr v1.2.3
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/imdario/mergo v0.3.13
	github.com/lib/pq v1.10.4
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.13 // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/opdev/backup-handler v0.0.0-20220602073855-51dc4aa0f95d
	go.uber.org/atomic v1.9.0 // indirect
	goa.design/goa/v3 v3.7.5 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3
	golang.org/x/net v0.0.0-20220531201128-c960675eff93 // indirect
	golang.org/x/oauth2 v0.0.0-20220524215830-622c5d57e401 // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.1.0 // indirect
	helm.sh/helm/v3 v3.9.0
	k8s.io/api v0.24.0
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/kube-openapi v0.0.0-20220413171646-5e7f5fdc6da6 // indirect
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/json v0.0.0-20220525155127-227cbc7cc124 // indirect
)

replace github.com/googleapis/gnostic => github.com/google/gnostic v0.5.5
