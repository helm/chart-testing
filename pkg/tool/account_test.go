package tool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutGitDomain(t *testing.T) {
	var testDataSlice = []struct {
		name     string
		repoURL  string
		expected string
		err      error
	}{
		{"GitHub SSH", "git@github.com:foo/bar", "github.com", nil},
		{"GitHub HTTPS", "https://github.com/foo/bar", "github.com", nil},
		{"GitHub HTTPS with username/password", "https://foo:token@github.com/foo/bar", "github.com", nil},
		{"Gitlab SSH", "git@gitlab.com:foo/bar", "gitlab.com", nil},
		{"Gitlab HTTPS", "https://gitlab.com/foo/bar", "gitlab.com", nil},
		{"Gitlab HTTPS with username/password", "https://gitlab-ci-token:password@gitlab.com/foo/bar", "gitlab.com", nil},
		{"Bitbucket SSH", "git@bitbucket.com:foo/bar", "bitbucket.com", nil},
		{"Bitbucket HTTPS", "https://bitbucket.com/foo/bar", "bitbucket.com", nil},
		{"Bitbucket HTTPS with username/password", "https://user:pass@bitbucket.com/foo/bar", "bitbucket.com", nil},
		{"Invalid", "foo/bar", "", fmt.Errorf("could not parse git repository domain for \"foo/bar\"")},
	}

	for _, testData := range testDataSlice {
		t.Run(testData.name, func(t *testing.T) {
			actual, err := parseOutGitRepoDomain(testData.repoURL)
			assert.Equal(t, err, testData.err)
			assert.Equal(t, testData.expected, actual)
		})
	}
}
