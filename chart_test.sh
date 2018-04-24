#!/usr/bin/env bash

# Copyright 2018 The Kubernetes Authors. All rights reserved.
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

readonly REPO_ROOT=$(git rev-parse --show-toplevel)
readonly SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

show_help() {
cat << EOF
Usage: $(basename "$0") <options>
    Lint, install, and test Helm charts.
    -h, --help        Display help
    --verbose         Display verbose output
    --no-lint         Skip chart linting
    --no-install      Skip chart installation
    --config          Path to the config file (optional)
    --                End of all options
EOF
}

main() {
    local no_lint=
    local no_install=
    local config=
    local verbose=

    while :; do
        case "${1:-}" in
            -h|--help)
                show_help
                exit
                ;;
            --verbose)
                verbose=true
                ;;
            --no-install)
                no_install=true
                ;;
            --no-lint)
                no_lint=true
                ;;
            --config)
                if [ -n "$2" ]; then
                    config="$2"
                    shift
                else
                    echo "ERROR: '--config' cannot be empty." >&2
                    exit 1
                fi
                ;;
            --) # End of all options.
                shift
                break
                ;;
            -?*)
                echo "WARN: Unknown option (ignored): $1" >&2
                ;;
            *)
                break
                ;;
        esac

        shift
    done

    if [[ -n "$config" ]]; then
        if [[ -f "$config" ]]; then
            # shellcheck disable=SC1090
            source "$config"
        else
            echo "ERROR: Specified config file does not exist: $config" >&2
            exit 1
        fi
    fi

    # shellcheck source=lib/chartlib.sh
    source "$SCRIPT_DIR/lib/chartlib.sh"

    [[ -n "$verbose" ]] && set -o xtrace

    pushd "$REPO_ROOT" > /dev/null

    read -ra changed_dirs <<< "$(chartlib::determine_changed_directories)"

    if [[ -n "${changed_dirs[*]}" ]]; then
        echo "Charts to be installed and tested: ${changed_dirs[*]}"

        chartlib::init_helm

        local error=
        for chart_dir in "${changed_dirs[@]}"; do
            echo ''
            echo '-----------------------------------------------------------------------'
            echo "Processing chart '$chart_dir'..."
            echo '-----------------------------------------------------------------------'
            echo ''

            local local_error=

            if [[ -z "$no_lint" ]]; then
                if ! chartlib::validate_chart "$chart_dir"; then
                    local_error=true
                fi
                if ! chartlib::lint_chart_with_all_configs "$chart_dir"; then
                    local_error=true
                fi
            fi

            if [[ -z "$no_install" && -z "$local_error" ]]; then
                if ! chartlib::install_chart_with_all_configs "$chart_dir"; then
                    local_error=true
                fi
            fi

            if [[ -n "$local_error" ]]; then
                error=true
            fi
        done

        if [[ -n "$error" ]]; then
            chartlib::error "Script terminated with error(s)."
            exit 1
        fi
    fi

    popd > /dev/null
}

main "$@"
