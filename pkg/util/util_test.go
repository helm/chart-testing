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

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlatten(t *testing.T) {
	var testDataSlice = []struct {
		input   []interface{}
		expected []string
	}{
		{[]interface{}{"foo", "bar", []string{"bla", "blubb"}}, []string{"foo", "bar", "bla", "blubb"}},
		{[]interface{}{"foo", "bar", "bla", "blubb"}, []string{"foo", "bar", "bla", "blubb"}},
		{[]interface{}{"foo", "bar", []interface{}{"bla", []string{"blubb"}}}, []string{"foo", "bar", "bla", "blubb"}},
		{[]interface{}{"foo", 42, []interface{}{"bla", []string{"blubb"}}}, nil},
	}

	for index, testData := range testDataSlice {
		t.Run(string(index), func(t *testing.T) {
			actual, err := Flatten(testData.input)
			assert.Equal(t, testData.expected, actual)
			if testData.expected != nil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	var testDataSlice = []struct {
		oldVersion string
		newVersion string
		expected   int
	}{
		{"1.2.3", "1.2.4+2", -1},
		{"1+foo", "1+bar", 0},
		{"1.4-beta", "1.3", 1},
		{"1.3-beta", "1.3", -1},
		{"1", "2", -1},
		{"3", "3", 0},
		{"3-alpha", "3-beta", -1},
	}

	for index, testData := range testDataSlice {
		t.Run(string(index), func(t *testing.T) {
			actual, _ := CompareVersions(testData.oldVersion, testData.newVersion)
			assert.Equal(t, testData.expected, actual)
		})
	}
}
