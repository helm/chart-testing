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

readonly IMAGE_REPOSITORY="myrepo/chart-testing"
readonly IMAGE_REPOSITORY="v1.0.0"
readonly REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"

main() {
    local config_container_id
    config_container_id=$(docker run -ti -d \
        -v "$GOOGLE_APPLICATION_CREDENTIALS:/service-account.json" \
        -v "$REPO_ROOT:/workdir" \
        -e "BUILD_ID=$PULL_NUMBER" \
        "$IMAGE_REPOSITORY:$IMAGE_TAG" cat)

    # shellcheck disable=SC2064
    trap "docker rm -f $config_container_id" EXIT

    docker exec "$config_container_id" gcloud auth activate-service-account --key-file /service-account.json
    docker exec "$config_container_id" gcloud container clusters get-credentials my-cluster --project my-project --zone us-west1-a
    docker exec "$config_container_id" kubectl cluster-info
    docker exec "$config_container_id" chart_test.sh --config /workdir/.testenv
}

main
