/*
Copyright 2021 Pachyderm.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"sort"

	"github.com/creasty/defaults"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"golang.org/x/mod/semver"
)

// log is for logging in this package.
var pachydermlog = logf.Log.WithName("pachyderm-resource")

// SetupWebhookWithManager setups the webhook
func (r *Pachyderm) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-aiml-pachyderm-com-v1beta1-pachyderm,mutating=true,failurePolicy=fail,sideEffects=None,groups=aiml.pachyderm.com,resources=pachyderms,verbs=create;update,versions=v1beta1,name=mpachyderm.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Pachyderm{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Pachyderm) Default() {
	pachydermlog.Info("default", "name", r.Name)

	if r.Spec.Pachd.Storage.Backend == AmazonStorageBackend {
		if err := defaults.Set(r.Spec.Pachd.Storage.Amazon); err != nil {
			fmt.Println("err:", err.Error())
		}
	}

	if r.Spec.Version == "" {
		r.Spec.Version = getDefaultVersion()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-aiml-pachyderm-com-v1beta1-pachyderm,mutating=false,failurePolicy=fail,sideEffects=None,groups=aiml.pachyderm.com,resources=pachyderms,verbs=create;update,versions=v1beta1,name=vpachyderm.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Pachyderm{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Pachyderm) ValidateCreate() error {
	pachydermlog.Info("validate create", "name", r.Name)

	if r.isUsingGCS() && r.Spec.Pachd.Storage.Google.CredentialSecret == "" {
		return errors.New("spec.pachd.storage.google.credentialSecret can not be empty")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Pachyderm) ValidateUpdate(old runtime.Object) error {
	pachydermlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Pachyderm) ValidateDelete() error {
	pachydermlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// returns true if Pachd storage is using Google Container storage
func (r *Pachyderm) isUsingGCS() bool {
	return r.Spec.Pachd.Storage.Google != nil && r.Spec.Pachd.Storage.Backend == GoogleStorageBackend
}

func isContainer() bool {
	fs, err := os.Stat("/run/secrets/kubernetes.io/serviceaccount")
	return (fs.IsDir() && err == nil)
}

func getVersions() ([]string, error) {
	var versionsDir string = "/charts"
	if !isContainer() {
		return []string{}, errors.New("not running in container")

	}

	files, err := ioutil.ReadDir(versionsDir)
	if err != nil {
		fmt.Println("error:", err.Error())
	}

	versions := []string{}
	for _, f := range files {
		if f.IsDir() {
			if version := fmt.Sprintf("v%s", f.Name()); semver.IsValid(version) {
				versions = append(versions, version)
			}
		}
	}

	return versions, nil
}

// getDefaultVersion returns the newest Pachyderm version based on semver version
func getDefaultVersion() string {
	versions, err := getVersions()
	if err != nil {
		return ""
	}

	sort.Slice(versions, func(i, j int) bool {
		return (semver.Compare(versions[i], versions[j]) == -1)
	})

	return versions[len(versions)-1]
}

func (r *Pachyderm) SetGoogleCredentials(credentials []byte) {
	r.Spec.Pachd.Storage.Google.CredentialsData = credentials
}

// IsDeleted returns true if the pachyderm resource
// has been marked for deletion
func (r *Pachyderm) IsDeleted() bool {
	return r.ObjectMeta.DeletionTimestamp != nil
}

func (r *Pachyderm) DeployPostgres() bool {
	return !r.Spec.Postgres.Disable
}
