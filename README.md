# Chart Testing

Bash library for linting and testing Helm charts.
Comes prepackaged as Docker image for easy use.

[chartlib.sh](lib/chartlib.sh) is a Bash library with useful function for linting and testing charts.
It is well documented and should be easily usable.
The script is meant to be sourced and can be configured via environment variables.

As a convenience, [chart_test.sh](chart_test.sh) is provided.
It supports linting and testing charts that have changed against a target branch.

## Prerequisites

It is recommended to use the provided Docker image which can be [found on Quay](quay.io/helmpack/chart-testing/).
It comes with all necessary tools installed.

* Bash 4.4 (https://tiswww.case.edu/php/chet/bash/bashtop.html)
* Helm (http://helm.sh)
* yq (https://github.com/kislyuk/yq)
* vert (https://github.com/Masterminds/vert)
* yamllint (https://github.com/adrienverge/yamllint)
* yamale (https://github.com/23andMe/Yamale)
* kubectl (https://kubernetes.io/docs/reference/kubectl/overview/)
* Tooling for your cluster

Note that older Bash versions may not work!

## Installation

Clone the repository and add it to the `PATH`.
The script must be run in the root directory of a Git repository.

```shell
$ chart_test.sh --help
Usage: chart_test.sh <options>
    Lint, install, and test Helm charts.
    -h, --help        Display help
    --verbose         Display verbose output
    --no-lint         Skip chart linting
    --no-install      Skip chart installation
    --all             Lint/install all charts
    --charts          Lint/install:
                        a standalone chart (e. g. stable/nginx)
                        a list of charts (e. g. stable/nginx,stable/cert-manager)
    --config          Path to the config file (optional)
    --                End of all options
```

## Configuration

The following environment variables can be set to configure [chartlib.sh](lib/chartlib.sh).
Note that this must be done before the script is sourced.

| Variable | Description | Default |
| - | - | - |
| `REMOTE` | The name of the Git remote to check against for changed charts | `origin` |
| `TARGET_BRANCH` | The name of the Git target branch to check against for changed charts | `master` |
| `CHART_DIRS` | Array of directories relative to the repo root containing charts | `(charts)` |
| `EXCLUDED_CHARTS` | Array of directories of charts that should be skipped | `()` |
| `CHART_REPOS` | Array of additional chart repos to add (`<name>=<url>`) | `()` |
| `TIMEOUT` | Timeout for chart installation in seconds | `300` |
| `LINT_CONF` | Config file for YAML linter | `/testing/etc/lintconf.yaml` (path of default config file in Docker image) |
| `CHART_YAML_SCHEMA` | YAML schema for `Chart.yaml` | `/testing/etc/chart_schema.yaml` (path of default schema file in Docker image) |
| `VALIDATE_MAINTAINERS`| If `true`, maintainer names in `Chart.yaml` are validated to be existing Github accounts | `true` |
| `GITHUB_INSTANCE`| Url of Github instance for maintainer validation | `https://github.com` |
| `CHECK_VERSION_INCREMENT`| If `true`, the chart version is checked to be incremented from the version on the remote target branch | `true` |

Note that `CHART_DIRS`, `EXCLUDED_CHARTS`, and `CHART_REPOS` must be configured as Bash arrays.

## Usage

The library is meant to be used for linting and testing pull requests.
It automatically detects charts changed against the target branch.
The environment variables mentioned in the configuration section above can be set in a config file for `chart_test.sh`.

By default, changes are detected against `origin/master`.
Depending on your CI setup, it may be necessary to configure and fetch a separate remote for this.

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
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-install --config .mytestenv
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

#### Linting Unchanged Charts

You can lint all charts with `--all` flag (chart version bump check will be ignored):

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-install --config .mytestenv --all
```

You can lint a list of charts (separated by comma) with `--charts` flag (chart version bump check will be ignored):

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-install --config .mytestenv --charts stable/nginx,stable/cert-manager
```

You can lint a single chart with `--charts` flag (chart version bump check will be ignored):

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-install --config .mytestenv --charts stable/nginx
```

### Installing and Testing Charts

Installing a chart requires access to a Kubernetes cluster.
You may have to create your own Docker image that extends from `quay.io/helmpack/chart-testing:v1.0.5` in order to install additional tools (e. g. `google-cloud-sdk` for GKE).
A container from such an image could run steps to authenticate to a Kubernetes cluster, where it initializes the `kubectl` context, before running `chart_test.sh`.

Charts are installed into newly created namespaces that will be deleted again afterwards.
By default, they are named by the chart, which may not be a good idea, especially when multiple PR jobs could be running for the same chart.
`chart_lib.sh` looks for an environment variable `BUILD_ID` and uses it to name the namespace.
Make sure you set it based on the pull request number.

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-lint --config .mytestenv
```

#### Installing Unchanged Charts

You can force to install all charts with `--all` flag:

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-lint --config .mytestenv --all
```

You can force to install a list of charts (separated by comma) with `--charts` flag:

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-lint --config .mytestenv --charts stable/nginx,stable/cert-manager
```

You can force to install one chart with `--charts` flag:

```shell
docker run --rm -v "$(pwd):/workdir" --workdir /workdir quay.io/helmpack/chart-testing:v1.0.5 chart_test.sh --no-lint --config .mytestenv --charts stable/nginx
```

#### GKE Example

An example for GKE is available in the [examples/gke](examples/gke) directory.
A custom `Dockerfile` additionally installs the `google-cloud-sdk` and a custom shell script puts everything together.

#### Docker for Mac Example

An example for Docker for Mac is available in the [examples/docker-for-mac](examples/docker-for-mac) directory.
This script can be run as is in the [charts](https://github.com/helm/charts) repo.
Make sure `Show system containers` is active for Docker's Kubernetes distribution, so the script can find the API server and configure `kubectl` so it can access the API server from within the container.
