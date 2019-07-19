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
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/helm/chart-testing/pkg/config"
	"github.com/helm/chart-testing/pkg/exec"
	"github.com/helm/chart-testing/pkg/tool"
	"github.com/helm/chart-testing/pkg/util"
	"github.com/pkg/errors"
)

const maxNameLength = 63

// Git is the Interface that wraps Git operations.
//
// FileExistsOnBranch checks whether file exists on the specified remote/branch.
//
// Show returns the contents of file on the specified remote/branch.
//
// AddWorktree checks out the contents of the repository at a commit ref into the specified path.
//
// RemoveWorktree removes the working tree at the specified path.
//
// MergeBase returns the SHA1 of the merge base of commit1 and commit2.
//
// ListChangedFilesInDirs diffs commit against HEAD and returns changed files for the specified dirs.
//
// GetUrlForRemote returns the repo URL for the specified remote.
//
// ValidateRepository checks that the current working directory is a valid git repository,
// and returns nil if valid.
type Git interface {
	FileExistsOnBranch(file string, remote string, branch string) bool
	Show(file string, remote string, branch string) (string, error)
	AddWorktree(path string, ref string) error
	RemoveWorktree(path string) error
	MergeBase(commit1 string, commit2 string) (string, error)
	ListChangedFilesInDirs(commit string, dirs ...string) ([]string, error)
	GetUrlForRemote(remote string) (string, error)
	ValidateRepository() error
}

// Helm is the interface that wraps Helm operations
//
// Init runs client-side Helm initialization
//
// AddRepo adds a chart repository to the local Helm configuration
//
// BuildDependencies builds the chart's dependencies
//
// LintWithValues runs `helm lint` for the given chart using the specified values file.
// Pass a zero value for valuesFile in order to run lint without specifying a values file.
//
// InstallWithValues runs `helm install` for the given chart using the specified values file.
// Pass a zero value for valuesFile in order to run install without specifying a values file.
//
// Upgrade runs `helm upgrade` against an existing release, and re-uses the previously computed values.
//
// Test runs `helm test` against an existing release. Set the cleanup argument to true in order
// to clean up test pods created by helm after the test command completes.
//
// DeleteRelease purges the specified Helm release.
type Helm interface {
	Init() error
	AddRepo(name string, url string, extraArgs []string) error
	BuildDependencies(chart string) error
	LintWithValues(chart string, valuesFile string) error
	InstallWithValues(chart string, valuesFile string, namespace string, release string) error
	Upgrade(chart string, release string) error
	Test(release string, cleanup bool) error
	DeleteRelease(release string)
	RenderTemplate(chart string, valuesFile string) (string, error)
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
// GetEvents prints all events for namespace
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
	WaitForDeployments(namespace string, selector string) error
	GetPodsforDeployment(namespace string, deployment string) ([]string, error)
	GetPods(args ...string) ([]string, error)
	GetEvents(namespace string) error
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

// DirectoryLister is the interface
//
// ListChildDirs lists direct child directories of parentDir given they pass the test function
type DirectoryLister interface {
	ListChildDirs(parentDir string, test func(string) bool) ([]string, error)
}

// ChartUtils is the interface that wraps chart-related methods
//
// LookupChartDir looks up the chart's root directory based on some chart file that has changed
type ChartUtils interface {
	LookupChartDir(chartDirs []string, dir string) (string, error)
}

// AccountValidator is the interface that wraps Git account validation
//
// Validate checks if account is valid on repoDomain
type AccountValidator interface {
	Validate(repoDomain string, account string) error
}

// Chart represents a Helm chart, and can be initalized with the NewChart method.
type Chart struct {
	path               string
	yaml               *util.ChartYaml
	ciValuesPaths      []string
	renderedChartCache map[string]string
}

// Yaml returns the Chart metadata
func (c *Chart) Yaml() *util.ChartYaml {
	return c.yaml
}

// Path returns the chart's directory path
func (c *Chart) Path() string {
	return c.path
}

func (c *Chart) String() string {
	return fmt.Sprintf(`%s => (version: "%s", path: "%s")`, c.yaml.Name, c.yaml.Version, c.Path())
}

// ValuesFilePathsForCI returns all file paths in the 'ci' subfolder of the chart directory matching the pattern '*-values.yaml'
func (c *Chart) ValuesFilePathsForCI() []string {
	return c.ciValuesPaths
}

// HasCIValuesFile checks whether a given CI values file is present.
func (c *Chart) HasCIValuesFile(path string) bool {
	fileName := filepath.Base(path)
	for _, file := range c.ValuesFilePathsForCI() {
		if fileName == filepath.Base(file) {
			return true
		}
	}
	return false
}

// CreateInstallParams generates a randomized release name and namespace based on the chart path
// and optional buildID. If a buildID is specified, it will be part of the generated namespace.
func (c *Chart) CreateInstallParams(buildID string) (release string, namespace string) {
	release = filepath.Base(c.Path())
	if release == "." || release == "/" {
		yaml := c.Yaml()
		release = yaml.Name
	}
	namespace = release
	if buildID != "" {
		namespace = fmt.Sprintf("%s-%s", namespace, buildID)
	}
	randomSuffix := util.RandomString(10)
	release = util.SanitizeName(fmt.Sprintf("%s-%s", release, randomSuffix), maxNameLength)
	namespace = util.SanitizeName(fmt.Sprintf("%s-%s", namespace, randomSuffix), maxNameLength)
	return
}

// NewChart parses the path to a chart directory and allocates a new Chart object. If chartPath is
// not a valid chart directory an error is returned.
func NewChart(chartPath string) (*Chart, error) {
	yaml, err := util.ReadChartYaml(chartPath)
	if err != nil {
		return nil, err
	}
	matches, _ := filepath.Glob(filepath.Join(chartPath, "ci", "*-values.yaml"))
	return &Chart{chartPath, yaml, matches, make(map[string]string)}, nil
}

type Testing struct {
	config                   config.Configuration
	processExecutor          exec.ProcessExecutor
	helm                     Helm
	kubectl                  Kubectl
	git                      Git
	linter                   Linter
	accountValidator         AccountValidator
	directoryLister          DirectoryLister
	chartUtils               ChartUtils
	previousRevisionWorktree string
	chartProcessors          []ChartProcessor
}

// TestResult holds test results for a specific chart
type TestResult struct {
	Error error // Wrap all of the errors
}

type ChartProcessor interface {
	ProcessChart(chart *Chart) error
}

// NewTesting creates a new Testing struct with the given config.
func NewTesting(config config.Configuration) Testing {
	procExec := exec.NewProcessExecutor(config.Debug)
	extraArgs := strings.Fields(config.HelmExtraArgs)
	t := Testing{
		config:           config,
		helm:             tool.NewHelm(procExec, extraArgs),
		git:              tool.NewGit(procExec),
		kubectl:          tool.NewKubectl(procExec),
		linter:           tool.NewLinter(procExec),
		processExecutor:  procExec,
		accountValidator: tool.AccountValidator{},
		directoryLister:  util.DirectoryLister{},
		chartUtils:       util.ChartUtils{},
	}

	t.refreshChartProcessors()

	return t
}

func (t *Testing) refreshChartProcessors() {
	t.chartProcessors = nil

	if t.config.CheckVersionIncrement {
		t.chartProcessors = append(t.chartProcessors, ValidateVersionIncrementProcessor{Remote: t.config.Remote, TargetBranch: t.config.TargetBranch, Git: t.git})
	}

	if t.config.ValidateChartSchema {
		t.chartProcessors = append(t.chartProcessors, ValidateSchemaProcessor{Linter: t.linter, ChartYamlSchema: t.config.ChartYamlSchema})
	}

	if t.config.ValidateYaml {
		t.chartProcessors = append(t.chartProcessors, ValidateYamlProcessor{Linter: t.linter, LintConf: t.config.LintConf})
	}

	if t.config.ValidateMaintainers {
		t.chartProcessors = append(t.chartProcessors, ValidateMaintainersProcessor{Git: t.git, AccountValidator: t.accountValidator, Remote: t.config.Remote})
	}

	t.chartProcessors = append(t.chartProcessors, LintWithValuesProcessor{Helm: t.helm})

	for _, custom := range t.config.CustomManifestProcessors {
		t.chartProcessors = append(t.chartProcessors, ExecManifestProcessor{exec: t.processExecutor, Command: strings.Split(custom, " ")})
	}
}

func (t *Testing) Process() error {
	charts, err := t.renderCharts()
	if err != nil {
		return err
	}

	results := t.processCharts(charts)

	t.PrintResults(charts, results)

	return nil
}

// computePreviousRevisionPath converts any file or directory path to the same path in the
// previous revision's working tree.
func (t *Testing) computePreviousRevisionPath(fileOrDirPath string) string {
	return filepath.Join(t.previousRevisionWorktree, fileOrDirPath)
}

func (t *Testing) renderCharts() ([]*Chart, error) {
	chartDirs, err := t.FindChartDirsToBeProcessed()
	if err != nil {
		return nil, errors.Wrap(err, "Error identifying charts to process")
	} else if len(chartDirs) == 0 {
		return nil, nil
	}

	// Read in Chart YAML files that are to be processed
	var charts []*Chart
	for _, dir := range chartDirs {
		chart, err := NewChart(dir)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
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

	repoArgs := map[string][]string{}

	for _, repo := range t.config.HelmRepoExtraArgs {
		repoSlice := strings.SplitN(repo, "=", 2)
		name := repoSlice[0]
		repoExtraArgs := strings.Fields(repoSlice[1])
		repoArgs[name] = repoExtraArgs
	}

	for _, repo := range t.config.ChartRepos {
		repoSlice := strings.SplitN(repo, "=", 2)
		name := repoSlice[0]
		url := repoSlice[1]

		repoExtraArgs := repoArgs[name]
		if err := t.helm.AddRepo(name, url, repoExtraArgs); err != nil {
			return nil, errors.Wrapf(err, "Error adding repo: %s=%s", name, url)
		}
	}

	// Checkout previous chart revisions and build their dependencies
	if t.config.Upgrade {
		mergeBase, err := t.computeMergeBase()
		if err != nil {
			return nil, errors.Wrap(err, "Error identifying merge base")
		}
		// Add worktree for the target revision
		worktreePath, err := ioutil.TempDir("./", "ct_previous_revision")
		if err != nil {
			return nil, errors.Wrap(err, "Could not create previous revision directory")
		}
		t.previousRevisionWorktree = worktreePath
		err = t.git.AddWorktree(worktreePath, mergeBase)
		if err != nil {
			return nil, errors.Wrap(err, "Could not create worktree for previous revision")
		}
		defer t.git.RemoveWorktree(worktreePath)

		for _, chart := range charts {
			if err := t.helm.BuildDependencies(t.computePreviousRevisionPath(chart.Path())); err != nil {
				// Only print error (don't exit) if building dependencies for previous revision fails.
				fmt.Println(errors.Wrapf(err, "Error building dependencies for previous revision of chart '%s'\n", chart))
			}
		}
	}

	for _, chart := range charts {
		if err := t.helm.BuildDependencies(chart.Path()); err != nil {
			return nil, errors.Wrapf(err, "Error building dependencies for chart '%s'", chart)
		}
	}

	return charts, nil
}

func (t *Testing) processCharts(charts []*Chart) []TestResult {
	results := make([]TestResult, len(charts))

	for _, processor := range t.chartProcessors {
		for idxChart, chart := range charts {
			// Don't process any other step if already failed, don't want to
			// override previous errors
			if results[idxChart].Error != nil {
				continue
			}

			err := processor.ProcessChart(chart)
			if err != nil {
				results[idxChart].Error = err
			}

		}
	}

	return results
}

// PrintResults writes test results to stdout.
func (t *Testing) PrintResults(charts []*Chart, results []TestResult) {
	util.PrintDelimiterLine("-")
	if results != nil {
		for idx, result := range results {

			if result.Error != nil {
				fmt.Printf(" %s %s > %s\n", "✖︎", charts[idx], result.Error)
			} else {
				fmt.Printf(" %s %s\n", "✔︎", charts[idx])
			}
		}
	} else {
		fmt.Println("No chart changes detected.")
	}
	util.PrintDelimiterLine("-")
}

type ExecManifestProcessor struct {
	exec exec.ProcessExecutor
	Helm
	Command []string
}

func (proc ExecManifestProcessor) ProcessChart(chart *Chart) error {
	valuesYaml := filepath.Join(chart.Path(), "values.yaml")
	valuesFiles := chart.ValuesFilePathsForCI()
	yamlFiles := append([]string{valuesYaml}, valuesFiles...)

	for _, val := range yamlFiles {
		renderedChart, ok := chart.renderedChartCache[val]
		if !ok {
			renderedChart, err := proc.Helm.RenderTemplate(chart.Path(), val)
			if err != nil {
				return err
			}
			chart.renderedChartCache[val] = renderedChart
		}

		err := proc.exec.RunProcessWithPipedInput(renderedChart, proc.Command[0], proc.Command[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

type ValidateVersionIncrementProcessor struct {
	Remote       string
	TargetBranch string
	Git
}

func (proc ValidateVersionIncrementProcessor) ProcessChart(chart *Chart) error {
	fmt.Printf("Checking chart '%s' for a version bump...\n", chart)

	oldVersion, err := GetOldChartVersion(chart, proc.Remote, proc.TargetBranch, proc.Git)
	if err != nil {
		return err
	}

	if oldVersion == "" {
		// new chart, skip version check
		return nil
	}

	fmt.Println("Old chart version:", oldVersion)

	chartYaml := chart.Yaml()
	newVersion := chartYaml.Version
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

type ValidateSchemaProcessor struct {
	Linter
	ChartYamlSchema string
}

func (proc ValidateSchemaProcessor) ProcessChart(chart *Chart) error {
	chartYaml := filepath.Join(chart.Path(), "Chart.yaml")

	err := proc.Linter.Yamale(chartYaml, proc.ChartYamlSchema)
	if err != nil {
		return err
	}

	return nil
}

type ValidateYamlProcessor struct {
	Linter
	LintConf string
}

func (proc ValidateYamlProcessor) ProcessChart(chart *Chart) error {
	chartYaml := filepath.Join(chart.Path(), "Chart.yaml")
	valuesYaml := filepath.Join(chart.Path(), "values.yaml")
	valuesFiles := chart.ValuesFilePathsForCI()

	yamlFiles := append([]string{chartYaml, valuesYaml}, valuesFiles...)
	for _, yamlFile := range yamlFiles {
		err := proc.Linter.YamlLint(yamlFile, proc.LintConf)
		if err != nil {
			return err
		}
	}

	return nil
}

type ValidateMaintainersProcessor struct {
	Git
	AccountValidator
	Remote string
}

// ValidateMaintainers validates maintainers in the Chart.yaml file. Maintainer names must be valid accounts
// (GitHub, Bitbucket, GitLab) names. Deprecated charts must not have maintainers.
func (proc ValidateMaintainersProcessor) ProcessChart(chart *Chart) error {
	return ValidateMaintainers(chart, proc.Remote, proc.Git, proc.AccountValidator)
}

func ValidateMaintainers(chart *Chart, remote string, git Git, accountValidator AccountValidator) error {
	fmt.Println("Validating maintainers...")

	chartYaml := chart.Yaml()

	if chartYaml.Deprecated {
		if len(chartYaml.Maintainers) > 0 {
			return errors.New("Deprecated chart must not have maintainers")
		}
		return nil
	}

	if len(chartYaml.Maintainers) == 0 {
		return errors.New("Chart doesn't have maintainers")
	}

	repoURL, err := git.GetUrlForRemote(remote)
	if err != nil {
		return err
	}

	for _, maintainer := range chartYaml.Maintainers {
		if err := accountValidator.Validate(repoURL, maintainer.Name); err != nil {
			return err
		}
	}

	return nil
}

type LintWithValuesProcessor struct {
	Helm
}

func (proc LintWithValuesProcessor) ProcessChart(chart *Chart) error {
	valuesFiles := chart.ValuesFilePathsForCI()

	// Lint with defaults if no values files are specified.
	if len(valuesFiles) == 0 {
		valuesFiles = append(valuesFiles, "")
	}

	for _, valuesFile := range valuesFiles {
		if valuesFile != "" {
			fmt.Printf("\nLinting chart with values file '%s'...\n\n", valuesFile)
		}
		if err := proc.Helm.LintWithValues(chart.Path(), valuesFile); err != nil {
			return err
		}
	}

	return nil
}

type InstallProcessor struct {
	testing *Testing
}

func (proc InstallProcessor) ProcessChart(chart *Chart) error {
	err := proc.testing.InstallChart(chart)
	if err != nil {
		return err
	}

	return nil
}

// InstallChart installs the specified chart into a new namespace, waits for resources to become ready, and eventually
// uninstalls it and deletes the namespace again.
func (t *Testing) InstallChart(chart *Chart) error {

	if t.config.Upgrade {
		// Test upgrade from previous version
		if err := t.UpgradeChart(chart); err != nil {
			return err
		}
		// Test upgrade of current version (related: https://github.com/helm/chart-testing/issues/19)
		if err := t.doUpgrade(chart, chart, true); err != nil {
			return err
		}
	}

	if err := t.doInstall(chart); err != nil {
		return err
	}

	return nil
}

// UpgradeChart tests in-place upgrades of the specified chart relative to its previous revisions. If the
// initial install or helm test of a previous revision of the chart fails, that release is ignored and no
// error will be returned. If the latest revision of the chart introduces a potentially breaking change
// according to the SemVer specification, upgrade testing will be skipped.
func (t *Testing) UpgradeChart(chart *Chart) error {
	breakingChangeAllowed, err := CheckBreakingChangeAllowed(chart, t.config.Remote, t.config.TargetBranch, t.git)

	if breakingChangeAllowed {
		if err != nil {
			fmt.Println(errors.Wrap(err, fmt.Sprintf("Skipping upgrade test of '%s' because", chart)))
		}
		return nil
	} else if err != nil {
		fmt.Printf("Error comparing chart versions for '%s'\n", chart)
		return err
	}

	oldChart, err := NewChart(t.computePreviousRevisionPath(chart.Path()))
	if err != nil {
		return err
	}

	if err = t.doUpgrade(oldChart, chart, false); err != nil {
		return err
	}

	return nil
}

func (t *Testing) doInstall(chart *Chart) error {
	fmt.Printf("Installing chart '%s'...\n", chart)
	valuesFiles := chart.ValuesFilePathsForCI()

	// Test with defaults if no values files are specified.
	if len(valuesFiles) == 0 {
		valuesFiles = append(valuesFiles, "")
	}

	for _, valuesFile := range valuesFiles {
		if valuesFile != "" {
			fmt.Printf("\nInstalling chart with values file '%s'...\n\n", valuesFile)
		}

		// Use anonymous function. Otherwise deferred calls would pile up
		// and be executed in reverse order after the loop.
		fun := func() error {
			namespace, release, releaseSelector, cleanup := t.generateInstallConfig(chart)
			defer cleanup()

			if err := t.helm.InstallWithValues(chart.Path(), valuesFile, namespace, release); err != nil {
				return err
			}
			return t.testRelease(release, namespace, releaseSelector, false)
		}

		if err := fun(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Testing) doUpgrade(oldChart, newChart *Chart, oldChartMustPass bool) error {
	fmt.Printf("Testing upgrades of chart '%s' relative to previous revision '%s'...\n", newChart, oldChart)
	valuesFiles := oldChart.ValuesFilePathsForCI()
	if len(valuesFiles) == 0 {
		valuesFiles = append(valuesFiles, "")
	}
	for _, valuesFile := range valuesFiles {
		if valuesFile != "" {
			if t.config.SkipMissingValues && !newChart.HasCIValuesFile(valuesFile) {
				fmt.Printf("Upgrade testing for values file '%s' skipped because a corresponding values file was not found in %s/ci", valuesFile, newChart.Path())
				continue
			}
			fmt.Printf("\nInstalling chart '%s' with values file '%s'...\n\n", oldChart, valuesFile)
		}

		// Use anonymous function. Otherwise deferred calls would pile up
		// and be executed in reverse order after the loop.
		fun := func() error {
			namespace, release, releaseSelector, cleanup := t.generateInstallConfig(oldChart)
			defer cleanup()

			// Install previous version of chart. If installation fails, ignore this release.
			if err := t.helm.InstallWithValues(oldChart.Path(), valuesFile, namespace, release); err != nil {
				if oldChartMustPass {
					return err
				}
				fmt.Println(errors.Wrap(err, fmt.Sprintf("Upgrade testing for release '%s' skipped because of previous revision installation error", release)))
				return nil
			}
			if err := t.testRelease(release, namespace, releaseSelector, true); err != nil {
				if oldChartMustPass {
					return err
				}
				fmt.Println(errors.Wrap(err, fmt.Sprintf("Upgrade testing for release '%s' skipped because of previous revision testing error", release)))
				return nil
			}

			if err := t.helm.Upgrade(oldChart.Path(), release); err != nil {
				return err
			}

			return t.testRelease(release, namespace, releaseSelector, false)
		}

		if err := fun(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Testing) testRelease(release, namespace, releaseSelector string, cleanupHelmTests bool) error {
	if err := t.kubectl.WaitForDeployments(namespace, releaseSelector); err != nil {
		return err
	}
	if err := t.helm.Test(release, cleanupHelmTests); err != nil {
		return err
	}
	return nil
}

func (t *Testing) generateInstallConfig(chart *Chart) (namespace, release, releaseSelector string, cleanup func()) {
	if t.config.Namespace != "" {
		namespace = t.config.Namespace
		release, _ = chart.CreateInstallParams(t.config.BuildId)
		releaseSelector = fmt.Sprintf("%s=%s", t.config.ReleaseLabel, release)
		cleanup = func() {
			t.PrintEventsPodDetailsAndLogs(namespace, releaseSelector)
			t.helm.DeleteRelease(release)
		}
	} else {
		release, namespace = chart.CreateInstallParams(t.config.BuildId)
		cleanup = func() {
			t.PrintEventsPodDetailsAndLogs(namespace, releaseSelector)
			t.helm.DeleteRelease(release)
			t.kubectl.DeleteNamespace(namespace)
		}
	}

	return
}

// FindChartDirsToBeProcessed identifies charts to be processed depending on the configuration
// (changed charts, all charts, or specific charts).
func (t *Testing) FindChartDirsToBeProcessed() ([]string, error) {
	cfg := t.config
	if cfg.ProcessAllCharts {
		return t.ReadAllChartDirectories()
	} else if len(cfg.Charts) > 0 {
		return t.config.Charts, nil
	}
	return t.ComputeChangedChartDirectories()
}

func (t *Testing) computeMergeBase() (string, error) {
	err := t.git.ValidateRepository()
	if err != nil {
		return "", errors.New("Must be in a git repository")
	}
	return t.git.MergeBase(fmt.Sprintf("%s/%s", t.config.Remote, t.config.TargetBranch), "HEAD")
}

// ComputeChangedChartDirectories takes the merge base of HEAD and the configured remote and target branch and computes a
// slice of changed charts from that in the configured chart directories excluding those configured to be excluded.
func (t *Testing) ComputeChangedChartDirectories() ([]string, error) {
	cfg := t.config

	mergeBase, err := t.computeMergeBase()
	if err != nil {
		return nil, err
	}

	allChangedChartFiles, err := t.git.ListChangedFilesInDirs(mergeBase, cfg.ChartDirs...)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating diff")
	}

	var changedChartDirs []string
	for _, file := range allChangedChartFiles {
		pathElements := strings.SplitN(filepath.ToSlash(file), "/", 3)
		if len(pathElements) < 2 || util.StringSliceContains(cfg.ExcludedCharts, pathElements[1]) {
			continue
		}
		dir := filepath.Dir(file)
		// Make sure directory is really a chart directory
		chartDir, err := t.chartUtils.LookupChartDir(cfg.ChartDirs, dir)
		if err == nil {
			// Only add it if not already in the list
			if !util.StringSliceContains(changedChartDirs, chartDir) {
				changedChartDirs = append(changedChartDirs, chartDir)
			}
		} else {
			fmt.Printf("Directory '%s' is no chart directory. Skipping...", chartDir)
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
				_, err := t.chartUtils.LookupChartDir(cfg.ChartDirs, dir)
				return err == nil && !util.StringSliceContains(cfg.ExcludedCharts, filepath.Base(dir))
			})
		if err != nil {
			return nil, errors.Wrap(err, "Error reading chart directories")
		}
		chartDirs = append(chartDirs, dirs...)
	}

	return chartDirs, nil
}

// CheckVersionIncrement checks that the new chart version is greater than the old one using semantic version comparison.
func CheckBreakingChangeAllowed(chart *Chart, remote, targetBranch string, git Git) (allowed bool, err error) {
	oldVersion, err := GetOldChartVersion(chart, remote, targetBranch, git)
	if err != nil {
		return false, err
	}
	if oldVersion == "" {
		// new chart, skip upgrade check
		return true, fmt.Errorf("chart has no previous revision")
	}

	newVersion := chart.Yaml().Version

	return util.BreakingChangeAllowed(oldVersion, newVersion)
}

// GetOldChartVersion gets the version of the old Chart.yaml file from the target branch.
func GetOldChartVersion(chart *Chart, remote, targetBranch string, git Git) (string, error) {
	chartPath := chart.Path()

	chartYamlFile := filepath.Join(chartPath, "Chart.yaml")
	if !git.FileExistsOnBranch(chartYamlFile, remote, targetBranch) {
		fmt.Printf("Unable to find chart on %s. New chart detected.\n", targetBranch)
		return "", nil
	}

	chartYamlContents, err := git.Show(chartYamlFile, remote, targetBranch)
	if err != nil {
		return "", errors.Wrap(err, "Error reading old Chart.yaml")
	}

	chartYaml, err := util.UnmarshalChartYaml([]byte(chartYamlContents))
	if err != nil {
		return "", errors.Wrap(err, "Error reading old chart version")
	}

	return chartYaml.Version, nil
}

func (t *Testing) PrintEventsPodDetailsAndLogs(namespace string, selector string) {
	util.PrintDelimiterLine("=")

	printDetails(namespace, "Events of namespace", ".", func(item string) error {
		return t.kubectl.GetEvents(namespace)
	}, namespace)

	pods, err := t.kubectl.GetPods(
		"--no-headers",
		"--namespace",
		namespace,
		"--selector",
		selector,
		"--output",
		"jsonpath={.items[*].metadata.name}",
	)
	if err != nil {
		fmt.Println("Error printing logs:", err)
		return
	}

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

func printDetails(resource string, text string, delimiterChar string, printFunc func(item string) error, items ...string) {
	for _, item := range items {
		item = strings.Trim(item, "'")

		util.PrintDelimiterLine(delimiterChar)
		fmt.Printf("==> %s %s\n", text, resource)
		util.PrintDelimiterLine(delimiterChar)

		if err := printFunc(item); err != nil {
			fmt.Println("Error printing details:", err)
			return
		}

		util.PrintDelimiterLine(delimiterChar)
		fmt.Printf("<== %s %s\n", text, resource)
		util.PrintDelimiterLine(delimiterChar)
	}
}
