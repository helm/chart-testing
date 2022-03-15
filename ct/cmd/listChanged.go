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
	addListChangedFlags(flags)
	return cmd
}

func addListChangedFlags(flags *flag.FlagSet) {
	flags.Bool("evaluate-dotignore-files", true, heredoc.Doc(`
		If enabled, all .ctingnore & .helmingore files inside the configured
		chart directories are evaluated and applied relative to the file path`))
	flags.StringSlice("considered-dotignore-files", []string{}, heredoc.Doc(`
		Specifies which .ignore files are considered. May be specified multiple
		times or separate values with commas. (default .ctignore,.helmignore)`))
}

func listChanged(cmd *cobra.Command, args []string) error {
	printConfig, err := cmd.Flags().GetBool("print-config")
	if err != nil {
		return err
	}
	configuration, err := config.LoadConfiguration(cfgFile, cmd, printConfig)
	if err != nil {
		return fmt.Errorf("Error loading configuration: %s", err)
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
