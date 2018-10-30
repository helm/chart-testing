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
)

func newLintAndInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lint-and-install",
		Aliases: []string{"li"},
		Short:   "Lint, install, and test a chart",
		Long: heredoc.Doc(`
			        __
			  _____/ /_
			 / ___/ __/
			/ /__/ /_
			\___/\__/

			Combines 'lint' and 'install' commands.`),
		Run: lintAndInstall,
	}

	flags := cmd.Flags()
	addLintFlags(flags)
	addInstallFlags(flags)
	addCommonLintAndInstallFlags(flags)
	return cmd
}

func lintAndInstall(cmd *cobra.Command, args []string) {
	fmt.Println("Linting and installing charts...")

	configuration, err := config.LoadConfiguration(cfgFile, cmd, bindRootFlags, bindLintFlags, bindInstallFlags)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testing := chart.NewTesting(*configuration)
	results, err := testing.LintAndInstallCharts()
	if err != nil {
		fmt.Println("Error linting and installing charts")
	} else {
		fmt.Println("All charts linted and installed successfully")
	}

	testing.PrintResults(results)

	if err != nil {
		os.Exit(1)
	}
}
