#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly CLUSTER_NAME=chart-testing
readonly K8S_VERSION=v1.13.4

create_kind_cluster() {
    kind create cluster --name "$CLUSTER_NAME" --image "kindest/node:$K8S_VERSION" --wait 60s
    KUBECONFIG="$(kind get kubeconfig-path --name="$CLUSTER_NAME")"
    export KUBECONFIG

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

install_local-path-provisioner() {
    # kind doesn't support Dynamic PVC provisioning yet, this is one way to get it working
    # https://github.com/rancher/local-path-provisioner

    # Remove default storage class. It will be recreated by local-path-provisioner
    kubectl delete storageclass standard

    echo 'Installing local-path-provisioner...'
    kubectl apply -f examples/kind/test/local-path-provisioner.yaml
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
    install_local-path-provisioner
    install_tiller
    test_e2e
}

main
