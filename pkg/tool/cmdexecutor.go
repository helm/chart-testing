package tool

import (
	"strings"
	"text/template"

	"github.com/helm/chart-testing/v3/pkg/exec"
	"github.com/mattn/go-shellwords"
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
	if err := template.Execute(&b, data); err != nil {
		return err
	}
	rendered := b.String()

	words, err := shellwords.Parse(rendered)
	name, args := words[0], words[1:]
	if err != nil {
		return err
	}
	return t.exec.RunProcess(name, args)
}
