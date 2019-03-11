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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	var testDataSlice = []struct {
		input    []interface{}
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

func TestTruncateLeft(t *testing.T) {
	var testDataSlice = []struct {
		input     string
		maxLength int
		expected  string
	}{
		{"way_shorter_than_max_length", 63, "way_shorter_than_max_length"},
		{"max_length", len("max_length"), "max_length"},
		{"way_longer_than_max_length", 10, "max_length"},
		{"one_shorter_than_max_length", len("one_shorter_than_max_length") + 1, "one_shorter_than_max_length"},
		{"_one_longer_than_max_length", len("_one_longer_than_max_length") - 1, "one_longer_than_max_length"},
	}

	for index, testData := range testDataSlice {
		t.Run(string(index), func(t *testing.T) {
			actual := TruncateLeft(testData.input, testData.maxLength)
			fmt.Printf("actual: %s,%d, input: %s,%d\n", actual, len(actual), testData.input, testData.maxLength)
			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestBreakingChangeAllowed(t *testing.T) {
	var testDataSlice = []struct {
		left     string
		right    string
		breaking bool
	}{
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.1.1", false},
		{"0.1.0", "0.2.0", true},
		{"0.1.0", "0.2.1", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.3", "1.2.4", false},
		{"1.2.3", "1.3.0", false},
		{"1.2.3", "2.0.0", true},
		{"1.2.3", "10.0.0", true},
		{"foo", "1.0.0", false}, // version parse error
		{"1.0.0", "bar", false}, // version parse error
	}

	for index, testData := range testDataSlice {
		t.Run(string(index), func(t *testing.T) {
			actual, _ := BreakingChangeAllowed(testData.left, testData.right)
			assert.Equal(t, testData.breaking, actual, fmt.Sprintf("input: %s,%s\n", testData.left, testData.right))
		})
	}
}
