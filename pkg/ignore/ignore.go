/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ignore

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing/fstest"

	helmignore "helm.sh/helm/v3/pkg/ignore"
)

func LoadRules(dir string) (*helmignore.Rules, error) {
	rules, err := helmignore.ParseFile(filepath.Join(dir, helmignore.HelmIgnore))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if rules == nil {
		rules = helmignore.Empty()
	}
	rules.AddDefaults()
	return rules, nil
}

func FilterFiles(files []string, rules *helmignore.Rules) ([]string, error) {
	fsys := fstest.MapFS{}
	for _, file := range files {
		fsys[file] = &fstest.MapFile{}
	}

	filteredFiles := []string{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		// Normalize to / since it will also work on Windows
		path = filepath.ToSlash(path)

		if fi.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(path, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If a .helmignore file matches, skip this file.
		if rules.Ignore(path, fi) {
			return nil
		}

		filteredFiles = append(filteredFiles, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filteredFiles, nil
}
