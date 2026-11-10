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
	"testing"

	"github.com/helm/chart-testing/v3/pkg/config"
	"github.com/helm/chart-testing/v3/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type diagnosticsHelm struct {
	Helm
	calls     *[]string
	fail      string
	err       error
	namespace string
}

func (h *diagnosticsHelm) InstallWithValues(_, _, namespace, _ string) error {
	h.namespace = namespace
	*h.calls = append(*h.calls, "install")
	if h.fail == "install" {
		return h.err
	}
	return nil
}

func (h *diagnosticsHelm) UpgradeWithValues(_, _, _, _ string) error {
	*h.calls = append(*h.calls, "upgrade")
	if h.fail == "upgrade" {
		return h.err
	}
	return nil
}

func (h *diagnosticsHelm) Test(_, _ string) error {
	*h.calls = append(*h.calls, "test")
	if h.fail == "test" {
		return h.err
	}
	return nil
}

func (h *diagnosticsHelm) DeleteRelease(_, _ string) {
	*h.calls = append(*h.calls, "delete release")
}

type diagnosticsKubectl struct {
	Kubectl
	calls    *[]string
	podArgs  []string
	eventsNS string
}

func (k *diagnosticsKubectl) CreateNamespace(_ string) error       { return nil }
func (k *diagnosticsKubectl) WaitForDeployments(_, _ string) error { return nil }
func (k *diagnosticsKubectl) DeleteNamespace(_ string) {
	*k.calls = append(*k.calls, "delete namespace")
}

func (k *diagnosticsKubectl) GetEvents(namespace string) error {
	k.eventsNS = namespace
	*k.calls = append(*k.calls, "events")
	return nil
}

func (k *diagnosticsKubectl) GetPods(args ...string) ([]string, error) {
	k.podArgs = args
	return []string{"pod"}, nil
}

func (k *diagnosticsKubectl) DescribePod(_, _ string) error {
	*k.calls = append(*k.calls, "describe")
	return nil
}

func (k *diagnosticsKubectl) GetInitContainers(_, _ string) ([]string, error) {
	return []string{"init"}, nil
}

func (k *diagnosticsKubectl) GetContainers(_, _ string) ([]string, error) {
	return []string{"main"}, nil
}

func (k *diagnosticsKubectl) Logs(_, _, container string) error {
	*k.calls = append(*k.calls, "logs "+container)
	return nil
}

func TestDiagnosticsWithCleanupOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		upgrade bool
		fail    string
		calls   []string
	}{
		{name: "install", calls: []string{"install", "test"}},
		{name: "failed install", fail: "install", calls: []string{"install"}},
		{name: "failed test", fail: "test", calls: []string{"install", "test"}},
		{name: "upgrade", upgrade: true, calls: []string{"install", "test", "upgrade", "test"}},
		{name: "failed upgrade", upgrade: true, fail: "upgrade", calls: []string{"install", "test", "upgrade"}},
		{name: "skipped old revision", upgrade: true, fail: "install", calls: []string{"install"}},
	} {
		for _, skipCleanup := range []bool{false, true} {
			for _, namespace := range []string{"", "custom"} {
				t.Run(fmt.Sprintf("%s/skip=%t/namespace=%s", tc.name, skipCleanup, namespace), func(t *testing.T) {
					var calls []string
					failure := errors.New("operation failed")
					helm := &diagnosticsHelm{calls: &calls, fail: tc.fail, err: failure}
					kubectl := &diagnosticsKubectl{calls: &calls}
					ct := newTestingMock(config.Configuration{
						Namespace: namespace, ReleaseName: "release", ReleaseLabel: "app.kubernetes.io/instance",
						SkipCleanUp: skipCleanup, PrintLogs: true,
					})
					ct.helm, ct.kubectl = helm, kubectl
					chart := &Chart{path: "chart", yaml: &util.ChartYaml{Name: "chart", Version: "1.0.0"}}
					var err error
					if tc.upgrade {
						err = ct.doUpgrade(chart, chart, false)
					} else {
						err = ct.doInstall(chart)
					}
					if tc.fail != "" && (!tc.upgrade || tc.fail != "install") {
						require.ErrorIs(t, err, failure)
					} else {
						require.NoError(t, err)
					}
					want := append([]string{}, tc.calls...)
					want = append(want, "events", "describe", "logs init", "logs main")
					if !skipCleanup {
						want = append(want, "delete release")
						if namespace == "" {
							want = append(want, "delete namespace")
						}
					}
					assert.Equal(t, want, calls, "diagnostics must run before optional teardown")
					assert.Equal(t, helm.namespace, kubectl.eventsNS)
					selector := ""
					if namespace != "" {
						selector = "app.kubernetes.io/instance=release"
					}
					assert.Equal(t, []string{"--no-headers", "--namespace", helm.namespace, "--selector", selector, "--output", "jsonpath={.items[*].metadata.name}"}, kubectl.podArgs)
				})
			}
		}
	}
}
