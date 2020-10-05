package tool

import (
	"strings"
	"text/template"

	"github.com/helm/chart-testing/v3/pkg/exec"
)

type CmdTemplateExecutor struct {
	exec exec.ProcessExecutor
}

func NewCmdTemplateExecutor(exec exec.ProcessExecutor) CmdTemplateExecutor {
	return CmdTemplateExecutor{
		exec: exec,
	}
}

func (t CmdTemplateExecutor) RunCommand(cmdTemplate string, data interface{}) error {
	var template = template.Must(template.New("command").Parse(cmdTemplate))
	var b strings.Builder
	err := template.Execute(&b, data)
	renderedCommand := b.String()
	return t.exec.RunProcess("sh", "-c", renderedCommand)
}
