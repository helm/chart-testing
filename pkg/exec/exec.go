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

package exec

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/helm/chart-testing/pkg/util"
	"github.com/pkg/errors"
)

type ProcessExecutor struct{}

func (p ProcessExecutor) RunProcessAndCaptureOutput(executable string, execArgs ...interface{}) (string, error) {
	return p.RunProcessInDirAndCaptureOutput("", executable, execArgs)
}

func (p ProcessExecutor) RunProcessInDirAndCaptureOutput(workingDirectory string, executable string, execArgs ...interface{}) (string, error) {
	args, err := util.Flatten(execArgs)
	fmt.Println(">>>", executable, strings.Join(args, " "))
	if err != nil {
		return "", errors.Wrap(err, "Invalid arguments supplied")
	}
	cmd := exec.Command(executable, args...)
	cmd.Dir = workingDirectory
	bytes, err := cmd.CombinedOutput()

	if err != nil {
		return "", errors.Wrap(err, "Error running process")
	}
	return strings.TrimSpace(string(bytes)), nil
}

func (p ProcessExecutor) RunProcess(executable string, execArgs ...interface{}) error {
	args, err := util.Flatten(execArgs)
	fmt.Println(">>>", executable, strings.Join(args, " "))
	if err != nil {
		return errors.Wrap(err, "Invalid arguments supplied")
	}
	cmd := exec.Command(executable, args...)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "Error getting StdoutPipe for command")
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "Error getting StderrPipe for command")
	}

	scanner := bufio.NewScanner(io.MultiReader(outReader, errReader))
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "Error running process")
	}

	err = cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "Error waiting for process")
	}

	return nil
}
