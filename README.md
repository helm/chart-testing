# Chart Testing

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/helm/chart-testing)](https://goreportcard.com/report/github.com/helm/chart-testing)
![ci](https://github.com/helm/chart-testing/workflows/ci/badge.svg)

`ct` is the tool for testing Helm charts.
It is meant to be used for linting and testing pull requests.
It automatically detects charts changed against the target branch.

## Installation

### Prerequisites

It is recommended to use the provided Docker image which can be [found on Quay](https://quay.io/repository/helmpack/chart-testing).
It comes with all necessary tools installed.

* [Helm](http://helm.sh)
* [Git](https://git-scm.com) (2.17.0 or later)
* [Yamllint](https://github.com/adrienverge/yamllint)
* [Yamale](https://github.com/23andMe/Yamale)
* [Kubectl](https://kubernetes.io/docs/reference/kubectl/overview/)

### Binary Distribution

Download the release distribution for your OS from the Releases page:

https://github.com/helm/chart-testing/releases

Unpack the `ct` binary, add it to your PATH, and you are good to go!

### Docker Image

A Docker image is available at `quay.io/helmpack/chart-testing` with list of
available tags [here](https://quay.io/repository/helmpack/chart-testing?tab=tags).

### Homebrew

```console
$ brew install chart-testing
```

## Usage

See documentation for individual commands:

* [ct](doc/ct.md)
* [ct install](doc/ct_install.md)
* [ct lint](doc/ct_lint.md)
* [ct lint-and-install](doc/ct_lint-and-install.md)
* [ct list-changed](doc/ct_list-changed.md)
* [ct version](doc/ct_version.md)

For a more extensive how-to guide, please see:

* [charts-repo-actions-demo](https://github.com/helm/charts-repo-actions-demo)

## Configuration

`ct` is a command-line application.
All command-line flags can also be set via environment variables or config file.
Environment variables must be prefixed with `CT_`.
Underscores must be used instead of hyphens.

CLI flags, environment variables, and a config file can be mixed.
The following order of precedence applies:

1. CLI flags
1. Environment variables
1. Config file

Note that linting requires config file for [yamllint](https://github.com/adrienverge/yamllint) and [yamale](https://github.com/23andMe/Yamale).
If not specified, these files are search in the current directory, the `.ct` directory in current directory, `$HOME/.ct`, and `/etc/ct`, in that order.
Samples are provided in the [etc](etc) folder.

### Examples

The following example show various way of configuring the same thing:

#### CLI

#### Remote repo

With remote repo:

    ct install --remote upstream --chart-dirs stable,incubator --build-id pr-42

#### Local repo

If you have a chart in current directory and ct installed on the host then you can run:

    ct install --chart-dirs . --charts .

With docker it works with:

    docker run -it --network host --workdir=/data --volume ~/.kube/config:/root/.kube/config:ro --volume $(pwd):/data quay.io/helmpack/chart-testing:v3.7.1 ct install --chart-dirs . --charts .

Notice that `workdir` param is important and must be the same as volume mounted.


#### Environment Variables

    export CT_REMOTE=upstream
    export CT_CHART_DIRS=stable,incubator
    export CT_BUILD_ID

    ct install

#### Config File

`config.yaml`:

```yaml
remote: upstream
chart-dirs:
  - stable
  - incubator
build-id: pr-42
```

#### Config Usage

    ct install --config config.yaml


`ct` supports any format [Viper](https://github.com/spf13/viper) can read, i. e. JSON, TOML, YAML, HCL, and Java properties files.

Notice that if no config file is specified, then `ct.yaml` (or any of the supported formats) is loaded from the current directory, `$HOME/.ct`, or `/etc/ct`, in that order, if found.


#### Using private chart repositories

When adding chart-repos you can specify additional arguments for the `helm repo add` command using `helm-repo-extra-args` on a per-repo basis.
You can also specify OCI registries which will be added using the `helm registry login` command, they also support the `helm-repo-extra-args` for authentication.
This could for example be used to authenticate a private chart repository.

`config.yaml`:

```yaml
chart-repos:
  - incubator=https://incubator.io
  - basic-auth=https://private.com
  - ssl-repo=https://self-signed.ca
  - oci-registry=oci://nice-oci-registry.pt
helm-repo-extra-args:
  - ssl-repo=--ca-file ./my-ca.crt
```

    ct install --config config.yaml --helm-repo-extra-args "basic-auth=--username user --password secret"

## Building from Source

`ct` is built using Go 1.13 or higher.

`build.sh` is used to build and release the tool.
It uses [Goreleaser](https://goreleaser.com/) under the covers.

Note: on MacOS you will need `GNU Coreutils readlink`.
You can install it with:

```console
brew install coreutils
```

Then add `gnubin` to your `$PATH`, with:

```console
echo 'export PATH="$(brew --prefix coreutils)/libexec/gnubin:$PATH"' >> ~/.bash_profile
bash --login
```

To use the build script:

```console
$ ./build.sh -h
Usage: build.sh <options>

Build ct using Goreleaser.

    -h, --help      Display help
    -d, --debug     Display verbose output and run Goreleaser with --debug
    -r, --release   Create a release using Goreleaser. This includes the creation
                    of a GitHub release and building and pushing the Docker image.
                    If this flag is not specified, Goreleaser is run with --snapshot
```

## Releasing

### Prepare Release

Before a release is created, versions have to be updated in the examples.
A pull request needs to be created for this, which should be merged right before the release is cut.
Here's a previous one for reference: https://github.com/helm/chart-testing/pull/89

### Create Release

The release workflow is [dispatched from github actions](https://github.com/helm/chart-testing/actions)
Versions must start with a lower-case `v`, e. g. `v3.7.1`.

## Supported versions

The previous MAJOR version will be supported for three months after each new MAJOR release.

Within this support window, pull requests for the previous MAJOR version should be made against the previous release branch.
For example, if the current MAJOR version is `v2`, the pull request base branch should be `release-v1`.

## Upgrading

When upgrading from `< v2.0.0` you will also need to change the usage in your scripts.
This is because, while the [v2.0.0](https://github.com/helm/chart-testing/releases/tag/v2.0.0) release has parity with `v1`, it was refactored from a bash library to Go so there are minor syntax differences.
Compare [v1 usage](https://github.com/helm/chart-testing/tree/release-v1#usage) with this (`v2`) version's README [usage](#usage) section above.
