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

func newListChangedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-changed",
		Aliases: []string{"ls-changed", "lsc"},
		Short:   "List changed charts",
		Long: heredoc.Doc(`
			"List changed charts based on configured charts directories,
			"remote, and target branch`),
		Run: listChanged,
	}

	flags := cmd.Flags()
	addCommonFlags(flags)
	return cmd
}

func listChanged(cmd *cobra.Command, args []string) {
	configuration, err := config.LoadConfiguration(cfgFile, cmd, false)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	testing := chart.NewTesting(*configuration)
	chartDirs, err := testing.ComputeChangedChartDirectories()
	if err != nil {
		os.Exit(1)
	}

	for _, dir := range chartDirs {
		fmt.Println(dir)
	}
}
