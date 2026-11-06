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
	"testing"
	"time"
)

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

	p := NewProcessExecutor(false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCh := make(chan error, 1)
			go func() {
				errCh <- p.RunProcess("sh", "-c", tt.script)
			}()

			select {
			case err := <-errCh:
				if err != nil {
					t.Fatalf("RunProcess returned an error: %v", err)
				}
			case <-time.After(30 * time.Second):
				t.Fatal("RunProcess did not return")
			}
		})
	}
}
