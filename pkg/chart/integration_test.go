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

//go:build integration
// +build integration

package chart

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/helm/chart-testing/v3/pkg/config"
	"github.com/helm/chart-testing/v3/pkg/exec"
	"github.com/helm/chart-testing/v3/pkg/tool"
	"github.com/helm/chart-testing/v3/pkg/util"
	"github.com/stretchr/testify/assert"
)

func newTestingHelmIntegration(cfg config.Configuration, extraSetArgs string) Testing {
	fakeMockLinter := new(fakeLinter)
	procExec := exec.NewProcessExecutor(true)
	extraArgs := strings.Fields(cfg.HelmExtraArgs)
	extraLintArgs := strings.Fields(cfg.HelmLintExtraArgs)

	return Testing{
		config:           cfg,
		directoryLister:  util.DirectoryLister{},
		git:              fakeGit{},
		utils:            util.Utils{},
		accountValidator: fakeAccountValidator{},
		linter:           fakeMockLinter,
		helm:             tool.NewHelm(procExec, extraArgs, extraLintArgs, strings.Fields(extraSetArgs)),
		kubectl:          tool.NewKubectl(procExec, 30*time.Second),
	}
}

func TestInstallChart(t *testing.T) {
	type testCase struct {
		name     string
		cfg      config.Configuration
		chartDir string
		output   TestResult
		extraSet string
	}

	cases := []testCase{
		{
			"install only in custom namespace",
			config.Configuration{
				Debug:        true,
				Namespace:    "foobar",
				ReleaseLabel: "app.kubernetes.io/instance",
			},
			"test_charts/must-pass-upgrade-install",
			TestResult{mustNewChart("test_charts/must-pass-upgrade-install"), nil},
			"",
		},
		{
			"install only in random namespace",
			config.Configuration{
				Debug: true,
			},
			"test_charts/must-pass-upgrade-install",
			TestResult{mustNewChart("test_charts/must-pass-upgrade-install"), nil},
			"",
		},
		{
			"install with override set",
			config.Configuration{
				Debug: true,
			},
			"test_charts/must-pass-upgrade-install",
			TestResult{mustNewChart("test_charts/must-pass-upgrade-install"), nil},
			"--set=image.tag=latest",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ct := newTestingHelmIntegration(tc.cfg, tc.extraSet)
			namespace := tc.cfg.Namespace
			if namespace != "" {
				ct.kubectl.CreateNamespace(namespace)
				defer ct.kubectl.DeleteNamespace(namespace)
			}
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
	ct := newTestingHelmIntegration(cfg, "")
	processError := fmt.Errorf("failed waiting for process: exit status 1")

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
		{
			"change immutable deployment.spec.selector.matchLabels field",
			"test_charts/simple-deployment",
			"test_charts/simple-deployment-different-selector",
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
