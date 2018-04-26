# Chart Testing

Bash library for linting and testing Helm charts. Comes prepackaged as Docker image for easy use.

[chartlib.sh](lib/chartlib.sh) is a Bash library with useful function for linting and testing charts. It is well documented and should be easily usable. The script is meant to be sourced and can be configured via environment variables.

As a convenience, [chartlib.sh](chart_test.sh) is provided. It supports linting and testing charts that have changed against a target branch.

## Configuration

The following environment variables can be set to configure [chartlib.sh](lib/chartlib.sh). Note that this must be done before the script is sourced.

| Variable | Description | Default |
| - | - | - |
| REMOTE | The name of the Git remote to check against for changed charts | origin |
| TARGET_BRANCH | The name of the Git target branch to check against for changed charts | master |
| CHART_DIRS | Directories relative to the repo root containing charts | charts |
| EXCLUDED_CHARTS | Directories of charts that should be skipped | |
| CHART_REPOS | Additional chart repos to add (<name>=<url>) | |
| TIMEOUT | Timeout for chart installation in seconds | 300 |
| LINT_CONF | Config file for YAML linter | /testing/etc/lintconf.yaml |
| CHART_YAML_SCHEMA | YAML schema for Chart.yaml | /testing/etc/chart_schema.yaml |

CHART_DIRS, EXCLUDED_CHARTS, and CHART_REPOS may be set as a string with values separated by a space of as a Bash array.

## Linting

to be continued...
