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
	"net/url"
	"regexp"
)

type AccountValidator struct{}

var scpStylePattern = regexp.MustCompile("^(?:[^@]+@)?(?<host>[^@/:]+):.+$")

func (v AccountValidator) Validate(repoURL string, account string) error {
	domain, err := parseOutGitRepoDomain(repoURL)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/%s", domain, account)
	response, err := http.Head(url) // nolint: gosec
	if err != nil {
		return fmt.Errorf("failed validating maintainers: %w", err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("failed validating maintainer %q: %s", account, response.Status)
	}
	return nil
}

func parseOutGitRepoDomain(repoURL string) (string, error) {
	// Git remotes can be either URLs or scp style remotes
	parsedURL, err := url.Parse(repoURL)

	if err != nil || len(parsedURL.Hostname()) < 1 {
		submatch := scpStylePattern.FindStringSubmatch(repoURL)
		if len(submatch) < 2 || len(submatch[1]) < 1 {
			return "", fmt.Errorf("could not parse git repository domain for %q", repoURL)
		}
		return submatch[1], nil
	}
	return parsedURL.Hostname(), nil
}
