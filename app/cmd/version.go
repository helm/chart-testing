// Copyright © 2018 The Helm Authors
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
	"github.com/spf13/cobra"
)

var (
	// GitCommit is updated with the Git tag by the Goreleaser build
	GitCommit = "unknown"
	// BuildDate is updated with the current ISO timestamp by the Goreleaser build
	BuildDate = "unknown"
	// Version is updated with the latest tag by the Goreleaser build
	Version = "unreleased"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: heredoc.Doc(`
			        __ 
			  _____/ /_
			 / ___/ __/
			/ /__/ /_  
			\___/\__/ 
			
			Print version information.`),
		Run:   version,
	}
}

func version(cmd *cobra.Command, args []string) {
	fmt.Println("Version:\t", Version)
	fmt.Println("Git commit:\t", GitCommit)
	fmt.Println("Date:\t\t", BuildDate)
	fmt.Println("License:\t Apache 2.0")
}
