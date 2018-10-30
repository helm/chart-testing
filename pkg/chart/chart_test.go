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

package chart

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/helm/chart-testing/pkg/util"

	"github.com/helm/chart-testing/pkg/config"
	"github.com/stretchr/testify/assert"
)

type fakeGit struct{}

func (g fakeGit) FileExistsOnBranch(file string, remote string, branch string) bool {
	return true
}

func (g fakeGit) Show(file string, remote string, branch string) (string, error) {
	return "", nil
}

func (g fakeGit) MergeBase(commit1 string, commit2 string) (string, error) {
	return "", nil
}

func (g fakeGit) ListChangedFilesInDirs(commit string, dirs ...string) ([]string, error) {
	return []string{
		"incubator/excluded/Chart.yaml",
		"incubator/excluded/values.yaml",
		"incubator/bar/README.md",
		"incubator/bar/README.md",
		"incubator/excluded/templates/configmap.yaml",
		"incubator/excluded/values.yaml",
		"stable/blah/Chart.yaml",
		"stable/blah/README.md",
		"stable/this-is-no-chart-dir/foo.md",
	}, nil
}

func (g fakeGit) GetUrlForRemote(remote string) (string, error) {
	return "git@github.com/helm/chart-testing", nil
}

type fakeDirLister struct{}

func (l fakeDirLister) ListChildDirs(parentDir string, test func(dir string) bool) ([]string, error) {
	if parentDir == "stable" {
		var dirs []string
		for _, dir := range []string{"stable/foo", "stable/excluded"} {
			if test(dir) {
				dirs = append(dirs, dir)
			}
		}
		return dirs, nil
	}
	return []string{"incubator/bar"}, nil
}

type fakeChartUtils struct{}

func (v fakeChartUtils) IsChartDir(dir string) bool {
	return dir != "stable/this-is-no-chart-dir"
}

func (v fakeChartUtils) ReadChartYaml(dir string) (*util.ChartYaml, error) {
	chartUtils := util.ChartUtils{}
	return chartUtils.ReadChartYaml(dir)
}

type fakeAccountValidator struct{}

func (v fakeAccountValidator) Validate(repoDomain string, account string) error {
	if strings.HasPrefix(account, "valid") {
		return nil
	}
	return errors.New(fmt.Sprintf("Error validating account: %s", account))
}

var ct Testing

func init() {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"stable", "incubator"},
	}
	ct = Testing{
		config:              cfg,
		directoryLister:     fakeDirLister{},
		git:                 fakeGit{},
		chartUtils:          fakeChartUtils{},
		accountValidator: fakeAccountValidator{},
	}
}

func TestComputeChangedChartDirectories(t *testing.T) {
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"incubator/bar", "stable/blah"}
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestReadAllChartDirectories(t *testing.T) {
	actual, err := ct.ReadAllChartDirectories()
	expected := []string{"stable/foo", "incubator/bar"}
	assert.Nil(t, err)
	assert.Equal(t, actual, expected)
}

func TestValidateMaintainers(t *testing.T) {
	var testDataSlice = []struct {
		name     string
		chartDir string
		expected bool
	}{
		{"valid", "testdata/valid_maintainers", true},
		{"invalid", "testdata/invalid_maintainers", false},
		{"no-maintainers", "testdata/no_maintainers", false},
		{"empty-maintainers", "testdata/empty_maintainers", false},
		{"valid-deprecated", "testdata/valid_maintainers_deprecated", false},
		{"no-maintainers-deprecated", "testdata/no_maintainers_deprecated", true},
	}

	for _, testData := range testDataSlice {
		t.Run(testData.name, func(t *testing.T) {
			err := ct.ValidateMaintainers(testData.chartDir)
			assert.Equal(t, testData.expected, err == nil)
		})
	}
}
