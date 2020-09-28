package tool

import (
	"github.com/helm/chart-testing/v3/pkg/exec"
)

type Tester struct {
	exec exec.ProcessExecutor
}

func NewTester(exec exec.ProcessExecutor) Tester {
	return Tester{
		exec: exec,
	}
}

func (t Tester) RunUnitTests(chartDirectory string) error {
	return t.exec.RunProcess("helm", "unittest", "-f", "tests/*.yaml", chartDirectory)
}
