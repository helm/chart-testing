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
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/helm/chart-testing/v3/pkg/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitListChangedFilesPreservesPaths(t *testing.T) {
	if _, err := osexec.LookPath("git"); err != nil {
		t.Skip("git is required")
	}
	t.Chdir(t.TempDir())
	run := func(args ...string) string {
		t.Helper()
		out, err := osexec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, "%s", out)
		return strings.TrimSpace(string(out))
	}
	run("init", "--quiet")
	run("config", "core.quotepath", "true")
	// Comparing against the empty tree avoids creating any fixture commits.
	emptyTree := run("hash-object", "-w", "-t", "tree", "--stdin")
	git := NewGit(exec.NewProcessExecutor(false))
	files, err := git.ListChangedFilesInDirs(emptyTree, "charts")
	require.NoError(t, err)
	assert.Empty(t, files)

	paths := []string{
		"charts/plain/values.yaml",
		"charts/café/values.yaml",
		"charts/приложение/values.yaml",
		"charts/with space/values.yaml",
		" leading/values.yaml",
	}
	if runtime.GOOS != "windows" {
		paths = append(paths, "charts/with\ttab/values.yaml", "charts/with\nnewline/values.yaml", "charts/with\"quote/values.yaml")
	}
	for _, path := range paths {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte("value: true\n"), 0o600))
	}
	run("add", "--all")
	files, err = git.ListChangedFilesInDirs(emptyTree, "charts", " leading")
	require.NoError(t, err)
	assert.ElementsMatch(t, paths, files)

	_, err = git.ListChangedFilesInDirs("missing-revision", "charts")
	require.ErrorContains(t, err, "failed creating diff")
}
