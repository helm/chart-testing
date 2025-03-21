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
		{"GitHub SSH 2", "ssh://github.com/foo/bar", "github.com", nil},
		{"GitHub SSH 3", "ssh://git@github.com:2222/foo/bar", "github.com", nil},
		{"GitHub SSH 4", "ssh://github.com:2222/foo/bar", "github.com", nil},
		{"GitHub HTTPS", "https://github.com/foo/bar", "github.com", nil},
		{"GitHub HTTPS with username/password", "https://foo:token@github.com/foo/bar", "github.com", nil},
		{"Gitlab SSH", "git@gitlab.com:foo/bar", "gitlab.com", nil},
		{"Gitlab HTTPS", "https://gitlab.com/foo/bar", "gitlab.com", nil},
		{"Gitlab HTTPS with username/password", "https://gitlab-ci-token:password@gitlab.com/foo/bar", "gitlab.com", nil},
		{"Bitbucket SSH", "git@bitbucket.com:foo/bar", "bitbucket.com", nil},
		{"Bitbucket HTTPS", "https://bitbucket.com/foo/bar", "bitbucket.com", nil},
		{"Bitbucket HTTPS with username/password", "https://user:pass@bitbucket.com/foo/bar", "bitbucket.com", nil},
		{"Domain name without dot", "foo/bar", "", fmt.Errorf("could not parse git repository domain for \"foo/bar\"")},
		{"Domain name without dot 2", "foo/some/path/bar.git", "", fmt.Errorf("could not parse git repository domain for \"foo/some/path/bar.git\"")},
		{"Invalid", "user@:2222/bar", "", fmt.Errorf("could not parse git repository domain for \"user@:2222/bar\"")},
	}

	for _, testData := range testDataSlice {
		t.Run(testData.name, func(t *testing.T) {
			actual, err := parseOutGitRepoDomain(testData.repoURL)
			assert.Equal(t, testData.err, err)
			assert.Equal(t, testData.expected, actual)
		})
	}
}
