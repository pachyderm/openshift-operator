package generators

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func (c *PachydermComponents) dashEnvironmentVars() []corev1.EnvVar {
	envs := []corev1.EnvVar{
		{
			Name:  "ISSUER_URI",
			Value: "",
		},
		{
			Name:  "OAUTH_REDIRECT_URI",
			Value: "",
		},
		{
			Name:  "OAUTH_CLIENT_ID",
			Value: "",
		},
		{
			Name:  "OAUTH_CLIENT_SECRET",
			Value: "",
		},
		{
			Name:  "GRAPHQL_PORT",
			Value: "",
		},
		{
			Name:  "OAUTH_PACHD_CLIENT_ID",
			Value: "",
		},
		{
			Name:  "PACHD_ADDRESS",
			Value: c.pachdAddess(),
		},
	}

	return envs
}

func (c *PachydermComponents) pachdAddess() string {
	return fmt.Sprintf("pachd-peer.%s.svc.cluster.local:30653", c.Namespace())
}

func (c *PachydermComponents) Namespace() string {
	return c.Pachyderm().Namespace
}
