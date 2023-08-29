package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDirNoUseHelmignore(t *testing.T) {
	expected := []string{
		"chart_without_helmignore/Chart.yaml",
		"chart_without_helmignore/README.md",
		"chart_without_helmignore/templates/toto.yaml",
	}
	files, err := LoadDir("chart_without_helmignore", false)
	assert.Equal(t, expected, files)
	assert.Nil(t, err)
	expected = []string{
		"chart_with_helmignore/.helmignore",
		"chart_with_helmignore/Chart.yaml",
		"chart_with_helmignore/README.md",
		"chart_with_helmignore/templates/toto.yaml",
	}
	files, err = LoadDir("chart_with_helmignore", false)
	assert.Equal(t, expected, files)
	assert.Nil(t, err)
}

func TestLoadDirHelmignore(t *testing.T) {
	expected := []string{
		"chart_without_helmignore/Chart.yaml",
		"chart_without_helmignore/README.md",
		"chart_without_helmignore/templates/toto.yaml",
	}
	files, err := LoadDir("chart_without_helmignore", true)
	assert.Equal(t, expected, files)
	assert.Nil(t, err)
	expected = []string{
		"chart_with_helmignore/.helmignore",
		"chart_with_helmignore/Chart.yaml",
		"chart_with_helmignore/templates/toto.yaml",
	}
	files, err = LoadDir("chart_with_helmignore", true)
	assert.Equal(t, expected, files)
	assert.Nil(t, err)
}
