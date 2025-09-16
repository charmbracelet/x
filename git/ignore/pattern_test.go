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

package gitignore

import (
	"testing"
)

func TestSimpleMatch_inclusion(t *testing.T) {
	p := ParsePattern("!vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	if r != Include {
		t.Errorf("Expected Include, got %v", r)
	}
}

func TestMatch_domainLonger_mismatch(t *testing.T) {
	p := ParsePattern("value", []string{"head", "middle", "tail"})
	r := p.Match([]string{"head", "middle"}, false)
	if r != NoMatch {
		t.Errorf("Expected NoMatch, got %v", r)
	}
}

func TestMatch_domainSameLength_mismatch(t *testing.T) {
	p := ParsePattern("value", []string{"head", "middle", "tail"})
	r := p.Match([]string{"head", "middle", "tail"}, false)
	if r != NoMatch {
		t.Errorf("Expected NoMatch, got %v", r)
	}
}

func TestMatch_domainMismatch_mismatch(t *testing.T) {
	p := ParsePattern("value", []string{"head", "middle", "tail"})
	r := p.Match([]string{"head", "middle", "_tail_", "value"}, false)
	if r != NoMatch {
		t.Errorf("Expected NoMatch, got %v", r)
	}
}

func TestSimpleMatch_withDomain(t *testing.T) {
	p := ParsePattern("middle/", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "middle", "tail"}, false)
	if r != Exclude {
		t.Errorf("Expected Exclude, got %v", r)
	}
}

func TestSimpleMatch_onlyMatchInDomain_mismatch(t *testing.T) {
	p := ParsePattern("volcano/", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "tail"}, true)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_atStart(t *testing.T) {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"value", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_inTheMiddle(t *testing.T) {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_atEnd(t *testing.T) {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "value"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_atStart_dirWanted(t *testing.T) {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"value", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_inTheMiddle_dirWanted(t *testing.T) {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_atEnd_dirWanted(t *testing.T) {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value"}, true)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_atEnd_dirWanted_notADir_mismatch(t *testing.T) {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_mismatch(t *testing.T) {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "val", "tail"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_valueLonger_mismatch(t *testing.T) {
	p := ParsePattern("val", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_withAsterisk(t *testing.T) {
	p := ParsePattern("v*o", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_withQuestionMark(t *testing.T) {
	p := ParsePattern("vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_magicChars(t *testing.T) {
	p := ParsePattern("v[ou]l[kc]ano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestSimpleMatch_wrongPattern_mismatch(t *testing.T) {
	p := ParsePattern("v[ou]l[", nil)
	r := p.Match([]string{"value", "vol["}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_fromRootWithSlash(t *testing.T) {
	p := ParsePattern("/value/vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_withDomain(t *testing.T) {
	p := ParsePattern("middle/tail/", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "middle", "tail"}, true)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_onlyMatchInDomain_mismatch(t *testing.T) {
	p := ParsePattern("volcano/tail", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "tail"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_fromRootWithoutSlash(t *testing.T) {
	p := ParsePattern("value/vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_fromRoot_mismatch(t *testing.T) {
	p := ParsePattern("value/vulkano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_fromRoot_tooShort_mismatch(t *testing.T) {
	p := ParsePattern("value/vul?ano", nil)
	r := p.Match([]string{"value"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_fromRoot_notAtRoot_mismatch(t *testing.T) {
	p := ParsePattern("/value/volcano", nil)
	r := p.Match([]string{"value", "value", "volcano"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_atStart(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"value", "volcano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_notAtStart(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"head", "value", "volcano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_mismatch(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"head", "value", "Volcano", "tail"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_isDir(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_isDirAtEnd(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano"}, true)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_isDir_mismatch(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "Colcano"}, true)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_leadingAsterisks_isDirNoDirAtEnd_mismatch(t *testing.T) {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_tailingAsterisks(t *testing.T) {
	p := ParsePattern("/*lue/vol?ano/**", nil)
	r := p.Match([]string{"value", "volcano", "tail", "moretail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_tailingAsterisks_exactMatch(t *testing.T) {
	p := ParsePattern("/*lue/vol?ano/**", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_emptyMatch(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_oneMatch(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "middle", "volcano"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_multiMatch(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_isDir_trailing(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, true)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_isDir_trailing_mismatch(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_middleAsterisks_isDir(t *testing.T) {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano", "tail"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_wrongDoubleAsterisk_mismatch(t *testing.T) {
	p := ParsePattern("/*lue/**foo/vol?ano", nil)
	r := p.Match([]string{"value", "foo", "volcano", "tail"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_magicChars(t *testing.T) {
	p := ParsePattern("**/head/v[ou]l[kc]ano", nil)
	r := p.Match([]string{"value", "head", "volcano"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_wrongPattern_noTraversal_mismatch(t *testing.T) {
	p := ParsePattern("**/head/v[ou]l[", nil)
	r := p.Match([]string{"value", "head", "vol["}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_wrongPattern_onTraversal_mismatch(t *testing.T) {
	p := ParsePattern("/value/**/v[ou]l[", nil)
	r := p.Match([]string{"value", "head", "vol["}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_issue_923(t *testing.T) {
	p := ParsePattern("**/android/**/GeneratedPluginRegistrant.java", nil)
	r := p.Match([]string{"packages", "flutter_tools", "lib", "src", "android", "gradle.dart"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_folderVersusFile(t *testing.T) {
	p := ParsePattern("/a*/**", nil)
	r := p.Match([]string{"ab"}, false)
	expected := NoMatch
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}

func TestGlobMatch_folderVersusFileAgain(t *testing.T) {
	p := ParsePattern("/a*/**/a*", nil)
	r := p.Match([]string{"ab", "ab"}, false)
	expected := Exclude
	if r != expected {
		t.Errorf("Expected %v, got %v", expected, r)
	}
}
