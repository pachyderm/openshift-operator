#!/bin/bash

OPERATOR_DIR=$(pwd)
OPERATOR_PKG_NAME=$(sed -n '/projectName/p' PROJECT | cut -d: -f2 | xargs)
OPERATOR_CSV="${OPERATOR_DIR}/$(find bundle/manifests/ -name '*.clusterserviceversion.yaml' -type f)"

annotations_exist() {
	echo -n "Adding supported openshift versions..."
	egrep -q "com.redhat.openshift.versions" ${OPERATOR_ANNOTATIONS_FILE}
	if [ $? -eq 0 ]
	then
		echo "done"
	else
		openshift_annotations
		if [ $? -eq 0 ]
		then
			echo "done"
		fi
	fi
}

default_channel_exists() {
	echo -n "Checking default channel..."
	egrep -q "operators.operatorframework.io.bundle.channel.default.v1" ${OPERATOR_ANNOTATIONS_FILE}
	if [ $? -eq 0 ]
	then
		echo "done"
	else
		echo "failed"
	fi
}

check_min_kube_version() {
	echo -n "Checking csv.spec.minKubeVersion..."
	egrep -q "minKubeVersion" ${OPERATOR_CSV}
	if [ $? -eq 0 ]
	then
		echo "done"
	else
		echo "failed"
	fi

}

openshift_annotations() {
cat <<EOF>> ${OPERATOR_ANNOTATIONS_FILE}

  # OpenShift annotations.
  com.redhat.openshift.versions: v4.6-v4.9
EOF
}


if [[ ${OPERATOR_DIR} =~ ${OPERATOR_PKG_NAME} ]]
then
	OPERATOR_ANNOTATIONS_FILE="${OPERATOR_DIR}/bundle/metadata/annotations.yaml"
fi

if [ -f ${OPERATOR_ANNOTATIONS_FILE} ]
then
	annotations_exist
fi

default_channel_exists
check_min_kube_version
