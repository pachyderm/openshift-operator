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
	"regexp"
	"strings"

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

	// encode storage info
	r.encodeStorageSecrets()

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

func (r *Pachyderm) prepareAmazonStorage() {
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

// DecodeAmazonStorage function allows controllers to get the decoded user input
// before using them as input for helm charts
func (r *Pachyderm) decodeAmazonStorage() {
	if r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution != "" {
		r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution = decodeString(r.Spec.Pachd.Storage.Amazon.CloudFrontDistribution)
	}
	if r.Spec.Pachd.Storage.Amazon.IAMRole != "" {
		r.Spec.Pachd.Storage.Amazon.IAMRole = decodeString(r.Spec.Pachd.Storage.Amazon.IAMRole)
	}
	if r.Spec.Pachd.Storage.Amazon.ID != "" {
		r.Spec.Pachd.Storage.Amazon.ID = decodeString(r.Spec.Pachd.Storage.Amazon.ID)
	}
	if r.Spec.Pachd.Storage.Amazon.Secret != "" {
		r.Spec.Pachd.Storage.Amazon.Secret = decodeString(r.Spec.Pachd.Storage.Amazon.Secret)
	}
	if r.Spec.Pachd.Storage.Amazon.Token != "" {
		r.Spec.Pachd.Storage.Amazon.Token = decodeString(r.Spec.Pachd.Storage.Amazon.Token)
	}
	if r.Spec.Pachd.Storage.Amazon.UploadACL != "" {
		r.Spec.Pachd.Storage.Amazon.UploadACL = decodeString(r.Spec.Pachd.Storage.Amazon.UploadACL)
	}
}

func isInputEncoded(input string) bool {
	output, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return false
	}
	re := regexp.MustCompile(`^\[#\]`)
	return re.Match(output)
}

// checks if input string is base64 encoded.
// else, base64 encoded input is returned
func encodeString(input string, override bool) string {
	if isInputEncoded(input) {
		return input
	}
	temp := []byte(fmt.Sprintf("[#].%s", input))
	return base64.StdEncoding.EncodeToString(temp)
}

func decodeString(input string) string {
	if isInputEncoded(input) {
		payload, _ := base64.StdEncoding.DecodeString(input)
		return strings.Split(string(payload), "].")[1]
	}

	return input
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

// TODO: encode minio, azure storage credentials
func (r *Pachyderm) encodeStorageSecrets() {
	switch r.Spec.Pachd.Storage.Backend {
	case "AMAZON":
		if r.Spec.Pachd.Storage.Amazon.isSet() {
			r.prepareAmazonStorage()
		}
	case "MICROSOFT":
		if r.Spec.Pachd.Storage.Microsoft.isSet() {
			fmt.Println("microsoft azure blob storage")
		}
	case "GOOGLE":
		if r.Spec.Pachd.Storage.Google.isSet() {
			fmt.Println("google container storage")
		}
	case "MINIO":
		if r.Spec.Pachd.Storage.Minio.isSet() {
			fmt.Println("minio object storage")
		}
	}
}

func (r *Pachyderm) DecodeStorageInput() {
	switch r.Spec.Pachd.Storage.Backend {
	case "AMAZON":
		r.decodeAmazonStorage()
	case "MICROSOFT":
	case "GOOGLE":
	case "MINIO":
	}
}

func (s *AmazonStorageOptions) isSet() bool {
	return s != nil
}

func (s *MicrosoftStorageOptions) isSet() bool {
	return s != nil
}

func (s *GoogleStorageOptions) isSet() bool {
	return s != nil
}

func (s *MinioStorageOptions) isSet() bool {
	return s != nil
}
