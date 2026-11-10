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

package exec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/helm/chart-testing/v3/pkg/util"
)

type ProcessExecutor struct {
	debug bool
}

func NewProcessExecutor(debug bool) ProcessExecutor {
	return ProcessExecutor{
		debug: debug,
	}
}

func (p ProcessExecutor) RunProcessAndCaptureOutput(executable string, execArgs ...interface{}) (string, error) {
	return p.RunProcessInDirAndCaptureOutput("", executable, execArgs)
}

func (p ProcessExecutor) RunProcessAndCaptureStdout(executable string, execArgs ...interface{}) (string, error) {
	return p.RunProcessInDirAndCaptureStdout("", executable, execArgs)
}

func (p ProcessExecutor) RunProcessInDirAndCaptureOutput(workingDirectory string, executable string, execArgs ...interface{}) (string, error) {
	cmd, err := p.CreateProcess(executable, execArgs...)
	if err != nil {
		return "", err
	}

	cmd.Dir = workingDirectory
	bytes, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("failed running process: %w", err)
	}
	return strings.TrimSpace(string(bytes)), nil
}

func (p ProcessExecutor) RunProcessInDirAndCaptureStdout(workingDirectory string, executable string, execArgs ...interface{}) (string, error) {
	cmd, err := p.CreateProcess(executable, execArgs...)
	if err != nil {
		return "", err
	}

	cmd.Dir = workingDirectory
	bytes, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed running process: %w", err)
	}
	return strings.TrimSpace(string(bytes)), nil
}

// outputDrainTimeout bounds how long RunProcess keeps reading output after
// the command has exited. Background processes started by the command inherit
// its stdout and stderr and can keep the pipe open indefinitely.
const outputDrainTimeout = time.Second

// processOutput forwards the combined stdout and stderr of a command and
// remembers the first write error. It is deliberately not an *os.File, so
// exec copies the output through a pipe that Wait closes.
type processOutput struct {
	w   io.Writer
	err error
}

func (o *processOutput) Write(b []byte) (int, error) {
	n, err := o.w.Write(b)
	if err != nil && o.err == nil {
		o.err = err
	}
	return n, err
}

func (p ProcessExecutor) RunProcess(executable string, execArgs ...interface{}) error {
	cmd, err := p.CreateProcess(executable, execArgs...)
	if err != nil {
		return err
	}

	// Using the same writer for stdout and stderr makes exec send both to a
	// single pipe, so neither can fill up while the other one is being read.
	// If writing fails, exec closes the read end and the command gets EPIPE
	// instead of blocking on a full pipe.
	out := &processOutput{w: os.Stdout}
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.WaitDelay = outputDrainTimeout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed running process: %w", err)
	}

	err = cmd.Wait()
	if out.err != nil {
		return fmt.Errorf("failed writing process output: %w", out.err)
	}
	if err != nil && !errors.Is(err, exec.ErrWaitDelay) {
		return fmt.Errorf("failed waiting for process: %w", err)
	}

	return nil
}

func (p ProcessExecutor) CreateProcess(executable string, execArgs ...interface{}) (*exec.Cmd, error) {
	args, err := util.Flatten(execArgs)
	if p.debug {
		fmt.Println(">>>", executable, strings.Join(args, " "))
	}
	if err != nil {
		return nil, fmt.Errorf("invalid arguments supplied: %w", err)
	}
	cmd := exec.Command(executable, args...)

	return cmd, nil
}

type fn func(port int) error

func (p ProcessExecutor) RunWithProxy(withProxy fn) error {
	randomPort, err := util.GetRandomPort()
	if err != nil {
		return fmt.Errorf("could not find a free port for running 'kubectl proxy': %w", err)
	}

	fmt.Printf("Running 'kubectl proxy' on port %d\n", randomPort)
	cmdProxy, err := p.CreateProcess("kubectl", "proxy", fmt.Sprintf("--port=%d", randomPort))
	if err != nil {
		return fmt.Errorf("failed creating the 'kubectl proxy' process: %w", err)
	}
	err = cmdProxy.Start()
	if err != nil {
		return fmt.Errorf("failed starting the 'kubectl proxy' process: %w", err)
	}

	err = withProxy(randomPort)

	_ = cmdProxy.Process.Signal(os.Kill)

	if err != nil {
		return fmt.Errorf("failed running command with proxy: %w", err)
	}

	return nil
}
