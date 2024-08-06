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

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and test a chart",
		Long: heredoc.Doc(`
			Run 'helm install', 'helm test', and optionally 'helm upgrade' on

			* changed charts (default)
			* specific charts (--charts)
			* all charts (--all)

			in given chart directories. If upgrade (--upgrade) is true, then this
			command will validate that 'helm test' passes for the following upgrade paths:

			* previous chart revision => current chart version (if non-breaking SemVer change)
			* current chart version => current chart version

			Charts may have multiple custom values files matching the glob pattern
			'*-values.yaml' in a directory named 'ci' in the root of the chart's
			directory. The chart is installed and tested for each of these files.
			If no custom values file is present, the chart is installed and
			tested with defaults.`),
		RunE: install,
	}

	flags := cmd.Flags()
	addInstallFlags(flags)
	addCommonLintAndInstallFlags(flags)
	return cmd
}

func addInstallFlags(flags *flag.FlagSet) {
	flags.String("build-id", "", heredoc.Doc(`
		An optional, arbitrary identifier that is added to the name of the namespace a
		chart is installed into. In a CI environment, this could be the build number or
		the ID of a pull request. If not specified, the name of the chart is used`))
	flags.Bool("upgrade", false, heredoc.Doc(`
		Whether to test an in-place upgrade of each chart from its previous revision if the
		current version should not introduce a breaking change according to the SemVer spec`))
	flags.Bool("skip-missing-values", false, heredoc.Doc(`
		When --upgrade has been passed, this flag will skip testing CI values files from the
		previous chart revision if they have been deleted or renamed at the current chart
		revision`))
	flags.String("namespace", "", heredoc.Doc(`
		Namespace to install the release(s) into. If not specified, each release will be
		installed in its own randomly generated namespace`))
	flags.String("release-label", "app.kubernetes.io/instance", heredoc.Doc(`
		The label to be used as a selector when inspecting resources created by charts.
		This is only used if namespace is specified`))
	flags.String("helm-extra-set-args", "", heredoc.Doc(`
		Additional arguments for Helm. Must be passed as a single quoted string
		(e.g. "--set=name=value"`))
	flags.Bool("skip-clean-up", false, heredoc.Doc(`
		Skip resources clean-up. Used if need to continue other flows or keep it around.`))
	flags.String("upgrade-strategy", "reuse-values", heredoc.Doc(`
		When --upgrade is used, this flag configures the upgrade strategy for chart values.
		Available options: reset-values, reuse-values, reset-then-reuse-values`))
}

func install(cmd *cobra.Command, _ []string) error {
	fmt.Println("Installing charts...")

	printConfig, err := cmd.Flags().GetBool("print-config")
	if err != nil {
		return err
	}
	configuration, err := config.LoadConfiguration(cfgFile, cmd, printConfig)
	if err != nil {
		return fmt.Errorf("failed loading configuration: %w", err)
	}

	extraSetArgs, err := cmd.Flags().GetString("helm-extra-set-args")
	if err != nil {
		return err
	}
	testing, err := chart.NewTesting(*configuration, extraSetArgs)
	if err != nil {
		fmt.Println(err)
	}
	results, err := testing.InstallCharts()
	testing.PrintResults(results)

	if err != nil {
		return fmt.Errorf("failed installing charts: %w", err)
	}

	fmt.Println("All charts installed successfully")
	return nil
}
