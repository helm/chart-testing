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

package tool

import (
	"fmt"
	"strings"

	"github.com/helm/chart-testing/pkg/exec"
	"github.com/pkg/errors"
)

type Git struct {
	exec exec.ProcessExecutor
}

func NewGit() Git {
	return Git{exec: exec.ProcessExecutor{}}
}

func (g Git) FileExistsOnBranch(file string, remote string, branch string) bool {
	fileSpec := fmt.Sprintf("%s/%s:%s", remote, branch, file)
	_, err := g.exec.RunProcessAndCaptureOutput("git", "cat-file", "-e", fileSpec)
	return err == nil
}

func (g Git) Show(file string, remote string, branch string) (string, error) {
	fileSpec := fmt.Sprintf("%s/%s:%s", remote, branch, file)
	return g.exec.RunProcessAndCaptureOutput("git", "show", fileSpec)
}

func (g Git) MergeBase(commit1 string, commit2 string) (string, error) {
	return g.exec.RunProcessAndCaptureOutput("git", "merge-base", commit1, commit2)
}

func (g Git) ListChangedFilesInDirs(commit string, dirs ...string) ([]string, error) {
	changedChartFilesString, err :=
		g.exec.RunProcessAndCaptureOutput("git", "diff", "--find-renames", "--name-only", commit, "--", dirs)
	if err != nil {
		return nil, errors.Wrap(err, "Could not determined changed charts: Error creating diff.")
	}
	if changedChartFilesString == "" {
		return nil, nil
	}
	return strings.Split(changedChartFilesString, "\n"), nil
}

func (g Git) GetUrlForRemote(remote string) (string, error) {
	return g.exec.RunProcessAndCaptureOutput("git", "ls-remote", "--get-url", remote)
}
