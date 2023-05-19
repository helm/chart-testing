// Copyright The Helm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/helm/chart-testing/v3/pkg/chart"
	"github.com/helm/chart-testing/v3/pkg/config"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func newLintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint and validate a chart",
		Long: heredoc.Doc(`
			Run 'helm lint', version checking, YAML schema validation
			on 'Chart.yaml', YAML linting on 'Chart.yaml' and 'values.yaml',
			and maintainer validation on

			* changed charts (default)
			* specific charts (--charts)
			* all charts (--all)

			in given chart directories.

			Charts may have multiple custom values files matching the glob pattern
			'*-values.yaml' in a directory named 'ci' in the root of the chart's
			directory. The chart is linted for each of these files. If no custom
			values file is present, the chart is linted with defaults.`),
		RunE: lint,
	}

	flags := cmd.Flags()
	addLintFlags(flags)
	addCommonLintAndInstallFlags(flags)
	return cmd
}

func addLintFlags(flags *flag.FlagSet) {
	flags.String("lint-conf", "", heredoc.Doc(`
		The config file for YAML linting. If not specified, 'lintconf.yaml'
		is searched in the current directory, '$HOME/.ct', and '/etc/ct', in
		that order`))
	flags.String("chart-yaml-schema", "", heredoc.Doc(`
		The schema for chart.yml validation. If not specified, 'chart_schema.yaml'
		is searched in the current directory, '$HOME/.ct', and '/etc/ct', in
		that order.`))
	flags.Bool("validate-maintainers", true, heredoc.Doc(`
		Enable validation of maintainer account names in chart.yml.
		Works for GitHub, GitLab, and Bitbucket`))
	flags.Bool("check-version-increment", true, "Activates a check for chart version increments")
	flags.Bool("validate-chart-schema", true, heredoc.Doc(`
		Enable schema validation of 'Chart.yaml' using Yamale`))
	flags.Bool("validate-yaml", true, heredoc.Doc(`
		Enable linting of 'Chart.yaml' and values files`))
	flags.StringSlice("additional-commands", []string{}, heredoc.Doc(`
		Additional commands to run per chart (default: [])
		Commands will be executed in the same order as provided in the list and will
		be rendered with go template before being executed.
		Example: "helm unittest --helm3 -f tests/*.yaml {{ .Path }}"`))
}

func lint(cmd *cobra.Command, _ []string) error {
	fmt.Println("Linting charts...")

	printConfig, err := cmd.Flags().GetBool("print-config")
	if err != nil {
		return err
	}
	configuration, err := config.LoadConfiguration(cfgFile, cmd, printConfig)
	if err != nil {
		return fmt.Errorf("failed loading configuration: %w", err)
	}

	emptyExtraSetArgs := ""
	testing, err := chart.NewTesting(*configuration, emptyExtraSetArgs)
	if err != nil {
		return err
	}
	results, err := testing.LintCharts()
	testing.PrintResults(results)

	if err != nil {
		return fmt.Errorf("failed linting charts: %w", err)
	}

	fmt.Println("All charts linted successfully")
	return nil
}
