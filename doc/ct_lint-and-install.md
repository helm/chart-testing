## ct lint-and-install

Lint, install, and test a chart

### Synopsis

Combines 'lint' and 'install' commands.

```
ct lint-and-install [flags]
```

### Options

```
      --additional-commands strings          Additional commands to run per chart (default: [])
                                             Commands will be executed in the same order as provided in the list and will
                                             be rendered with go template before being executed.
                                             Example: "helm unittest --helm3 -f tests/*.yaml {{ .Path }}"
      --all                                  Process all charts except those explicitly excluded.
                                             Disables changed charts detection and version increment checking
      --build-id string                      An optional, arbitrary identifier that is added to the name of the namespace a
                                             chart is installed into. In a CI environment, this could be the build number or
                                             the ID of a pull request. If not specified, the name of the chart is used
      --chart-dirs strings                   Directories containing Helm charts. May be specified multiple times
                                             or separate values with commas (default [charts])
      --chart-repos strings                  Additional chart repositories for dependency resolutions.
                                             Repositories should be formatted as 'name=url' (ex: local=http://127.0.0.1:8879/charts).
                                             May be specified multiple times or separate values with commas
      --chart-yaml-schema string             The schema for chart.yml validation. If not specified, 'chart_schema.yaml'
                                             is searched in the current directory, '$HOME/.ct', and '/etc/ct', in
                                             that order.
      --charts strings                       Specific charts to test. Disables changed charts detection and
                                             version increment checking. May be specified multiple times
                                             or separate values with commas
      --check-version-increment              Activates a check for chart version increments (default true)
      --config string                        Config file
      --debug                                Print CLI calls of external tools to stdout (caution: setting this may
                                             expose sensitive data when helm-repo-extra-args contains passwords)
      --exclude-deprecated                   Skip charts that are marked as deprecated
      --excluded-charts strings              Charts that should be skipped. May be specified multiple times
                                             or separate values with commas
      --github-groups                        Change the delimiters for github to create collapsible groups
                                             for command output
      --helm-dependency-extra-args strings   Additional arguments for 'helm dependency build' (e.g. ["--skip-refresh"]
      --helm-extra-args string               Additional arguments for Helm. Must be passed as a single quoted string
                                             (e.g. '--timeout 500s')
      --helm-extra-set-args string           Additional arguments for Helm. Must be passed as a single quoted string
                                             (e.g. "--set=name=value"
      --helm-lint-extra-args string          Additional arguments for Helm lint subcommand. Must be passed as a single quoted string
                                             (e.g. '--quiet')
      --helm-repo-extra-args strings         Additional arguments for the 'helm repo add' command to be
                                             specified on a per-repo basis with an equals sign as delimiter
                                             (e.g. 'myrepo=--username test --password secret'). May be specified
                                             multiple times or separate values with commas
  -h, --help                                 help for lint-and-install
      --lint-conf string                     The config file for YAML linting. If not specified, 'lintconf.yaml'
                                             is searched in the current directory, '$HOME/.ct', and '/etc/ct', in
                                             that order
      --namespace string                     Namespace to install the release(s) into. If not specified, each release will be
                                             installed in its own randomly generated namespace
      --print-config                         Prints the configuration to stderr (caution: setting this may
                                             expose sensitive data when helm-repo-extra-args contains passwords)
      --release-label string                 The label to be used as a selector when inspecting resources created by charts.
                                             This is only used if namespace is specified (default "app.kubernetes.io/instance")
      --remote string                        The name of the Git remote used to identify changed charts (default "origin")
      --since string                         The Git reference used to identify changed charts (default "HEAD")
      --skip-clean-up                        Skip resources clean-up. Used if need to continue other flows or keep it around.
      --skip-missing-values                  When --upgrade has been passed, this flag will skip testing CI values files from the
                                             previous chart revision if they have been deleted or renamed at the current chart
                                             revision
      --target-branch string                 The name of the target branch used to identify changed charts (default "main")
      --upgrade                              Whether to test an in-place upgrade of each chart from its previous revision if the
                                             current version should not introduce a breaking change according to the SemVer spec
      --use-helmignore                       Use .helmignore when identifying changed charts
      --validate-chart-schema                Enable schema validation of 'Chart.yaml' using Yamale (default true)
      --validate-maintainers                 Enable validation of maintainer account names in chart.yml.
                                             Works for GitHub, GitLab, and Bitbucket (default true)
      --validate-yaml                        Enable linting of 'Chart.yaml' and values files (default true)
```

### SEE ALSO

* [ct](ct.md)	 - The Helm chart testing tool

