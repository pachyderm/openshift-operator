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
	"reflect"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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

	// TODO(user): fill in your defaulting logic.
	if err := r.setDefaults(); err != nil {
		pachydermlog.Error(err, "failed setting defaults")
	}
}

func (r *Pachyderm) setDefaults() error {
	val := reflect.ValueOf(r).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		if defaultVal := typ.Field(i).Tag.Get("default"); defaultVal != "-" {
			if err := setField(val.Field(i), defaultVal); err != nil {
				return err
			}
		}
	}
	return nil
}

func setField(field reflect.Value, defaultVal string) error {
	switch field.Kind() {
	case reflect.Int32:
		if val, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(int32(val)).Convert(field.Type()))
		}
	case reflect.String:
		field.Set(reflect.ValueOf(defaultVal).Convert(field.Type()))
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-aiml-pachyderm-com-v1beta1-pachyderm,mutating=false,failurePolicy=fail,sideEffects=None,groups=aiml.pachyderm.com,resources=pachyderms,verbs=create;update,versions=v1beta1,name=vpachyderm.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Pachyderm{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Pachyderm) ValidateCreate() error {
	pachydermlog.Info("validate create", "name", r.Name)

	if r.Spec.Pachd.Storage.AmazonStorage != nil {
		r.prepareAmazonStorage()
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

func (r *Pachyderm) prepareAmazonStorage() {
	v := reflect.ValueOf(r.Spec.Pachd.Storage.AmazonStorage).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && !field.IsZero() {
			fieldVal := encodeString(field.Interface().(string))
			field.Set(reflect.ValueOf(fieldVal).Convert(field.Type()))
		}
	}
}

func encodeString(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}
