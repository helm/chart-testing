package tool

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type fakeProcessExecutor struct {
	mock.Mock
}

func (c *fakeProcessExecutor) RunProcess(executable string, execArgs ...interface{}) error {
	c.Called(executable, execArgs[0])
	return nil
}

func TestCmdTemplateExecutor_RunCommand(t *testing.T) {
	type args struct {
		cmdTemplate string
		data        interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		validate func(t *testing.T, executor *fakeProcessExecutor)
	}{
		{
			name: "command without arguments",
			args: args{
				cmdTemplate: "echo",
				data:        nil,
			},
			validate: func(t *testing.T, executor *fakeProcessExecutor) {
				executor.AssertCalled(t, "RunProcess", "echo", []string{})
			},
			wantErr: false,
		},
		{
			name: "command with args",
			args: args{
				cmdTemplate: "echo hello world",
			},
			validate: func(t *testing.T, executor *fakeProcessExecutor) {
				executor.AssertCalled(t, "RunProcess", "echo", []string{"hello", "world"})
			},
			wantErr: false,
		},
		{
			name: "interpolate args",
			args: args{
				cmdTemplate: "helm unittest --helm3 -f tests/*.yaml {{ .Path }}",
				data:        map[string]string{"Path": "charts/my-chart"},
			},
			validate: func(t *testing.T, executor *fakeProcessExecutor) {
				executor.AssertCalled(t, "RunProcess", "helm", []string{"unittest", "--helm3", "-f", "tests/*.yaml", "charts/my-chart"})
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processExecutor := new(fakeProcessExecutor)
			processExecutor.On("RunProcess", mock.Anything, mock.Anything).Return(nil)
			templateExecutor := CmdTemplateExecutor{
				exec: processExecutor,
			}
			if err := templateExecutor.RunCommand(tt.args.cmdTemplate, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.validate(t, processExecutor)
		})
	}
}
