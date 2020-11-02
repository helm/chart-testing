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
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
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
		parentSlashChildDir := filepath.Join(parentDir, dirName)
		if test(parentSlashChildDir) {
			dirs = append(dirs, parentSlashChildDir)
		}
	}

	return dirs, nil
}

type ChartUtils struct{}

func (u ChartUtils) LookupChartDir(chartDirs []string, dir string) (string, error) {
	for _, chartDir := range chartDirs {
		currentDir := dir
		for {
			chartYaml := filepath.Join(currentDir, "Chart.yaml")
			parent := filepath.Dir(filepath.Dir(chartYaml))
			chartDir = strings.TrimRight(chartDir, "/") // remove any trailing slash from the dir

			// check directory has a Chart.yaml and that it is in a
			// direct subdirectory of a configured charts directory
			if FileExists(chartYaml) && (parent == chartDir) {
				return currentDir, nil
			}

			currentDir = filepath.Dir(currentDir)
			relativeDir, _ := filepath.Rel(chartDir, currentDir)
			joined := filepath.Join(chartDir, relativeDir)
			if (joined == chartDir) || strings.HasPrefix(relativeDir, "..") {
				break
			}
		}
	}
	return "", errors.New("no chart directory")
}

// ReadChartYaml attempts to parse Chart.yaml within the specified directory
// and return a newly allocated ChartYaml object. If no Chart.yaml is present
// or there is an error unmarshaling the file contents, an error will be returned.
func ReadChartYaml(dir string) (*ChartYaml, error) {
	yamlBytes, err := ioutil.ReadFile(filepath.Join(dir, "Chart.yaml"))
	if err != nil {
		return nil, errors.Wrap(err, "Could not read 'Chart.yaml'")
	}
	return UnmarshalChartYaml(yamlBytes)
}

// UnmarshalChartYaml parses the yaml encoded data and returns a newly
// allocated ChartYaml object.
func UnmarshalChartYaml(yamlBytes []byte) (*ChartYaml, error) {
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

func BreakingChangeAllowed(left string, right string) (bool, error) {
	leftVersion, err := semver.NewVersion(left)
	if err != nil {
		return false, errors.Wrap(err, "Error parsing semantic version")
	}
	rightVersion, err := semver.NewVersion(right)
	if err != nil {
		return false, errors.Wrap(err, "Error parsing semantic version")
	}

	constraintOp := "^"
	if leftVersion.Major() == 0 {
		constraintOp = "~"
	}
	c, err := semver.NewConstraint(fmt.Sprintf("%s %s", constraintOp, leftVersion.String()))
	if err != nil {
		return false, errors.Wrap(err, "Error parsing semantic version constraint")
	}

	minor, reasons := c.Validate(rightVersion)
	if len(reasons) > 0 {
		err = multierror.Append(err, reasons...)
	}

	return !minor, err
}

// Deprecated: To be removed in v4. Use PrintDelimiterLineToWriter instead.
func PrintDelimiterLine(delimiterChar string) {
	PrintDelimiterLineToWriter(os.Stdout, delimiterChar)
}

func PrintDelimiterLineToWriter(w io.Writer, delimiterChar string) {
	delim := make([]string, 120)
	for i := 0; i < 120; i++ {
		delim[i] = delimiterChar
	}
	fmt.Fprintln(w, strings.Join(delim, ""))
}

func SanitizeName(s string, maxLength int) string {
	reg := regexp.MustCompile("^[^a-zA-Z0-9]+")

	excess := len(s) - maxLength
	result := s
	if excess > 0 {
		result = s[excess:]
	}
	return reg.ReplaceAllString(result, "")
}

func GetRandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	defer listener.Close()
	if err != nil {
		return 0, err
	}

	return listener.Addr().(*net.TCPAddr).Port, nil
}
