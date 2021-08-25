package generators

import (
	"encoding/json"
	"io/ioutil"

	aimlv1beta1 "github.com/pachyderm/openshift-operator/api/v1beta1"
)

// ImageCatalog consists the certified images
// required to deploy a Pachyderm cluster
type ImageCatalog struct {
	Pachd    *aimlv1beta1.ImageOverride `json:"pachd,omitempty"`
	Console  *aimlv1beta1.ImageOverride `json:"console,omitempty"`
	Postgres *aimlv1beta1.ImageOverride `json:"postgres,omitempty"`
	Etcd     *aimlv1beta1.ImageOverride `json:"etcd,omitempty"`
	Worker   *aimlv1beta1.ImageOverride `json:"worker,omitempty"`
}

func getDefaultCertifiedImages(images string) (*ImageCatalog, error) {
	data, err := ioutil.ReadFile(images)
	if err != nil {
		return nil, err
	}

	catalog := &ImageCatalog{}
	if err := json.Unmarshal(data, catalog); err != nil {
		return nil, err
	}

	return catalog, nil
}

func (i *ImageCatalog) inject(pd *aimlv1beta1.Pachyderm) {
	pd.Spec.Console.Image = setImageOptions(pd.Spec.Console.Image, i.Console)
	pd.Spec.Pachd.Image = setImageOptions(pd.Spec.Pachd.Image, i.Pachd)
	pd.Spec.Worker.Image = setImageOptions(pd.Spec.Worker.Image, i.Worker)
	pd.Spec.Etcd.Image = setImageOptions(pd.Spec.Etcd.Image, i.Etcd)
}

func setImageOptions(user, shipped *aimlv1beta1.ImageOverride) *aimlv1beta1.ImageOverride {
	if user == nil {
		return shipped
	}

	if user.Tag != "" {
		shipped.Tag = user.Tag
	}

	if user.PullPolicy != "" {
		shipped.PullPolicy = user.PullPolicy
	}

	if user.Repository != "" {
		shipped.Repository = user.Repository
	}

	return shipped
}

func postgresqlImage(pd *aimlv1beta1.Pachyderm) *aimlv1beta1.ImageOverride {
	directory, err := getChartDirectory(pd.Spec.Version)
	if err != nil {
		return nil
	}

	catalog, err := getDefaultCertifiedImages(directory.Images)
	if err != nil {
		return nil
	}

	return catalog.Postgres
}
