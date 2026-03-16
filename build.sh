#!/bin/bash

set -e -u

VERSION=${VERSION:-0.7}
RELEASE=${RELEASE:-0}

OUTPUT_DIR=output
INSTALL_DIR=${OUTPUT_DIR}/install/
INSTALL_PREFIX=opt/zapier/pag
PROG_DIR=${INSTALL_DIR}/${INSTALL_PREFIX}
RELEASE_DIR=${PWD}/${OUTPUT_DIR}/release

main() {
    local prog=${PROG_DIR}/prom-aggregation-gateway
    local config=$PWD/spec/pag.yml

    rm -r -f ${OUTPUT_DIR}
    mkdir -p ${PROG_DIR}
    go build -o $prog

    (
        mkdir ${RELEASE_DIR}
        cd ${INSTALL_DIR}
        VERSION=$VERSION RELEASE=$RELEASE nfpm pkg --packager rpm --config $config --target ${RELEASE_DIR}
    )
}

main
