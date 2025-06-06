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

	"github.com/helm/chart-testing/v3/pkg/config"
	"github.com/helm/chart-testing/v3/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	helmignore "helm.sh/helm/v3/pkg/ignore"
)

type fakeGit struct{}

func (g fakeGit) FileExistsOnBranch(_ string, _ string, _ string) bool {
	return true
}

func (g fakeGit) Show(_ string, _ string, _ string) (string, error) {
	return "", nil
}

func (g fakeGit) MergeBase(_ string, _ string) (string, error) {
	return "HEAD", nil
}

func (g fakeGit) ListChangedFilesInDirs(_ string, _ ...string) ([]string, error) {
	return []string{
		"test_charts/foo/Chart.yaml",
		"test_charts/bar/Chart.yaml",
		"test_charts/bar/bar_sub/templates/bar_sub.yaml",
		"test_charts/excluded/Chart.yaml",
		"test_chart_at_root/templates/foo.yaml",
		"test_chart_at_multi_level/foo/bar/Chart.yaml",
		"test_chart_at_multi_level/foo/baz/Chart.yaml",
		"test_chart_at_multi_level/foo/excluded/Chart.yaml",
		"some_non_chart_dir/some_non_chart_file",
		"some_non_chart_file",
	}, nil
}

func (g fakeGit) AddWorktree(_ string, _ string) error {
	return nil
}

func (g fakeGit) RemoveWorktree(_ string) error {
	return nil
}

func (g fakeGit) GetURLForRemote(_ string) (string, error) {
	return "git@github.com/helm/chart-testing", nil
}

func (g fakeGit) ValidateRepository() error {
	return nil
}

func (g fakeGit) BranchExists(_ string) bool {
	return true
}

type fakeAccountValidator struct{}

func (v fakeAccountValidator) Validate(_ string, account string) error {
	if strings.HasPrefix(account, "valid") {
		return nil
	}
	return fmt.Errorf("failed validating account: %s", account)
}

type fakeLinter struct {
	mock.Mock
}

func (l *fakeLinter) YamlLint(yamlFile, configFile string) error {
	l.Called(yamlFile, configFile)
	return nil
}
func (l *fakeLinter) Yamale(yamlFile, schemaFile string) error {
	l.Called(yamlFile, schemaFile)
	return nil
}

type fakeHelm struct {
	mock.Mock
}

func (h *fakeHelm) AddRepo(_, _ string, _ []string) error { return nil }
func (h *fakeHelm) BuildDependencies(_ string) error      { return nil }
func (h *fakeHelm) BuildDependenciesWithArgs(chart string, extraArgs []string) error {
	h.Called(chart, extraArgs)
	return nil
}
func (h *fakeHelm) LintWithValues(_ string, _ string) error { return nil }
func (h *fakeHelm) InstallWithValues(_ string, _ string, _ string, _ string) error {
	return nil
}
func (h *fakeHelm) UpgradeWithValues(_ string, _ string, _ string, _ string) error {
	return nil
}
func (h *fakeHelm) Test(_ string, _ string) error {
	return nil
}
func (h *fakeHelm) DeleteRelease(_ string, _ string) {}

func (h *fakeHelm) Version() (string, error) {
	return "v3.0.0", nil
}

type fakeCmdExecutor struct {
	mock.Mock
}

func (c *fakeCmdExecutor) RunCommand(cmdTemplate string, data interface{}) error {
	c.Called(cmdTemplate, data)
	return nil
}

var ct Testing

func init() {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"test_charts", "."},
	}

	ct = newTestingMock(cfg)
}

func newTestingMock(cfg config.Configuration) Testing {
	fakeMockLinter := new(fakeLinter)
	return Testing{
		config:           cfg,
		directoryLister:  util.DirectoryLister{},
		git:              fakeGit{},
		utils:            util.Utils{},
		accountValidator: fakeAccountValidator{},
		linter:           fakeMockLinter,
		helm:             new(fakeHelm),
		loadRules: func(dir string) (*helmignore.Rules, error) {
			rules := helmignore.Empty()
			if dir == "test_charts/foo" {
				var err error
				rules, err = helmignore.Parse(strings.NewReader("Chart.yaml\n"))
				if err != nil {
					return nil, err
				}
				rules.AddDefaults()
			}
			if dir == "test_chart_at_multi_level/foo/baz" {
				var err error
				rules, err = helmignore.Parse(strings.NewReader("Chart.yaml\n"))
				if err != nil {
					return nil, err
				}
				rules.AddDefaults()
			}
			return rules, nil
		},
	}
}

func TestComputeChangedChartDirectories(t *testing.T) {
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"test_charts/foo", "test_charts/bar", "test_chart_at_root"}
	for _, chart := range actual {
		assert.Contains(t, expected, chart)
	}
	assert.Len(t, actual, 3)
	assert.Nil(t, err)
}

func TestComputeChangedChartDirectoriesWithHelmignore(t *testing.T) {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"test_charts", "."},
		UseHelmignore:  true,
	}
	ct := newTestingMock(cfg)
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"test_charts/bar", "test_chart_at_root"}
	assert.Nil(t, err)
	assert.ElementsMatch(t, expected, actual)
}

func TestComputeChangedChartDirectoriesWithMultiLevelChart(t *testing.T) {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"test_chart_at_multi_level/foo"},
	}
	ct := newTestingMock(cfg)
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"test_chart_at_multi_level/foo/bar", "test_chart_at_multi_level/foo/baz"}
	for _, chart := range actual {
		assert.Contains(t, expected, chart)
	}
	assert.Len(t, actual, 2)
	assert.Nil(t, err)
}

func TestComputeChangedChartDirectoriesWithMultiLevelChartWithHelmIgnore(t *testing.T) {
	cfg := config.Configuration{
		ExcludedCharts: []string{"excluded"},
		ChartDirs:      []string{"test_chart_at_multi_level/foo"},
		UseHelmignore:  true,
	}
	ct := newTestingMock(cfg)
	actual, err := ct.ComputeChangedChartDirectories()
	expected := []string{"test_chart_at_multi_level/foo/bar"}
	assert.Nil(t, err)
	assert.ElementsMatch(t, expected, actual)
}

func TestReadAllChartDirectories(t *testing.T) {
	actual, err := ct.ReadAllChartDirectories()
	expected := []string{
		"test_charts/foo",
		"test_charts/bar",
		"test_charts/must-pass-upgrade-install",
		"test_charts/mutating-deployment-selector",
		"test_charts/simple-deployment",
		"test_charts/simple-deployment-different-selector",
		"test_charts/mutating-sfs-volumeclaim",
		"test_chart_at_root",
	}
	for _, chart := range actual {
		assert.Contains(t, expected, chart)
	}
	assert.Len(t, actual, 8)
	assert.Nil(t, err)
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
			chart, err := NewChart(testData.chartDir)
			assert.Nil(t, err)
			validationErr := ct.ValidateMaintainers(chart)
			assert.Equal(t, testData.expected, validationErr == nil)
		})
	}
}

func TestLintChartMaintainerValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(validate bool) {
		ct.config.ValidateMaintainers = validate

		var suffix string
		if validate {
			suffix = "with-validation"
		} else {
			suffix = "without-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("maintainers-%s", suffix), "testdata/valid_maintainers", true},
			{fmt.Sprintf("no-maintainers-%s", suffix), "testdata/no_maintainers", !validate},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				chart, err := NewChart(testData.chartDir)
				assert.Nil(t, err)
				result := ct.LintChart(chart)
				assert.Equal(t, testData.expected, result.Error == nil)
			})
		}
	}

	runTests(true)
	runTests(false)
}

func TestLintChartSchemaValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(validate bool, callsYamlLint int, callsYamale int) {
		fakeMockLinter := new(fakeLinter)

		fakeMockLinter.On("Yamale", mock.Anything, mock.Anything).Return(true)
		fakeMockLinter.On("YamlLint", mock.Anything, mock.Anything).Return(true)

		ct.linter = fakeMockLinter
		ct.config.ValidateChartSchema = validate
		ct.config.ValidateMaintainers = false
		ct.config.ValidateYaml = false

		var suffix string
		if validate {
			suffix = "with-validation"
		} else {
			suffix = "without-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("schema-%s", suffix), "testdata/test_lints", true},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				chart, err := NewChart(testData.chartDir)
				assert.Nil(t, err)
				result := ct.LintChart(chart)
				assert.Equal(t, testData.expected, result.Error == nil)
				fakeMockLinter.AssertNumberOfCalls(t, "Yamale", callsYamale)
				fakeMockLinter.AssertNumberOfCalls(t, "YamlLint", callsYamlLint)
			})
		}
	}

	runTests(true, 0, 1)
	runTests(false, 0, 0)
}

func TestLintYamlValidation(t *testing.T) {
	type testData struct {
		name     string
		chartDir string
		expected bool
	}

	runTests := func(validate bool, callsYamlLint int, callsYamale int) {
		fakeMockLinter := new(fakeLinter)

		fakeMockLinter.On("Yamale", mock.Anything, mock.Anything).Return(true)
		fakeMockLinter.On("YamlLint", mock.Anything, mock.Anything).Return(true)

		ct.linter = fakeMockLinter
		ct.config.ValidateYaml = validate
		ct.config.ValidateChartSchema = false
		ct.config.ValidateMaintainers = false

		var suffix string
		if validate {
			suffix = "with-validation"
		} else {
			suffix = "without-validation"
		}

		testCases := []testData{
			{fmt.Sprintf("lint-%s", suffix), "testdata/test_lints", true},
		}

		for _, testData := range testCases {
			t.Run(testData.name, func(t *testing.T) {
				chart, err := NewChart(testData.chartDir)
				assert.Nil(t, err)
				result := ct.LintChart(chart)
				assert.Equal(t, testData.expected, result.Error == nil)
				fakeMockLinter.AssertNumberOfCalls(t, "Yamale", callsYamale)
				fakeMockLinter.AssertNumberOfCalls(t, "YamlLint", callsYamlLint)
			})
		}
	}

	runTests(true, 2, 0)
	runTests(false, 0, 0)
}

func TestLintDependencyExtraArgs(t *testing.T) {
	chart := "testdata/test_lints"
	args := []string{"--skip-refresh"}

	fakeMockHelm := new(fakeHelm)
	ct.helm = fakeMockHelm
	ct.config.HelmDependencyExtraArgs = args
	ct.config.Charts = []string{chart}

	t.Run("lint-helm-dependency-extra-args", func(t *testing.T) {
		call := fakeMockHelm.On("BuildDependenciesWithArgs", chart, args).Return(nil)
		call.Repeatability = 1

		results, err := ct.LintCharts()
		assert.Nil(t, err)
		for _, result := range results {
			assert.Nil(t, result.Error)
		}
		// -1 is set after Repeatability runs out
		assert.Equal(t, -1, call.Repeatability)
	})
}

func TestGenerateInstallConfig(t *testing.T) {
	type testData struct {
		name  string
		cfg   config.Configuration
		chart *Chart
	}

	testCases := []testData{
		{
			"custom namespace",
			config.Configuration{
				Namespace:    "default",
				ReleaseLabel: "app.kubernetes.io/instance",
			},
			&Chart{
				yaml: &util.ChartYaml{
					Name: "bar",
				},
			},
		},
		{
			"random namespace",
			config.Configuration{
				ReleaseLabel: "app.kubernetes.io/instance",
			},
			&Chart{
				yaml: &util.ChartYaml{
					Name: "bar",
				},
			},
		},
		{
			"long chart name",
			config.Configuration{
				ReleaseLabel: "app.kubernetes.io/instance",
			},
			&Chart{
				yaml: &util.ChartYaml{
					Name: "test_charts/barbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbarbar",
				},
			},
		},
	}

	for _, testData := range testCases {
		t.Run(testData.name, func(t *testing.T) {
			ct := newTestingMock(testData.cfg)

			namespace, release, releaseSelector, _ := ct.generateInstallConfig(testData.chart)
			assert.NotEqual(t, "", namespace)
			assert.NotEqual(t, "", release)
			assert.True(t, len(release) < 64, "release should be less than 64 chars")
			assert.True(t, len(namespace) < 64, "namespace should be less than 64 chars")
			if testData.cfg.Namespace != "" {
				assert.Equal(t, testData.cfg.Namespace, namespace)
				assert.Equal(t, fmt.Sprintf("%s=%s", testData.cfg.ReleaseLabel, release), releaseSelector)
			} else {
				assert.Equal(t, "", releaseSelector)
				assert.Contains(t, namespace, release)
			}
		})
	}
}

func TestChart_HasCIValuesFile(t *testing.T) {
	type testData struct {
		name     string
		chart    *Chart
		file     string
		expected bool
	}

	testCases := []testData{
		{
			name: "has file",
			chart: &Chart{
				ciValuesPaths: []string{"foo-values.yaml"},
			},
			file:     "foo-values.yaml",
			expected: true,
		},
		{
			name: "different paths",
			chart: &Chart{
				ciValuesPaths: []string{"ci/foo-values.yaml"},
			},
			file:     "foo/bar/foo-values.yaml",
			expected: true,
		},
		{
			name: "does not have file",
			chart: &Chart{
				ciValuesPaths: []string{"foo-values.yaml"},
			},
			file:     "bar-values.yaml",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.chart.HasCIValuesFile(tc.file)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestChart_AdditionalCommandsAreRun(t *testing.T) {
	type testData struct {
		name            string
		cfg             config.Configuration
		callsRunCommand int
	}

	testCases := []testData{
		{
			name:            "no additional commands",
			cfg:             config.Configuration{},
			callsRunCommand: 0,
		},
		{
			name: "one command",
			cfg: config.Configuration{
				AdditionalCommands: []string{"helm unittest --helm3 -f tests/*.yaml {{ .Path }}"},
			},
			callsRunCommand: 1,
		},
		{
			name: "multiple commands",
			cfg: config.Configuration{
				AdditionalCommands: []string{"echo", "helm unittest --helm3 -f tests/*.yaml {{ .Path }}"},
			},
			callsRunCommand: 2,
		},
	}

	for _, testData := range testCases {
		t.Run(testData.name, func(t *testing.T) {
			fakeCmdExecutor := new(fakeCmdExecutor)
			fakeCmdExecutor.On("RunCommand", mock.Anything, mock.Anything).Return(nil)

			ct := newTestingMock(testData.cfg)
			ct.cmdExecutor = fakeCmdExecutor

			ct.LintChart(&Chart{})

			fakeCmdExecutor.AssertNumberOfCalls(t, "RunCommand", testData.callsRunCommand)
		})
	}
}
