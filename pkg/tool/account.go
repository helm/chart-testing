// Copyright © 2018 The Helm Authors
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
	"github.com/pkg/errors"
	"net/http"
	"regexp"
)

type AccountValidator struct{}

var repoDomainPattern = regexp.MustCompile("(?:https://|git@)([^/:]+)")

func (v AccountValidator) Validate(repoUrl string, account string) error {
	domain := parseOutGitRepoDomain(repoUrl)
	url := fmt.Sprintf("https://%s/%s", domain, account)
	response, err := http.Head(url)
	if err != nil {
		return errors.Wrap(err, "Error validating maintainers")
	}
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Error validating maintainer '%s': %s", account, response.Status))
	}
	return nil
}

func parseOutGitRepoDomain(repoUrl string) string {
	// This works for GitHub, Bitbucket, and Gitlab
	submatch := repoDomainPattern.FindStringSubmatch(repoUrl)
	return submatch[1]
}
