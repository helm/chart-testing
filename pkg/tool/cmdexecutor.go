package tool

import (
	"strings"
	"text/template"

	"github.com/mattn/go-shellwords"
)

type ProcessExecutor interface {
	RunProcess(executable string, execArgs ...interface{}) error
}

type CmdTemplateExecutor struct {
	exec ProcessExecutor
}

func NewCmdTemplateExecutor(exec ProcessExecutor) CmdTemplateExecutor {
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
	if err != nil {
		return err
	}
	name, args := words[0], words[1:]
	return t.exec.RunProcess(name, args)
}
