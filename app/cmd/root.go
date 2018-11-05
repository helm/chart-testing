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

	"github.com/spf13/viper"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	flag "github.com/spf13/pflag"
)

var (
	cfgFile string
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ct",
		Short: "The Helm chart testing tool",
		Long: heredoc.Doc(`
			Lint and test

			* changed charts
			* specific charts
			* all charts

			in given chart directories.`),
	}

	cmd.AddCommand(newLintCmd())
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newLintAndInstallCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newGenerateDocsCmd())

	return cmd
}

// Execute runs the application
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addCommonLintAndInstallFlags(flags *pflag.FlagSet) {
	flags.StringVar(&cfgFile, "config", "", "Config file")
	flags.String("remote", "origin", "The name of the Git remote used to identify changed charts")
	flags.String("target-branch", "master", "The name of the target branch used to identify changed charts")
	flags.Bool("all", false, heredoc.Doc(`
		Process all charts except those explicitly excluded.
		Disables changed charts detection and version increment checking`))
	flags.StringSlice("charts", []string{}, heredoc.Doc(`
		Specific charts to test. Disables changed charts detection and
		version increment checking. May be specified multiple times
		or separate values with commas`))
	flags.StringSlice("chart-dirs", []string{"charts"}, heredoc.Doc(`
		Directories containing Helm charts. May be specified multiple times
		or separate values with commas`))
	flags.StringSlice("chart-repos", []string{}, heredoc.Doc(`
		Additional chart repos to add so dependencies can be resolved. May be
		specified multiple times or separate values with commas`))
	flags.StringSlice("excluded-charts", []string{}, heredoc.Doc(`
		Charts that should be skipped. May be specified multiple times
		or separate values with commas`))
}

func bindFlags(options []string, flagSet *flag.FlagSet, v *viper.Viper) error {
	for _, option := range options {
		if err := v.BindPFlag(option, flagSet.Lookup(option)); err != nil {
			return err
		}
	}
	return nil
}

func bindRootFlags(flagSet *flag.FlagSet, v *viper.Viper) error {
	options := []string{"remote", "target-branch", "all", "charts", "chart-dirs", "chart-repos", "excluded-charts"}
	return bindFlags(options, flagSet, v)
}
