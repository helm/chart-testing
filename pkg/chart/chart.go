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
	"path"
	"path/filepath"
	"strings"

	"github.com/helm/chart-testing/pkg/config"
	"github.com/helm/chart-testing/pkg/tool"
	"github.com/helm/chart-testing/pkg/util"
	"github.com/pkg/errors"
)

// Git is the Interface that wraps Git operations.
//
// FileExistsOnBranch checks whether file exists on the specified remote/branch.
//
// Show returns the contents of file on the specified remote/branch.
//
// MergeBase returns the SHA1 of the merge base of commit1 and commit2.
//
// ListChangedFilesInDirs diffs commit against HEAD and returns changed files for the specified dirs.
//
// GetUrlForRemote returns the repo URL for the specified remote.
type Git interface {
	FileExistsOnBranch(file string, remote string, branch string) bool
	Show(file string, remote string, branch string) (string, error)
	MergeBase(commit1 string, commit2 string) (string, error)
	ListChangedFilesInDirs(commit string, dirs ...string) ([]string, error)
	GetUrlForRemote(remote string) (string, error)
}

// Helm is the interface that wraps Helm operations
//
// Init runs client-side Helm initialization
//
// AddRepo adds a chart repository to the local Helm configuration
//
// BuildDependencies builds the chart's dependencies
//
// Lint runs `helm lint` for the given chart
//
// LintWithValues runs `helm lint` for the given chart using the specified values file
//
// Install runs `helm install` for the given chart
//
// InstallWithValues runs `helm install` for the given chart using the specified values file
//
// DeleteRelease purges the specified Helm release.
type Helm interface {
	Init() error
	AddRepo(name string, url string) error
	BuildDependencies(chart string) error
	Lint(chart string) error
	LintWithValues(chart string, valuesFile string) error
	Install(chart string, namespace string, release string) error
	InstallWithValues(chart string, valuesFile string, namespace string, release string) error
	DeleteRelease(release string)
}

// Kubectl is the interface that wraps kubectl operations
//
// DeleteNamespace deletes a namespace
//
// WaitForDeployments waits for a deployment to become ready
//
// GetPodsforDeployment gets all pods for a deployment
//
// GetPods gets pods for the given args
//
// DescribePod prints the pod's description
//
// Logs prints the logs of container
//
// GetInitContainers gets all init containers of pod
//
// GetContainers gets all containers of pod
type Kubectl interface {
	DeleteNamespace(namespace string)
	WaitForDeployments(namespace string) error
	GetPodsforDeployment(namespace string, deployment string) ([]string, error)
	GetPods(args ...string) ([]string, error)
	DescribePod(namespace string, pod string) error
	Logs(namespace string, pod string, container string) error
	GetInitContainers(namespace string, pod string) ([]string, error)
	GetContainers(namespace string, pod string) ([]string, error)
}

// Linter is the interface that wrap linting operations
//
// YamlLint runs `yamllint` on the specified file with the specified configuration
//
// Yamale runs `yamale` on the specified file with the specified schema file
type Linter interface {
	YamlLint(yamlFile string, configFile string) error
	Yamale(yamlFile string, schemaFile string) error
}

// DiretoryLister is the interface
//
// ListChildDirs lists direct child directories of parentDir given they pass the test function
type DirectoryLister interface {
	ListChildDirs(parentDir string, test func(string) bool) ([]string, error)
}

// ChartUtils is the interface that wraps chart-related methods
//
// IsChartdir checks if a directory is a chart directory
//
// ReadChartYaml reads the `Chart.yaml` from the specified directory
type ChartUtils interface {
	IsChartDir(dir string) bool
	ReadChartYaml(dir string) (*util.ChartYaml, error)
}

// AccountValidator is the interface that wraps Git account validation
//
// Validate checks if account is valid on repoDomain
type AccountValidator interface {
	Validate(repoDomain string, account string) error
}

type Testing struct {
	config           config.Configuration
	helm             Helm
	kubectl          Kubectl
	git              Git
	linter           Linter
	accountValidator AccountValidator
	directoryLister  DirectoryLister
	chartUtils       ChartUtils
}

// TestResults holds results and overall status
type TestResults struct {
	OverallSuccess bool
	TestResults    []TestResult
}

// TestResult holds test results for a specific chart
type TestResult struct {
	Chart string
	Error error
}

// NewTesting creates a new Testing struct with the given config.
func NewTesting(config config.Configuration) Testing {
	kubectl := tool.NewKubectl()
	testing := Testing{
		config:           config,
		helm:             tool.NewHelm(kubectl, config.Timeout, config.TillerNamespace),
		git:              tool.NewGit(),
		kubectl:          kubectl,
		linter:           tool.NewLinter(),
		accountValidator: tool.AccountValidator{},
		directoryLister:  util.DirectoryLister{},
		chartUtils:       util.ChartUtils{},
	}
	return testing
}

func (t *Testing) processCharts(action func(chart string, valuesFiles []string) TestResult) ([]TestResult, error) {
	var results []TestResult
	charts, err := t.FindChartsToBeProcessed()
	if err != nil {
		return nil, errors.Wrap(err, "Error identifying charts to process")
	} else if len(charts) == 0 {
		return results, nil
	}

	fmt.Println()
	util.PrintDelimiterLine("-")
	fmt.Println(" Charts to be processed:")
	util.PrintDelimiterLine("-")
	for _, chart := range charts {
		fmt.Printf(" %s\n", chart)
	}
	util.PrintDelimiterLine("-")
	fmt.Println()

	if err := t.helm.Init(); err != nil {
		return nil, errors.Wrap(err, "Error initializing Helm")
	}

	for _, repo := range t.config.ChartRepos {
		repoSlice := strings.SplitN(repo, "=", 2)
		name := repoSlice[0]
		url := repoSlice[1]
		if err := t.helm.AddRepo(name, url); err != nil {
			return nil, errors.Wrapf(err, "Error adding repo: %s=%s", name, url)
		}
	}

	testResults := TestResults{
		OverallSuccess: true,
		TestResults:    results,
	}

	for _, chart := range charts {
		valuesFiles := t.FindValuesFilesForCI(chart)

		if err := t.helm.BuildDependencies(chart); err != nil {
			return nil, errors.Wrapf(err, "Error building dependencies for chart '%s'", chart)
		}

		result := action(chart, valuesFiles)
		if result.Error != nil {
			testResults.OverallSuccess = false
		}
		results = append(results, result)
	}
	if testResults.OverallSuccess {
		return results, nil
	}

	return results, errors.New("Error processing charts")
}

// LintCharts lints charts (changed, all, specific) depending on the configuration.
func (t *Testing) LintCharts() ([]TestResult, error) {
	return t.processCharts(t.LintChart)
}

// InstallCharts install charts (changed, all, specific) depending on the configuration.
func (t *Testing) InstallCharts() ([]TestResult, error) {
	return t.processCharts(t.InstallChart)
}

// LintAndInstallChart first lints and then installs charts (changed, all, specific) depending on the configuration.
func (t *Testing) LintAndInstallCharts() ([]TestResult, error) {
	return t.processCharts(t.LintAndInstallChart)
}

// PrintResults writes test results to stdout.
func (t *Testing) PrintResults(results []TestResult) {
	util.PrintDelimiterLine("-")
	if results != nil {
		for _, result := range results {
			err := result.Error
			if err != nil {
				fmt.Printf(" %s %s > %s\n", "✖︎", result.Chart, err)
			} else {
				fmt.Printf(" %s %s\n", "✔︎", result.Chart)
			}
		}
	} else {
		fmt.Println("No chart changes detected.")
	}
	util.PrintDelimiterLine("-")
}

// LintChart lints the specified chart.
func (t *Testing) LintChart(chart string, valuesFiles []string) TestResult {
	fmt.Printf("Linting chart '%s'\n", chart)

	result := TestResult{Chart: chart}

	if t.config.CheckVersionIncrement {
		if err := t.CheckVersionIncrement(chart); err != nil {
			result.Error = err
			return result
		}
	}

	chartYaml := path.Join(chart, "Chart.yaml")
	valuesYaml := path.Join(chart, "values.yaml")

	if err := t.linter.Yamale(chartYaml, t.config.ChartYamlSchema); err != nil {
		result.Error = err
		return result
	}
	if err := t.linter.YamlLint(chartYaml, t.config.LintConf); err != nil {
		result.Error = err
		return result
	}
	if err := t.linter.YamlLint(valuesYaml, t.config.LintConf); err != nil {
		result.Error = err
		return result
	}

	if err := t.ValidateMaintainers(chart); err != nil {
		result.Error = err
		return result
	}

	if len(valuesFiles) > 0 {
		for _, valuesFile := range valuesFiles {
			if err := t.helm.LintWithValues(chart, valuesFile); err != nil {
				result.Error = err
				break
			}
		}
	} else {
		if err := t.helm.Lint(chart); err != nil {
			result.Error = err
		}
	}

	return result
}

// InstallChart installs the specified chart into a new namespace, waits for resources to become ready, and eventually
// uninstalls it and deletes the namespace again.
func (t *Testing) InstallChart(chart string, valuesFiles []string) TestResult {
	fmt.Printf("Installing chart '%s'...\n", chart)

	result := TestResult{Chart: chart}

	if len(valuesFiles) > 0 {
		for _, valuesFile := range valuesFiles {
			release, namespace := util.CreateInstallParams(chart, t.config.BuildId)

			defer t.kubectl.DeleteNamespace(namespace)
			defer t.helm.DeleteRelease(release)
			defer t.PrintPodDetailsAndLogs(namespace)

			if err := t.helm.InstallWithValues(chart, valuesFile, namespace, release); err != nil {
				result.Error = err
				break
			}
		}
	} else {
		release, namespace := util.CreateInstallParams(chart, t.config.BuildId)

		defer t.kubectl.DeleteNamespace(namespace)
		defer t.helm.DeleteRelease(release)
		defer t.PrintPodDetailsAndLogs(namespace)

		if err := t.helm.Install(chart, namespace, release); err != nil {
			result.Error = err
		}
	}

	return result
}

// LintAndInstallChart first lints and then installs the specified chart.
func (t *Testing) LintAndInstallChart(chart string, valuesFiles []string) TestResult {
	result := t.LintChart(chart, valuesFiles)
	if result.Error != nil {
		return result
	}
	return t.InstallChart(chart, valuesFiles)
}

// FindChartsToBeProcessed identifies charts to be processed depending on the configuration
// (changed charts, all charts, or specific charts).
func (t *Testing) FindChartsToBeProcessed() ([]string, error) {
	cfg := t.config
	if cfg.ProcessAllCharts {
		return t.ReadAllChartDirectories()
	} else if len(cfg.Charts) > 0 {
		return t.config.Charts, nil
	}
	return t.ComputeChangedChartDirectories()
}

// FindValuesFilesForCI returns all files in the 'ci' subfolder of the chart directory matching the pattern '*-values.yaml'
func (t *Testing) FindValuesFilesForCI(chart string) []string {
	ciDir := path.Join(chart, "ci/*-values.yaml")
	matches, _ := filepath.Glob(ciDir)
	return matches
}

// ComputeChangedChartDirectories takes the merge base of HEAD and the configured remote and target branch and computes a
// slice of changed charts from that in the configured chart directories excluding those configured to be excluded.
func (t *Testing) ComputeChangedChartDirectories() ([]string, error) {
	cfg := t.config

	mergeBase, err := t.git.MergeBase(fmt.Sprintf("%s/%s", cfg.Remote, cfg.TargetBranch), "HEAD")
	if err != nil {
		return nil, errors.Wrap(err, "Could not determined changed charts: Error identifying merge base.")
	}
	allChangedChartFiles, err := t.git.ListChangedFilesInDirs(mergeBase, cfg.ChartDirs...)
	if err != nil {
		return nil, errors.Wrap(err, "Could not determined changed charts: Error icreating diff.")
	}

	var changedChartDirs []string
	for _, file := range allChangedChartFiles {
		pathElements := strings.SplitN(filepath.ToSlash(file), "/", 3)
		if util.StringSliceContains(cfg.ExcludedCharts, pathElements[1]) {
			continue
		}
		dir := path.Join(pathElements[0], pathElements[1])
		// Only add if not already in list and double-check if it is a chart directory
		if !util.StringSliceContains(changedChartDirs, dir) && t.chartUtils.IsChartDir(dir) {
			changedChartDirs = append(changedChartDirs, dir)
		}
	}

	return changedChartDirs, nil
}

// ReadAllChartDirectories returns a slice of all charts in the configured chart directories except those
// configured to be excluded.
func (t *Testing) ReadAllChartDirectories() ([]string, error) {
	cfg := t.config

	var chartDirs []string
	for _, chartParentDir := range cfg.ChartDirs {
		dirs, err := t.directoryLister.ListChildDirs(chartParentDir,
			func(dir string) bool {
				return t.chartUtils.IsChartDir(dir) && !util.StringSliceContains(cfg.ExcludedCharts, path.Base(dir))
			})
		if err != nil {
			return nil, errors.Wrap(err, "Error reading chart directories")
		}
		chartDirs = append(chartDirs, dirs...)
	}

	return chartDirs, nil
}

// CheckVersionIncrement checks that the new chart version is greater than the old one using semantic version comparison.
func (t *Testing) CheckVersionIncrement(chart string) error {
	fmt.Printf("Checking chart '%s' for a version bump...\n", chart)

	oldVersion, err := t.GetOldChartVersion(chart)
	if err != nil {
		return err
	}
	if oldVersion == "" {
		// new chart, skip version check
		return nil
	}

	fmt.Println("Old chart version:", oldVersion)

	newVersion, err := t.GetNewChartVersion(chart)
	if err != nil {
		return err
	}
	fmt.Println("New chart version:", newVersion)

	result, err := util.CompareVersions(oldVersion, newVersion)
	if err != nil {
		return err
	}

	if result >= 0 {
		return errors.New("Chart version not ok. Needs a version bump!")
	}

	fmt.Println("Chart version ok.")
	return nil
}

// GetOldChartVersion gets the version of the old Chart.yaml file from the target branch.
func (t *Testing) GetOldChartVersion(chart string) (string, error) {
	cfg := t.config

	chartYamlFile := path.Join(chart, "Chart.yaml")
	if !t.git.FileExistsOnBranch(chartYamlFile, cfg.Remote, cfg.TargetBranch) {
		fmt.Printf("Unable to find chart on %s. New chart detected.\n", cfg.TargetBranch)
		return "", nil
	}

	chartYamlContents, err := t.git.Show(chartYamlFile, cfg.Remote, cfg.TargetBranch)
	if err != nil {
		return "", errors.Wrap(err, "Error reading old Chart.yaml")
	}

	chartYaml, err := util.ReadChartYaml([]byte(chartYamlContents))
	if err != nil {
		return "", errors.Wrap(err, "Error reading old chart version")
	}

	return chartYaml.Version, nil
}

// GetNewChartVersion gets the new version from the currently checked out Chart.yaml file.
func (t *Testing) GetNewChartVersion(chart string) (string, error) {
	chartYaml, err := t.chartUtils.ReadChartYaml(chart)
	if err != nil {
		return "", errors.Wrap(err, "Error reading new chart version")
	}
	return chartYaml.Version, nil
}

// ValidateMaintainers validates maintainers in the Chart.yaml file. Maintainer names must be valid accounts
// (GitHub, Bitbucket, GitLab) names. Deprecated charts must not have maintainers.
func (t *Testing) ValidateMaintainers(chart string) error {
	fmt.Println("Validating maintainers...")

	chartYaml, err := t.chartUtils.ReadChartYaml(chart)
	if err != nil {
		return err
	}

	if chartYaml.Deprecated {
		if len(chartYaml.Maintainers) > 0 {
			return errors.New("Deprecated chart must not have maintainers")
		}
		return nil
	}

	if len(chartYaml.Maintainers) == 0 {
		return errors.New("Chart doesn't have maintainers")
	}

	repoUrl, err := t.git.GetUrlForRemote(t.config.Remote)
	if err != nil {
		return err
	}

	for _, maintainer := range chartYaml.Maintainers {
		if err := t.accountValidator.Validate(repoUrl, maintainer.Name); err != nil {
			return err
		}
	}

	return nil
}

func (t *Testing) PrintPodDetailsAndLogs(namespace string) {
	pods, err := t.kubectl.GetPods("--no-headers", "--namespace", namespace, "--output", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		fmt.Println("Error printing logs:", err)
		return
	}

	util.PrintDelimiterLine("=")

	for _, pod := range pods {
		printDetails(pod, "Description of pod", "~", func(item string) error {
			return t.kubectl.DescribePod(namespace, pod)
		}, pod)

		initContainers, err := t.kubectl.GetInitContainers(namespace, pod)
		if err != nil {
			fmt.Println("Error printing logs:", err)
			return
		}

		printDetails(pod, "Logs of init container", "-",
			func(item string) error {
				return t.kubectl.Logs(namespace, pod, item)
			}, initContainers...)

		containers, err := t.kubectl.GetContainers(namespace, pod)
		if err != nil {
			fmt.Println("Error printing logs:", err)
			return
		}

		printDetails(pod, "Logs of container", "-",
			func(item string) error {
				return t.kubectl.Logs(namespace, pod, item)
			},
			containers...)
	}

	util.PrintDelimiterLine("=")
}

func printDetails(pod string, text string, delimiterChar string, printFunc func(item string) error, items ...string) {
	for _, item := range items {
		item = strings.Trim(item, "'")

		util.PrintDelimiterLine(delimiterChar)
		fmt.Printf("==> %s %s\n", text, pod)
		util.PrintDelimiterLine(delimiterChar)

		if err := printFunc(item); err != nil {
			fmt.Println("Error printing details:", err)
			return
		}

		util.PrintDelimiterLine(delimiterChar)
		fmt.Printf("<== %s %s\n", text, pod)
		util.PrintDelimiterLine(delimiterChar)
	}
}
