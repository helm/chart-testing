package tool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseOutGitDomain(t *testing.T) {
	var testDataSlice = []struct {
		name     string
		repoUrl  string
		expected string
	}{
		{"GitHub SSH", "git@github.com:foo/bar", "github.com"},
		{"GitHub HTTPS", "https://github.com/foo/bar", "github.com"},
		{"Gitlab SSH", "git@gitlab.com:foo/bar", "gitlab.com"},
		{"Gitlab HTTPS", "https://gitlab.com/foo/bar", "gitlab.com"},
		{"Bitbucket SSH", "git@bitbucket.com:foo/bar", "bitbucket.com"},
		{"Bitbucket HTTPS", "https://bitbucket.com/foo/bar", "bitbucket.com"},
	}

	for _, testData := range testDataSlice {
		t.Run(testData.name, func(t *testing.T) {
			actual := parseOutGitRepoDomain(testData.repoUrl)
			assert.Equal(t, testData.expected, actual)
		})
	}
}
