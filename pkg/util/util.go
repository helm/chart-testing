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

package util

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

const chars = "1234567890abcdefghijklmnopqrstuvwxyz"

type Maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type ChartYaml struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Deprecated  bool   `yaml:"deprecated"`
	Maintainers []Maintainer
}

func Flatten(items []interface{}) ([]string, error) {
	return doFlatten([]string{}, items)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func doFlatten(result []string, items interface{}) ([]string, error) {
	var err error

	switch v := items.(type) {
	case string:
		result = append(result, v)
	case []string:
		result = append(result, v...)
	case []interface{}:
		for _, item := range v {
			result, err = doFlatten(result, item)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, errors.New(fmt.Sprintf("Flatten does not support %T", v))
	}

	return result, err
}

func StringSliceContains(slice []string, s string) bool {
	for _, element := range slice {
		if s == element {
			return true
		}
	}
	return false
}

func FileExists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

// RandomString string creates a random string of numbers and lower-case ascii characters with the specified length.
func RandomString(length int) string {
	n := len(chars)
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = chars[rand.Intn(n)]
	}
	return string(bytes)
}

type DirectoryLister struct{}

// ListChildDirs lists subdirectories of parentDir matching the test function.
func (l DirectoryLister) ListChildDirs(parentDir string, test func(dir string) bool) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(parentDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, dir := range fileInfos {
		dirName := dir.Name()
		parentSlashChildDir := path.Join(parentDir, dirName)
		if test(parentSlashChildDir) {
			dirs = append(dirs, parentSlashChildDir)
		}
	}

	return dirs, nil
}

type ChartUtils struct{}

func (u ChartUtils) IsChartDir(dir string) bool {
	return FileExists(path.Join(dir, "Chart.yaml"))
}

func (u ChartUtils) ReadChartYaml(dir string) (*ChartYaml, error) {
	yamlBytes, err := ioutil.ReadFile(path.Join(dir, "Chart.yaml"))
	if err != nil {
		return nil, errors.Wrap(err, "Could not read 'Chart.yaml'")
	}
	return ReadChartYaml(yamlBytes)
}

func ReadChartYaml(yamlBytes []byte) (*ChartYaml, error) {
	chartYaml := &ChartYaml{}

	if err := yaml.Unmarshal(yamlBytes, chartYaml); err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal 'Chart.yaml'")
	}

	return chartYaml, nil
}

func CompareVersions(left string, right string) (int, error) {
	leftVersion, err := semver.NewVersion(left)
	if err != nil {
		return 0, errors.Wrap(err, "Error parsing semantic version")
	}
	rightVersion, err := semver.NewVersion(right)
	if err != nil {
		return 0, errors.Wrap(err, "Error parsing semantic version")
	}
	return leftVersion.Compare(rightVersion), nil
}

func CreateInstallParams(chart string, buildId string) (release string, namespace string) {
	release = path.Base(chart)
	namespace = release
	if buildId != "" {
		namespace += buildId
	}
	randomSuffix := RandomString(10)
	release = fmt.Sprintf("%s-%s", release, randomSuffix)
	namespace = fmt.Sprintf("%s-%s", namespace, randomSuffix)
	return
}

func PrintDelimiterLine(delimiterChar string) {
	delim := make([]string, 120)
	for i := 0; i < 120; i++ {
		delim[i] = delimiterChar
	}
	fmt.Println(strings.Join(delim, ""))
}
