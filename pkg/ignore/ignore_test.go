package ignore

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	helmignore "helm.sh/helm/v3/pkg/ignore"
)

func TestFilter(t *testing.T) {
	rules, err := helmignore.Parse(strings.NewReader("/bar/\nREADME.md\n"))
	assert.Nil(t, err)
	files := []string{"Chart.yaml", "bar/xxx", "template/svc.yaml", "baz/bar/biz.txt", "README.md"}
	actual, err := FilterFiles(files, rules)
	assert.Nil(t, err)
	expected := []string{"Chart.yaml", "baz/bar/biz.txt", "template/svc.yaml"}
	assert.ElementsMatch(t, expected, actual)
}
