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
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/helm/chart-testing/v3/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type collectingKubectl struct {
	Kubectl
	failure        string
	described      []string
	logs           []string
	containerReads int
}

func (k *collectingKubectl) GetEvents(_ string) error {
	if k.failure == "events" {
		return errors.New("events unavailable")
	}
	return nil
}

func (k *collectingKubectl) GetPods(_ ...string) ([]string, error) {
	return []string{"first", "second"}, nil
}

func (k *collectingKubectl) DescribePod(_, pod string) error {
	k.described = append(k.described, pod)
	return nil
}

func (k *collectingKubectl) GetInitContainers(_, pod string) ([]string, error) {
	k.containerReads++
	if pod == "first" && k.failure == "init list" {
		return nil, errors.New("init containers unavailable")
	}
	return []string{"init1", "init2"}, nil
}

func (k *collectingKubectl) GetContainers(_, pod string) ([]string, error) {
	k.containerReads++
	if pod == "first" && k.failure == "container list" {
		return nil, errors.New("containers unavailable")
	}
	return []string{"main1", "main2"}, nil
}

func (k *collectingKubectl) Logs(_, pod, container string) error {
	key := pod + "/" + container
	k.logs = append(k.logs, key)
	if key == k.failure {
		return errors.New("container logs unavailable")
	}
	return nil
}

func TestDiagnosticsContinueAfterErrors(t *testing.T) {
	allLogs := []string{"first/init1", "first/init2", "first/main1", "first/main2", "second/init1", "second/init2", "second/main1", "second/main2"}
	for _, tc := range []struct {
		name      string
		failure   string
		printLogs bool
		wantLogs  []string
	}{
		{name: "success", printLogs: true, wantLogs: allLogs},
		{name: "events error", failure: "events", printLogs: true, wantLogs: allLogs},
		{name: "init list error", failure: "init list", printLogs: true, wantLogs: allLogs[2:]},
		{name: "container list error", failure: "container list", printLogs: true, wantLogs: []string{"first/init1", "first/init2", "second/init1", "second/init2", "second/main1", "second/main2"}},
		{name: "init log error", failure: "first/init1", printLogs: true, wantLogs: allLogs},
		{name: "container log error", failure: "first/main1", printLogs: true, wantLogs: allLogs},
		{name: "logs disabled", failure: "init list"},
	} {
		for _, groups := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/groups=%t", tc.name, groups), func(t *testing.T) {
				output, err := os.CreateTemp(t.TempDir(), "diagnostics")
				require.NoError(t, err)
				originalStdout := os.Stdout
				t.Cleanup(func() {
					os.Stdout = originalStdout
					_ = output.Close()
				})
				kubectl := &collectingKubectl{failure: tc.failure}
				ct := Testing{config: config.Configuration{PrintLogs: tc.printLogs, GithubGroups: groups}, kubectl: kubectl}
				os.Stdout = output
				ct.PrintEventsPodDetailsAndLogs("testing", "app=test")
				os.Stdout = originalStdout
				data, err := os.ReadFile(output.Name())
				require.NoError(t, err)
				assert.Equal(t, []string{"first", "second"}, kubectl.described)
				assert.Equal(t, tc.wantLogs, kubectl.logs)
				if !tc.printLogs {
					assert.Zero(t, kubectl.containerReads)
				} else if tc.failure != "" {
					assert.Contains(t, string(data), "unavailable")
				}
				assert.Equal(t, strings.Count(string(data), "::group::"), strings.Count(string(data), "::endgroup::"), "all GitHub output groups must be closed")
			})
		}
	}
}
