#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CLUSTER_NAME=chart-testing
readonly CLUSTER_NAME

K8S_VERSION=v1.17.0
readonly K8S_VERSION

create_kind_cluster() {
    kind create cluster --name "$CLUSTER_NAME" --image "kindest/node:$K8S_VERSION" --wait 60s

    kubectl cluster-info || kubectl cluster-info dump
    echo

    kubectl get nodes
    echo

    echo 'Cluster ready!'
    echo
}

install_tiller() {
    echo 'Installing Tiller...'
    kubectl --namespace kube-system --output yaml create serviceaccount tiller --dry-run | kubectl apply -f -
    kubectl create --output yaml clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller --dry-run | kubectl apply -f -
    helm init --service-account tiller --upgrade --wait
    echo
}

test_e2e() {
    go test -cover -race -tags=integration ./...
    echo
}

cleanup() {
    kind delete cluster --name "$CLUSTER_NAME"
    echo 'Done!'
}

main() {
    trap cleanup EXIT

    create_kind_cluster
    install_tiller
    test_e2e
}

main
