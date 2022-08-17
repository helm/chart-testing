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

	"github.com/helm/chart-testing/v3/pkg/chart"
	"github.com/helm/chart-testing/v3/pkg/config"

	"github.com/spf13/cobra"
)

func newLintAndInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lint-and-install",
		Aliases: []string{"li"},
		Short:   "Lint, install, and test a chart",
		Long:    "Combines 'lint' and 'install' commands.",
		RunE:    lintAndInstall,
	}

	flags := cmd.Flags()
	addLintFlags(flags)
	addInstallFlags(flags)
	addCommonLintAndInstallFlags(flags)
	return cmd
}

func lintAndInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Linting and installing charts...")

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
		return err
	}
	results, err := testing.LintAndInstallCharts()
	testing.PrintResults(results)

	if err != nil {
		return fmt.Errorf("failed linting and installing charts: %w", err)
	}

	fmt.Println("All charts linted and installed successfully")
	return nil
}
