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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// runProcess runs a shell script through RunProcess and fails the test if it
// does not return within the timeout.
func runProcess(t *testing.T, timeout time.Duration, script string) error {
	t.Helper()

	p := NewProcessExecutor(false)
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.RunProcess("sh", "-c", script)
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(timeout):
		t.Fatal("RunProcess did not return")
		return nil
	}
}

func TestRunProcessDoesNotBlock(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			// More stderr than a pipe buffer holds, written before any stdout.
			name:   "large stderr",
			script: `yes warning | head -n 200000 >&2; echo done`,
		},
		{
			// A single line longer than bufio.Scanner's default limit.
			name:   "long line",
			script: `head -c 1048576 /dev/zero | tr '\0' 'x'; echo; echo done`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runProcess(t, 30*time.Second, tt.script); err != nil {
				t.Fatalf("RunProcess returned an error: %v", err)
			}
		})
	}
}

// startBackgroundSleep returns a script fragment that starts a long sleep in
// the background, inheriting the command's stdout and stderr, and kills it
// when the test ends.
func startBackgroundSleep(t *testing.T) string {
	t.Helper()

	pidFile := filepath.Join(t.TempDir(), "pid")
	t.Cleanup(func() {
		b, err := os.ReadFile(pidFile)
		if err != nil {
			return
		}
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	})
	return "sleep 60 & echo $! > " + pidFile + ";"
}

func TestRunProcessDoesNotWaitForBackgroundChildren(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr bool
	}{
		{
			name:   "command succeeds",
			script: "echo started",
		},
		{
			name:    "command fails",
			script:  "exit 3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := startBackgroundSleep(t) + tt.script
			start := time.Now()
			err := runProcess(t, 20*time.Second, script)
			if tt.wantErr != (err != nil) {
				t.Fatalf("RunProcess returned error %v, want error: %t", err, tt.wantErr)
			}
			if elapsed := time.Since(start); elapsed > 10*time.Second {
				t.Fatalf("RunProcess took %s, it waited for the background child", elapsed)
			}
		})
	}
}

func TestRunProcessReturnsWhenStdoutIsBroken(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	stdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = stdout
		writer.Close()
	})

	// Much more than a pipe buffer holds, so the command blocks if nothing
	// drains or closes its end of the pipe.
	err = runProcess(t, 20*time.Second, `head -c 2097152 /dev/zero`)
	if !errors.Is(err, syscall.EPIPE) {
		t.Fatalf("RunProcess returned error %v, want the broken pipe error from writing its output", err)
	}
}
