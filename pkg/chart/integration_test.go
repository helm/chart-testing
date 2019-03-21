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

// +build integration

package chart

import (
	"fmt"
	"strings"
	"testing"

	"github.com/helm/chart-testing/pkg/config"
	"github.com/helm/chart-testing/pkg/exec"
	"github.com/helm/chart-testing/pkg/tool"
	"github.com/helm/chart-testing/pkg/util"
	"github.com/stretchr/testify/assert"
)

func newTestingHelmIntegration(cfg config.Configuration) Testing {
	fakeMockLinter := new(fakeLinter)
	procExec := exec.NewProcessExecutor(true)
	extraArgs := strings.Fields(cfg.HelmExtraArgs)
	return Testing{
		config:           cfg,
		directoryLister:  util.DirectoryLister{},
		git:              fakeGit{},
		chartUtils:       util.ChartUtils{},
		accountValidator: fakeAccountValidator{},
		linter:           fakeMockLinter,
		helm:             tool.NewHelm(procExec, extraArgs),
		kubectl:          tool.NewKubectl(procExec),
	}
}

func TestInstallChart(t *testing.T) {
	type testCase struct {
		name     string
		cfg      config.Configuration
		chartDir string
		output   TestResult
	}

	cases := []testCase{
		{
			"install only in custom namespace",
			config.Configuration{
				Debug:        true,
				Namespace:    "default",
				ReleaseLabel: "app.kubernetes.io/instance",
			},
			"test_charts/must-pass-upgrade-install",
			TestResult{mustNewChart("test_charts/must-pass-upgrade-install"), nil},
		},
		{
			"install only in random namespace",
			config.Configuration{
				Debug: true,
			},
			"test_charts/must-pass-upgrade-install",
			TestResult{mustNewChart("test_charts/must-pass-upgrade-install"), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := newTestingHelmIntegration(tc.cfg)
			result := ct.InstallChart(mustNewChart(tc.chartDir))

			if result.Error != tc.output.Error {
				if result.Error != nil && tc.output.Error != nil {
					assert.Equal(t, tc.output.Error.Error(), result.Error.Error())
				} else {
					assert.Equal(t, tc.output.Error, result.Error)
				}
			}
		})
	}
}

func TestUpgradeChart(t *testing.T) {
	type testCase struct {
		name string
		old  string
		new  string
		err  error
	}

	cfg := config.Configuration{
		Debug:   true,
		Upgrade: true,
	}
	ct := newTestingHelmIntegration(cfg)
	processError := fmt.Errorf("Error waiting for process: exit status 1")

	cases := []testCase{
		{
			"upgrade nginx",
			"test_charts/must-pass-upgrade-install",
			"test_charts/must-pass-upgrade-install",
			nil,
		},
		{
			"change immutable deployment.spec.selector field",
			"test_charts/mutating-deployment-selector",
			"test_charts/mutating-deployment-selector",
			processError,
		},
		{
			"change immutable statefulset.spec.volumeClaimTemplates field",
			"test_charts/mutating-sfs-volumeclaim",
			"test_charts/mutating-sfs-volumeclaim",
			processError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ct.doUpgrade(mustNewChart(tc.old), mustNewChart(tc.new), true)

			if err != tc.err {
				if err != nil && tc.err != nil {
					assert.Equal(t, tc.err.Error(), err.Error())
				} else {
					assert.Equal(t, tc.err, err)
				}
			}
		})
	}
}

func mustNewChart(chartPath string) *Chart {
	c, err := NewChart(chartPath)
	if err != nil {
		panic(err)
	}
	return c
}
