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
)

func newListChangedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-changed",
		Aliases: []string{"ls-changed", "lsc"},
		Short:   "List changed charts",
		Long: heredoc.Doc(`
			"List changed charts based on configured charts directories,
			"remote, and target branch`),
		RunE: listChanged,
	}

	flags := cmd.Flags()
	addCommonFlags(flags)
	return cmd
}

func listChanged(cmd *cobra.Command, _ []string) error {
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
	chartDirs, err := testing.ComputeChangedChartDirectories()
	if err != nil {
		return err
	}

	for _, dir := range chartDirs {
		fmt.Println(dir)
	}

	return nil
}
