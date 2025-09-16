// Copyright 2019 The go-git authors
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
//
// This file was originally part of https://github.com/go-git/go-git/blob/main/plumbing/format/gitattributes
// and has been modified to provide tests for a simplified matcher that works with a single pattern.
//
// This file was originally part of https://github.com/go-git/go-git/blob/main/plumbing/format/gitattributes
// and has been modified to be dependency free.

package gitignore

import (
	"testing"
)

func TestMatcher_Match(t *testing.T) {
	pattern := ParsePattern("**/middle/v[uo]l?ano", nil)
	matcher := NewMatcher(pattern)

	if !matcher.Match([]string{"head", "middle", "vulkano"}, false) {
		t.Error("Expected to match 'head/middle/vulkano'")
	}

	if matcher.Match([]string{"head", "middle", "other"}, false) {
		t.Error("Expected not to match 'head/middle/other'")
	}
}

func TestMatcher_Exclude(t *testing.T) {
	pattern := ParsePattern("!volcano", nil)
	matcher := NewMatcher(pattern)

	// Include patterns return false
	if matcher.Match([]string{"volcano"}, false) {
		t.Error("Include patterns should return false")
	}
}

// Test that demonstrates how to handle exclusion patterns with Matcher
func TestMatcher_ExcludeHandling(t *testing.T) {
	// For exclusion patterns, Matcher will return false
	// because exclusion means "don't exclude" which is effectively "include"
	excludePattern := ParsePattern("!volcano", nil)
	matcher := NewMatcher(excludePattern)

	// This returns false because it's an inclusion pattern
	if matcher.Match([]string{"volcano"}, false) {
		t.Error("Include patterns should return false")
	}
}

// Test the "exclude everything except" example from git documentation
// Note: This is a simplified version that tests individual patterns
func TestMatcher_EverythingExceptExample(t *testing.T) {
	// Test /* pattern (exclude everything)
	pattern1 := ParsePattern("/*", nil)
	matcher1 := NewMatcher(pattern1)

	if !matcher1.Match([]string{"foo"}, true) { // Should match and exclude
		t.Error("Expected to match 'foo'")
	}

	if !matcher1.Match([]string{"baz"}, false) { // Should match and exclude
		t.Error("Expected to match 'baz'")
	}

	// Test !/foo pattern (but don't exclude foo)
	pattern2 := ParsePattern("!/foo", nil)
	matcher2 := NewMatcher(pattern2)

	if matcher2.Match([]string{"foo"}, true) { // Should match but include (not exclude)
		t.Error("Include patterns should return false")
	}

	// Test /foo/* pattern (exclude files in foo directory)
	pattern3 := ParsePattern("/foo/*", nil)
	matcher3 := NewMatcher(pattern3)

	if !matcher3.Match([]string{"foo", "bar"}, false) { // Should match and exclude
		t.Error("Expected to match 'foo/bar'")
	}

	// Test !/foo/bar pattern (but don't exclude foo/bar)
	pattern4 := ParsePattern("!/foo/bar", nil)
	matcher4 := NewMatcher(pattern4)

	if matcher4.Match([]string{"foo", "bar"}, false) { // Should match but include (not exclude)
		t.Error("Include patterns should return false")
	}
}
