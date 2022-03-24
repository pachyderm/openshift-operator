#!/bin/bash

setup_catalog_directory() {
    cleanup

    echo "Setup custom index image..."
    mkdir pachyderm-operator-index
    opm alpha generate dockerfile pachyderm-operator-index
    opm init pachyderm-operator --default-channel=stable --description=./README.md --output yaml > pachyderm-operator-index/operator.yaml
    opm render quay.io/opdev/pachyderm-bundle:${VERSION} --output yaml >> pachyderm-operator-index/operator.yaml
}

prepare_index_image() {
cat << EOF >> pachyderm-operator-index/operator.yaml
---
schema: olm.channel
package: pachyderm-operator
name: stable
entries:
- name: pachyderm-operator.v${VERSION}
EOF
}

build_index_image() {
    opm validate ./pachyderm-operator-index
    if [ $? -eq 0 ]
    then
        ${BUILD_TOOL} build -t quay.io/opdev/pachyderm-index:latest -f pachyderm-operator-index.Dockerfile .
    fi
}

cleanup() {
    if [ -d ./pachyderm-operator-index ]
    then
        rm -rf ./pachyderm-operator-index ./pachyderm-operator-index.Dockerfile
    fi
}

setup_catalog_directory
prepare_index_image
build_index_image
cleanup
