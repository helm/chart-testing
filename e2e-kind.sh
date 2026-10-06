#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLUSTER_NAME=chart-testing
readonly CLUSTER_NAME

K8S_VERSION=v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5
readonly K8S_VERSION

create_kind_cluster() {
    kind create cluster --name "${CLUSTER_NAME}" --image "kindest/node:${K8S_VERSION}" --wait 60s

    kubectl cluster-info || kubectl cluster-info dump
    echo

    kubectl get nodes
    echo

    echo 'Cluster ready!'
    echo
}

test_e2e() {
    go test -cover -race -tags=integration ./...
    echo
}

cleanup() {
    kind delete cluster --name "${CLUSTER_NAME}"
    echo 'Done!'
}

main() {
    trap cleanup EXIT

    create_kind_cluster
    test_e2e
}

main
