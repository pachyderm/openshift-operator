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
	"encoding/base64"
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

	if r.Spec.Pachd.Storage.Backend == "AMAZON" {
		r.prepareAmazonStorage()
	}

	// if backend is "local" but, spec.pachd.storage.local is nil
	// populate the hostPath
	if r.Spec.Pachd.Storage.Backend == "LOCAL" {
		r.prepareLocalStorage()
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
	return r.Spec.Pachd.Storage.Google != nil && r.Spec.Pachd.Storage.Backend == "GOOGLE"
}

func (r *Pachyderm) prepareLocalStorage() {
	if r.Spec.Pachd.Storage.Local == nil {
		r.Spec.Pachd.Storage.Local = &LocalStorageOptions{}
		if err := defaults.Set(r.Spec.Pachd.Storage.Local); err != nil {
			fmt.Println("err:", err.Error())
		}
	}
}

func (r *Pachyderm) prepareAmazonStorage() {
	if r.Spec.Pachd.Storage.Amazon != nil {
		// apply defaults
		if err := defaults.Set(r.Spec.Pachd.Storage.Amazon); err != nil {
			fmt.Println("err:", err.Error())
		}

		if r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution != "" {
			r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution = encodeString(r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution, false)
		}
		if r.Spec.Pachd.Storage.Amazon.IAMRole != "" {
			r.Spec.Pachd.Storage.Amazon.IAMRole = encodeString(r.Spec.Pachd.Storage.Amazon.IAMRole, false)
		}
		if r.Spec.Pachd.Storage.Amazon.ID != "" {
			r.Spec.Pachd.Storage.Amazon.ID = encodeString(r.Spec.Pachd.Storage.Amazon.ID, true)
		}
		if r.Spec.Pachd.Storage.Amazon.Secret != "" {
			r.Spec.Pachd.Storage.Amazon.Secret = encodeString(r.Spec.Pachd.Storage.Amazon.Secret, true)
		}
		if r.Spec.Pachd.Storage.Amazon.Token != "" {
			r.Spec.Pachd.Storage.Amazon.Token = encodeString(r.Spec.Pachd.Storage.Amazon.Token, true)
		}
		if r.Spec.Pachd.Storage.Amazon.UploadACL != "" {
			r.Spec.Pachd.Storage.Amazon.UploadACL = encodeString(r.Spec.Pachd.Storage.Amazon.UploadACL, false)
		}
	}
}

// checks if input string is base64 encoded.
// If yes, input variable is returned
// else, base64 encoded input is returned
func encodeString(input string, override bool) string {
	if !override {
		if IsBase64Encoded(input) {
			return input
		}
	}

	if encodeCount(input) == 2 {
		return input
	}

	return base64.StdEncoding.EncodeToString([]byte(input))
}

// IsBase64Encoded checks if user input is already base64 encoded
func IsBase64Encoded(input string) bool {
	if _, err := base64.StdEncoding.DecodeString(input); err == nil {
		return true
	}
	return false
}

func encodeCount(input string) int {
	var count int
	var err error
	result := input

	for count = 0; err == nil; count++ {
		out, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			break
		}
		result = string(out)
	}
	return count
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
