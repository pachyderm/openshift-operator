# Pachyderm Operator

The pachyderm operator is an application responsible for installing, watching and managing Pachyderm resources in your Openshift cluster.

Pachyderm is a data science platform that combines Data Lineage with End-to-End Pipelines on Kubernetes, engineered for the enterprise.

## Usage

**1. Using AWS S3 for Pachd storage**

-  Create a secret which contains the AWS S3 storage information

```
$ oc create secret generic pachyderm-aws --from-literal access-id=ABCDEFGHIJKLMNOPQR --from-literal access-secret=dkhfjdshfj/fjkdshfiuUjmfhdsjkhfjdhs/KLhdfuiseh --from-literal bucket=pachyderm-bucket --from-literal region=us-east-1`

secret/pachyderm-aws created
$
```

- Create a Pachyderm custom resource in the same namespace

```
$ cat <<EOF> pachyderm-cr.yaml
apiVersion: aiml.pachyderm.com/v1beta1
kind: Pachyderm
metadata:
  name: pachyderm-sample
  namespace: pachyderm-test
spec:
  console:
    disable: true
  pachd:
    metrics:
      disable: false
    storage:
      amazon:
        credentialSecretName: pachyderm-aws
      backend: AMAZON
EOF
$ oc create -f pachyderm-cr.yaml
pachyderm.aiml.pachyderm.com/pachyderm-sample created
$ 
```

- Ensure pachyderm is up and running

```
$ oc get pachyderm pachyderm-sample -o yaml | yq e '.status' -
phase: Running
$   
```


