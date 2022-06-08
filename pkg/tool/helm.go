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
	"fmt"
	"strings"

	"github.com/helm/chart-testing/v3/pkg/exec"
)

type Helm struct {
	exec         exec.ProcessExecutor
	extraArgs    []string
	extraSetArgs []string
}

func NewHelm(exec exec.ProcessExecutor, extraArgs []string, extraSetArgs []string) Helm {
	return Helm{
		exec:         exec,
		extraArgs:    extraArgs,
		extraSetArgs: extraSetArgs,
	}
}

func (h Helm) AddRepo(name string, url string, extraArgs []string) error {
	const ociPrefix string = "oci://"

	if strings.HasPrefix(url, ociPrefix) {
		registryDomain := url[len(ociPrefix):]
		return h.exec.RunProcess("helm", "registry", "login", registryDomain, extraArgs)
	}

	return h.exec.RunProcess("helm", "repo", "add", name, url, extraArgs)
}

func (h Helm) BuildDependencies(chart string) error {
	return h.BuildDependenciesWithArgs(chart, []string{})
}

func (h Helm) BuildDependenciesWithArgs(chart string, extraArgs []string) error {
	return h.exec.RunProcess("helm", "dependency", "build", chart, extraArgs)
}

func (h Helm) LintWithValues(chart string, valuesFile string) error {
	var values []string
	if valuesFile != "" {
		values = []string{"--values", valuesFile}
	}

	return h.exec.RunProcess("helm", "lint", chart, values, h.extraArgs)
}

func (h Helm) InstallWithValues(chart string, valuesFile string, namespace string, release string) error {
	var values []string
	if valuesFile != "" {
		values = []string{"--values", valuesFile}
	}

	return h.exec.RunProcess("helm", "install", release, chart, "--namespace", namespace,
		"--wait", values, h.extraArgs, h.extraSetArgs)
}

func (h Helm) Upgrade(chart string, namespace string, release string) error {
	return h.exec.RunProcess("helm", "upgrade", release, chart, "--namespace", namespace,
		"--reuse-values", "--wait", h.extraArgs, h.extraSetArgs)
}

func (h Helm) Test(namespace string, release string) error {
	return h.exec.RunProcess("helm", "test", release, "--namespace", namespace, h.extraArgs)
}

func (h Helm) DeleteRelease(namespace string, release string) {
	fmt.Printf("Deleting release %q...\n", release)
	if err := h.exec.RunProcess("helm", "uninstall", release, "--namespace", namespace, h.extraArgs); err != nil {
		fmt.Println("Error deleting Helm release:", err)
	}
}

func (h Helm) Version() (string, error) {
	return h.exec.RunProcessAndCaptureStdout("helm", "version", "--template", "{{ .Version }}")
}
