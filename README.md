# Chart Testing

`ct` is the the tool for testing Helm charts.
It is meant to be used for linting and testing pull requests.
It automatically detects charts changed against the target branch.

## Installation

### Prerequisites

It is recommended to use the provided Docker image which can be [found on Quay](quay.io/helmpack/chart-testing/).
It comes with all necessary tools installed.

* Helm (http://helm.sh)
* yamllint (https://github.com/adrienverge/yamllint)
* yamale (https://github.com/23andMe/Yamale)
* kubectl (https://kubernetes.io/docs/reference/kubectl/overview/)
* Tooling for your cluster

### Binary Distribution

Download the release distribution for your OS from the Releases page:

https://github.com/helm/chart-testing/releases

Unpack the `ct` binary, add it to your PATH, and you are good to go!

### Docker Image

A Docker image is available at `quay.io/helmpack/chart-testing`.

## Usage

See documentation for individual commands:

* [ct doc-gen](ct_doc-gen.md)
* [ct install](ct_install.md)
* [ct lint](ct_lint.md)
* [ct lint-and-install](ct_lint-and-install.md)
* [ct version](ct_version.md)


## Configuration

`ct` is a command-line application.
All command-line flags can also be set via environment variables or config file.
Environment variables must be prefixed with `CT_`. Underscores must be used instead of hyphens.

CLI flags, environment variables, and a config file can be mixed. The following order of precedence applies:

1. CLI flags
1. Environment variables
1. Config file

### Examples

The following example show various way of configuring the same thing:

#### CLI

    ct install --remote upstream --chart-dirs stable,incubator --build-id pr-42

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

`ct install --config config.yaml`

`ct` supports any format [Viper](https://github.com/spf13/viper) can read, i. e. JSON, TOML, YAML, HCL, and Java properties files.

## Building from Source

`ct` is built using Go 1.11. Older versions may work but have not been tested.

`build.sh` is used to build and release the tool. It uses [Goreleaser](https://goreleaser.com/) under the covers.

Note: on MacOS you will need to `GNU Coreutils readlink`. You can install with:

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

CircleCI creates releases automatically when a new tag is pushed. Tags are created using `tag.sh`.

```console
$ ./tag.sh -h
Usage: tag.sh <options>

Create and push a tag.

    -h, --help        Display help
    -d, --debug       Display verbose output
    -r, --remote      The name of the remote to push the tag to (default: upstream)
    -f, --force       Force an existing tag to be overwritten
    -t, --tag         The name of the tag to create
    -s, --skip-push   Skip pushing the tag
```
