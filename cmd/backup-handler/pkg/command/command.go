package command

import (
	"bytes"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

func kubeConfig() (*rest.Config, error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		cfg := filepath.Join("~/", ".kube", "config")
		if os.Getenv("KUBECONFIG") != "" {
			cfg = os.Getenv("KUBECONFIG")
		}
		return clientcmd.BuildConfigFromFlags("", cfg)
	}

	return conf, nil
}

func kubeClient(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

// ExecResponse provides a mechanism for an
// executed command to return output following the
// completion of a command
type ExecResponse struct {
	stdout string
	stderr string
}

type ExecOptions struct {
	Pod       string
	Container string
	Namespace string
	Command   []string
}

func (r *ExecResponse) Output() string {
	return r.stdout
}

func (r *ExecResponse) Error() string {
	return r.stderr
}

// ExecuteCommand takes a command and runs it on a specific pod
func ExecuteCommand(options ExecOptions) (*ExecResponse, error) {
	config, err := kubeConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubeClient(config)
	if err != nil {
		return nil, err
	}

	request := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.Pod).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.Container)
	request.VersionedParams(
		&corev1.PodExecOptions{
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
			Container: options.Container,
			Command:   options.Command,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL())
	if err != nil {
		return nil, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(
		remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: &stdout,
			Stderr: &stderr,
		},
	)
	if err != nil {
		return nil, err
	}

	return &ExecResponse{
		stdout: stdout.String(),
		stderr: stderr.String(),
	}, nil
}
