# Chart Testing

Bash library for linting and testing Helm charts. Comes prepackaged as Docker image for easy use.

[chartlib.sh](lib/chartlib.sh) is a Bash library with useful function for linting and testing charts. It is well documented and should be easily usable. The script is meant to be sourced and can be configured via environment variables.

As a convenience, [chart_test.sh](chart_test.sh) is provided. It supports linting and testing charts that have changed against a target branch.

## Prerequisites

It is recommended to use the provided Docker image. It comes with all necessary tools installed.

* Bash 4.4 (https://tiswww.case.edu/php/chet/bash/bashtop.html)
* Helm (http://helm.sh)
* yq (https://github.com/kislyuk/yq)
* vert (https://github.com/Masterminds/vert)
* yamllint (https://github.com/adrienverge/yamllint)
* yamale (https://github.com/23andMe/Yamale)
* kubectl (https://kubernetes.io/docs/reference/kubectl/overview/)
* Tooling for your cluster

Note that older Bash versions may no work!

## Installation

Clone the repository and add it to the `PATH`. The script must be run in the root directory of a Git repository.

```shell
$ chart_test.sh --help
Usage: chart_test.sh <options>
    Lint, install, and test Helm charts.
    -h, --help        Display help
    --verbose         Display verbose output
    --no-lint         Skip chart linting
    --no-install      Skip chart installation
    --config          Path to the config file (optional)
```

## Configuration

The following environment variables can be set to configure [chartlib.sh](lib/chartlib.sh). Note that this must be done before the script is sourced.

| Variable | Description | Default |
| - | - | - |
| `REMOTE` | The name of the Git remote to check against for changed charts | `origin` |
| `TARGET_BRANCH` | The name of the Git target branch to check against for changed charts | `master` |
| `CHART_DIRS` | Directories relative to the repo root containing charts | `charts` |
| `EXCLUDED_CHARTS` | Directories of charts that should be skipped | |
| `CHART_REPOS` | Additional chart repos to add (`<name>=<url>`) | |
| `TIMEOUT` | Timeout for chart installation in seconds | `300` |
| `LINT_CONF` | Config file for YAML linter | `/testing/etc/lintconf.yaml` (path of default config file in Docker image) |
| `CHART_YAML_SCHEMA` | YAML schema for `Chart.yaml` | `/testing/etc/chart_schema.yaml` (path of default schema file in Docker image) |
| `VALIDATE_MAINTAINERS`| If `true`, maintainer names in `Chart.yaml` are validated to be existing Github accounts | `true` |

`CHART_DIRS`, `EXCLUDED_CHARTS`, and `CHART_REPOS` may be set as strings with values separated by spaces or as Bash arrays.

## Usage

The library is meant to be used for linting and testing pull requests. It automatically detects charts changed against the target branch. The environment variables mentioned in the configuration section above can be set in a config file for `chart_test.sh`.

By default, changes are detected against `origin/master`. Depending on your CI setup, it may be necessary to configure and fetch a separate remote for this.

```shell
REMOTE=myremote
```
```shell
git remote add myremote <repo_url></repo_url>
git fetch myremote
chart-test.sh
```

### Linting Charts

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir gcr.io/kubernetes-charts-ci/chart-testing:v1.0.0 chart_test.sh --no-install --config .mytestenv
```

*Sample Output*

```
-----------------------------------------------------------------------
Environment:
REMOTE=k8s
TARGET_BRANCH=master
CHART_DIRS=stable
EXCLUDED_CHARTS=
CHART_REPOS=
TIMEOUT=600
LINT_CONF=/testing/etc/lintconf.yaml
CHART_YAML_SCHEMA=/testing/etc/chart_schema.yaml
VALIDATE_MAINTAINERS=true
-----------------------------------------------------------------------
Charts to be installed and tested: stable/dummy
Initializing Helm client...
Creating /home/testing/.helm
Creating /home/testing/.helm/repository
Creating /home/testing/.helm/repository/cache
Creating /home/testing/.helm/repository/local
Creating /home/testing/.helm/plugins
Creating /home/testing/.helm/starters
Creating /home/testing/.helm/cache/archive
Creating /home/testing/.helm/repository/repositories.yaml
Adding stable repo with URL: https://kubernetes-charts.storage.googleapis.com
Adding local repo with URL: http://127.0.0.1:8879/charts
$HELM_HOME has been configured at /home/testing/.helm.
Not installing Tiller due to 'client-only' flag having been set
Happy Helming!

-----------------------------------------------------------------------
Processing chart 'stable/dummy'...
-----------------------------------------------------------------------

Validating chart 'stable/dummy'...
Checking chart 'stable/dummy' for a version bump...
Unable to find chart on master. New chart detected.
Linting 'stable/dummy/Chart.yaml'...
Linting 'stable/dummy/values.yaml'...
Validating Chart.yaml
Validating /workdir/stable/dummy/Chart.yaml...
Validation success! 👍
Validating maintainers
Verifying maintainer 'unguiculus'...
Using custom values file 'stable/dummy/ci/ci-values.yaml'...
Linting chart 'stable/dummy'...
==> Linting stable/dummy
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, no failures
Done.
```

### Installing and Testing Charts

Installing a chart requires access to a Kubernetes cluster. You may have to create your own Docker image that extends from `gcr.io/kubernetes-charts-ci/chart-testing:v1.0.0` in order to install additional tools (e. g. `google-cloud-sdk` for GKE). You could run a container in the background, run the required steps to authenticatre and initialize the `kubectl` context before you, and eventually run `chart_test.sh`.

Charts are installed into newly created namespaces that will be deleted again afterwards. By default, they are named by the chart, which may not be a good idea, especially when multiple PR jobs could be running for the same chart. `chart_lib.sh` looks for an environment variable `BUILD_ID` and uses it to name the namespace. Make sure you set it based on the pull request number.

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir gcr.io/kubernetes-charts-ci/chart-testing:v1.0.0 chart_test.sh --no-lint --config .mytestenv
```

#### GKE Example

An example for GKE is available in the [examples/gke](examples/gke) directory. A custom `Dockerfile` additionally installs the `google-cloud-sdk` and a custom shell script puts everything together.
