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

import "github.com/helm/chart-testing/v3/pkg/exec"

type Linter struct {
	exec exec.ProcessExecutor
}

func NewLinter(exec exec.ProcessExecutor) Linter {
	return Linter{
		exec: exec,
	}
}

func (l Linter) ExecutablesExist() error {
	if err := l.exec.ExecutableExists("yamllint"); err != nil {
		return err
	}

	if err := l.exec.ExecutableExists("yamale"); err != nil {
		return err
	}

	return nil
}

func (l Linter) YamlLint(yamlFile string, configFile string) error {
	return l.exec.RunProcess("yamllint", "--config-file", configFile, yamlFile)
}

func (l Linter) Yamale(yamlFile string, schemaFile string) error {
	return l.exec.RunProcess("yamale", "--schema", schemaFile, yamlFile)
}
