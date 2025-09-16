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
// and has been modified to provide a simplified matcher that works with a single pattern.

package gitignore

// Matcher defines a matcher for a single gitignore pattern
type Matcher struct {
	pattern Pattern
}

// NewMatcher constructs a new simple matcher for a single pattern
func NewMatcher(pattern Pattern) *Matcher {
	return &Matcher{pattern: pattern}
}

// Match matches the given path against the single pattern
func (m *Matcher) Match(path []string, isDir bool) bool {
	match := m.pattern.Match(path, isDir)
	return match == Exclude
}
