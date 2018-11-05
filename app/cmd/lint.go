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
	"os"

	"github.com/MakeNowJust/heredoc"

	"github.com/helm/chart-testing/pkg/chart"
	"github.com/helm/chart-testing/pkg/config"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
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
		Run: lint,
	}

	flags := cmd.Flags()
	addLintFlags(flags)
	addCommonLintAndInstallFlags(flags)
	return cmd
}

func addLintFlags(flags *flag.FlagSet) {
	flags.String("lint-conf", "", heredoc.Doc(`
			The config file for YAML linting. If not specified, 'lintconf.yaml' is
			searched in '/etc/ct', '$HOME/ct', and the current directory`))
	flags.String("chart-yaml-schema", "", heredoc.Doc(`
			The schema for chart.yml validation. If not specified, 'chart_schema.yaml'
			is searched in '/etc/ct', '$HOME/ct', and the current directory`))
	flags.Bool("validate-maintainers", true, heredoc.Doc(`
			Enabled validation of maintainer account names in chart.yml (default: true).
			Works for GitHub, GitLab, and Bitbucket`))
	flags.Bool("check-version-increment", true, "Activates a check for chart version increments (default: true)")
}

func lint(cmd *cobra.Command, args []string) {
	fmt.Println("Linting charts...")

	configuration, err := config.LoadConfiguration(cfgFile, cmd, bindRootFlags, bindLintFlags)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testing := chart.NewTesting(*configuration)
	results, err := testing.LintCharts()
	if err != nil {
		fmt.Println("Error linting charts")
	} else {
		fmt.Println("All charts linted successfully")
	}

	testing.PrintResults(results)

	if err != nil {
		os.Exit(1)
	}
}

func bindLintFlags(flagSet *flag.FlagSet, v *viper.Viper) error {
	options := []string{"lint-conf", "chart-yaml-schema", "validate-maintainers", "check-version-increment"}
	return bindFlags(options, flagSet, v)
}
