### Pachyderm helm chart

This directory (/hack/charts) contains the Pachyderm helm charts corresponding to the Pachyderm releases supported by the OpenShift operator

When updating this operator to support a new version of Pachyderm, we'll need to update the helm charts as well. This can be done by running:
```
   mkdir <latest version>; cd <latest version>
   helm repo add pach https://helm.pachyderm.com
   helm repo update
   helm pull pach/pachyderm
   cp ../<prev version>/images.json ./
   # this file contains the docker images deployed by the operator. These
   # override the images in Pachyderm's helm chart, replacing them with
   # Red Hat Marketplace-compatible images (those that have been reviewed and
   # approved by Red Hat)
   ${EDITOR} images.json
   cp ../<prev version>/values.yaml ./
   # this file contains the values used to instantiate Pachyderm's helm chart.
   # Changes to Pachyderm's helm chart will likely require changes to this file.
   ${EDITOR} values.yaml
```

