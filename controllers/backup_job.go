package controllers

import (
	"fmt"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func backupJob(pde *aimlv1beta1.PachydermExport) *batchv1.Job {
	labels := map[string]string{
		"owned-by":   pde.Name,
		"managed-by": "pachyderm-operator",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupJobName(pde),
			Namespace: pde.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "helper",
							Image:           "quay.io/eochieng/pachyderm-backup-helper:latest",
							Env:             []corev1.EnvVar{},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
}

func backupJobName(pde *aimlv1beta1.PachydermExport) string {
	return fmt.Sprintf("%s-job", pde.Name)
}
