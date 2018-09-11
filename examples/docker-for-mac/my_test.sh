#!/usr/bin/env bash

# Copyright 2018 The Helm Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

readonly IMAGE_TAG=v1.0.4
readonly IMAGE_REPOSITORY="gcr.io/kubernetes-charts-ci/chart-testing"

main() {
    local testcontainer_id
    testcontainer_id=$(create_testcontainer)

    # shellcheck disable=SC2064
    trap "docker container rm --force $testcontainer_id > /dev/null" EXIT

    configure_kubectl "$testcontainer_id"
    run_test
}

lookup_apiserver_container_id() {
    docker container list --filter name=k8s_kube-apiserver --format '{{ .ID }}'
}

get_apiserver_arg() {
    local container_id="$1"
    local arg="$2"
    docker container inspect "$container_id" | jq -r ".[].Args[] | capture(\"$arg=(?<arg>.*)\") | .arg"
}

create_testcontainer() {
    docker container run --user 1000:1000 --interactive --tty --detach \
        --volume "$(pwd):/workdir" --workdir /workdir \
        "$IMAGE_REPOSITORY:$IMAGE_TAG" cat
}

configure_kubectl() {
    local testcontainer_id="$1"

    local apiserver_id
    apiserver_id=$(lookup_apiserver_container_id)

    if [[ -z "$apiserver_id" ]]; then
        echo "ERROR: API-Server container not found. Make sure 'Show system containers' is enabled in Docker4Mac 'Preferences'!" >&2
        exit 1
    fi

    local ip
    ip=$(get_apiserver_arg "$apiserver_id" --advertise-address)

    local port
    port=$(get_apiserver_arg "$apiserver_id" --secure-port)

    docker cp "$HOME/.kube" "$testcontainer_id:/tmp/.kube"
    docker exec --user root "$testcontainer_id" chown -R 1000:1000 /tmp
    docker exec "$testcontainer_id" kubectl config set-cluster docker-for-desktop-cluster "--server=https://$ip:$port"
    docker exec "$testcontainer_id" kubectl config set-cluster docker-for-desktop-cluster --insecure-skip-tls-verify=true
    docker exec "$testcontainer_id" kubectl config use-context docker-for-desktop
}

run_test() {
    git remote add k8s https://github.com/helm/charts.git &> /dev/null || true
    git fetch k8s
    docker exec "$testcontainer_id" chart_test.sh --config test/.testenv

    echo "Done Testing!"
}

main
