// Copyright The Helm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tool

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
)

type AccountValidator struct{}

var repoDomainPattern = regexp.MustCompile("(?:https://(?:[^@:]+:[^@:]+@)?|git@)([^/:]+)")

func (v AccountValidator) Validate(repoURL string, account string) error {
	domain, err := parseOutGitRepoDomain(repoURL)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/%s", domain, account)
	response, err := http.Head(url)
	if err != nil {
		return errors.Wrap(err, "Error validating maintainers")
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("Error validating maintainer '%s': %s", account, response.Status)
	}
	return nil
}

func parseOutGitRepoDomain(repoURL string) (string, error) {
	// This works for GitHub, Bitbucket, and Gitlab
	submatch := repoDomainPattern.FindStringSubmatch(repoURL)
	if submatch == nil || len(submatch) < 2 {
		return "", fmt.Errorf("Could not parse git repository domain for '%s'", repoURL)
	}
	return submatch[1], nil
}
