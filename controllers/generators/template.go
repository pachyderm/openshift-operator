package generators

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/releaseutil"
	corev1 "k8s.io/api/core/v1"
)

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"resources": resourceList,
	}
}

func getValues(pd *aimlv1beta1.Pachyderm, directory *ChartDirectory) (chartutil.Values, error) {
	var values bytes.Buffer
	tmpl := template.Must(template.New(filepath.Base(directory.Values)).
		Funcs(templateFuncs()).ParseFiles(directory.Values))
	if err := tmpl.Execute(&values, pd); err != nil {
		return chartutil.Values{}, err
	}

	return chartutil.ReadValues(values.Bytes())
}

func loadPachydermTemplates(pd *aimlv1beta1.Pachyderm) (map[string]string, error) {
	directory, err := getChartDirectory(pd.Spec.Version)
	if err != nil {
		return nil, err
	}

	chart, err := loader.Load(directory.Chart)
	if err != nil {
		return nil, err
	}

	settings := cli.New()
	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		settings.RESTClientGetter(),
		pd.Namespace,
		os.Getenv("HELM_DRIVER"),
		func(formatter string, v ...interface{}) {
			fmt.Printf(formatter, v)
		})
	if err != nil {
		return nil, err
	}

	values, err := getValues(pd, directory)
	if err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = pd.Namespace
	client.ReleaseName = pd.Name
	client.DisableHooks = true
	client.ClientOnly = true
	client.DryRun = true
	release, err := client.Run(chart, values.AsMap())
	if err != nil {
		return nil, err
	}

	return releaseutil.SplitManifests(release.Manifest), nil
}

type ChartDirectory struct {
	// Path to the chart values.yaml file
	Values string
	// Name of .tgz file containing chart
	Chart string
}

func getChartDirectory(version string) (*ChartDirectory, error) {
	chartDir := filepath.Join("/", "charts", version[1:])
	if _, err := os.Stat(chartDir); err != nil {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		chartDir = filepath.Join(wd, "hack", "charts", version[1:])
		if _, err := os.Stat(chartDir); err != nil {
			return nil, err
		}
	}

	entry, err := os.ReadDir(chartDir)
	if err != nil {
		return nil, err
	}

	var chartName, valuesFile string
	for _, f := range entry {
		if strings.Contains(f.Name(), "values.y") {
			valuesFile = f.Name()
		}
		if strings.Contains(f.Name(), ".tgz") {
			chartName = f.Name()
		}
	}

	return &ChartDirectory{
		Chart:  filepath.Join(chartDir, chartName),
		Values: filepath.Join(chartDir, valuesFile),
	}, nil
}

// Resources provides a way to access resource
// limits from the template
type Resources struct {
	Requests map[string]string
	Limits   map[string]string
}

func resourceList(resources corev1.ResourceList) map[string]string {
	if reflect.DeepEqual(resources, corev1.ResourceList{}) ||
		resources == nil {
		return nil
	}

	response := map[string]string{}
	for name, limit := range resources {
		response[string(name)] = limit.String()
	}

	return response
}
