#!/usr/bin/env bash

# Copyright The Helm Authors. All rights reserved.
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

readonly SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

show_help() {
cat << EOF
Usage: $(basename "$0") <options>

Build ct using Goreleaser.

    -h, --help      Display help
    -d, --debug     Display verbose output and run Goreleaser with --debug
    -r, --release   Create a release using Goreleaser. This includes the creation
                    of a GitHub release and building and pushing the Docker image.
                    If this flag is not specified, Goreleaser is run with --snapshot
EOF
}

main() {
    local debug=
    local release=

    while :; do
        case "${1:-}" in
            -h|--help)
                show_help
                exit
                ;;
            -d|--debug)
                debug=true
                ;;
            -r|--release)
                release=true
                ;;
            *)
                break
                ;;
        esac

        shift
    done

    local goreleaser_args=(--rm-dist)

    if [[ -n "$debug" ]]; then
        goreleaser_args+=( --debug)
        set -x
    fi

    if [[ -z "$release" ]]; then
        goreleaser_args+=( --snapshot)
    fi

    pushd "$SCRIPT_DIR" > /dev/null

    dep ensure -v
    go test ./...
    goreleaser "${goreleaser_args[@]}"

    popd > /dev/null
}

main "$@"
