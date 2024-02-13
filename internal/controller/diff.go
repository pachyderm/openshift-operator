package controller

import (
	"reflect"

	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
)

func serviceChanged(current, new *corev1.Service) bool {
	tempSvc := *current

	if err := mergo.Merge(new.ObjectMeta, current.ObjectMeta); err != nil {
		return false
	}

	if !reflect.DeepEqual(new.Spec.Ports, current.Spec.Ports) {
		current.Spec.Ports = new.Spec.Ports
	}

	return !reflect.DeepEqual(tempSvc, *current)
}
