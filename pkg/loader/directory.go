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

package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/helm/chart-testing/v3/pkg/loader/ignore"
	"github.com/helm/chart-testing/v3/pkg/loader/sympath"
)

// LoadDir loads from a directory.
//
// This loads charts only from directories.
func LoadDir(dir string, useHelmignore bool) ([]string, error) {
	topdir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	rules := ignore.Empty()
	if useHelmignore {
		ifile := filepath.Join(topdir, ignore.HelmIgnore)
		if _, err := os.Stat(ifile); err == nil {
			r, err := ignore.ParseFile(ifile)
			if err != nil {
				return nil, err
			}
			rules = r
		}
		rules.AddDefaults()
	}

	files := []string{}
	topdir += string(filepath.Separator)

	walk := func(name string, fi os.FileInfo, err error) error {
		n := strings.TrimPrefix(name, topdir)
		if n == "" {
			// No need to process top level. Avoid bug with helmignore .* matching
			// empty names. See issue 1779.
			return nil
		}

		// Normalize to / since it will also work on Windows
		n = filepath.ToSlash(n)

		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(n, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If a .helmignore file matches, skip this file.
		if rules.Ignore(n, fi) {
			return nil
		}

		// Irregular files include devices, sockets, and other uses of files that
		// are not regular files. In Go they have a file mode type bit set.
		// See https://golang.org/pkg/os/#FileMode for examples.
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", name)
		}

		files = append(files, n)
		return nil
	}
	if err = sympath.Walk(topdir, walk); err != nil {
		return nil, err
	}

	return files, nil
}
