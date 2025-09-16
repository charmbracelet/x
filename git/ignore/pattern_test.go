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

func TestSimpleMatch_onlyMatchInDomain_mismatch() {
	p := ParsePattern("volcano/", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "tail"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestSimpleMatch_atStart() {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"value", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_inTheMiddle() {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_atEnd() {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "value"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_atStart_dirWanted() {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"value", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_inTheMiddle_dirWanted() {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_atEnd_dirWanted() {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_atEnd_dirWanted_notADir_mismatch() {
	p := ParsePattern("value/", nil)
	r := p.Match([]string{"head", "value"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestSimpleMatch_mismatch() {
	p := ParsePattern("value", nil)
	r := p.Match([]string{"head", "val", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestSimpleMatch_valueLonger_mismatch() {
	p := ParsePattern("val", nil)
	r := p.Match([]string{"head", "value", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestSimpleMatch_withAsterisk() {
	p := ParsePattern("v*o", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_withQuestionMark() {
	p := ParsePattern("vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_magicChars() {
	p := ParsePattern("v[ou]l[kc]ano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestSimpleMatch_wrongPattern_mismatch() {
	p := ParsePattern("v[ou]l[", nil)
	r := p.Match([]string{"value", "vol["}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_fromRootWithSlash() {
	p := ParsePattern("/value/vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_withDomain() {
	p := ParsePattern("middle/tail/", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "middle", "tail"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_onlyMatchInDomain_mismatch() {
	p := ParsePattern("volcano/tail", []string{"value", "volcano"})
	r := p.Match([]string{"value", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_fromRootWithoutSlash() {
	p := ParsePattern("value/vul?ano", nil)
	r := p.Match([]string{"value", "vulkano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_fromRoot_mismatch() {
	p := ParsePattern("value/vulkano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_fromRoot_tooShort_mismatch() {
	p := ParsePattern("value/vul?ano", nil)
	r := p.Match([]string{"value"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_fromRoot_notAtRoot_mismatch() {
	p := ParsePattern("/value/volcano", nil)
	r := p.Match([]string{"value", "value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_leadingAsterisks_atStart() {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"value", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_leadingAsterisks_notAtStart() {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"head", "value", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_leadingAsterisks_mismatch() {
	p := ParsePattern("**/*lue/vol?ano", nil)
	r := p.Match([]string{"head", "value", "Volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_leadingAsterisks_isDir() {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_leadingAsterisks_isDirAtEnd() {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_leadingAsterisks_isDir_mismatch() {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "Colcano"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_leadingAsterisks_isDirNoDirAtEnd_mismatch() {
	p := ParsePattern("**/*lue/vol?ano/", nil)
	r := p.Match([]string{"head", "value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_tailingAsterisks() {
	p := ParsePattern("/*lue/vol?ano/**", nil)
	r := p.Match([]string{"value", "volcano", "tail", "moretail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_tailingAsterisks_exactMatch() {
	p := ParsePattern("/*lue/vol?ano/**", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_middleAsterisks_emptyMatch() {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_middleAsterisks_oneMatch() {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "middle", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_middleAsterisks_multiMatch() {
	p := ParsePattern("/*lue/**/vol?ano", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_middleAsterisks_isDir_trailing() {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, true)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_middleAsterisks_isDir_trailing_mismatch() {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_middleAsterisks_isDir() {
	p := ParsePattern("/*lue/**/vol?ano/", nil)
	r := p.Match([]string{"value", "middle1", "middle2", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_wrongDoubleAsterisk_mismatch() {
	p := ParsePattern("/*lue/**foo/vol?ano", nil)
	r := p.Match([]string{"value", "foo", "volcano", "tail"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_magicChars() {
	p := ParsePattern("**/head/v[ou]l[kc]ano", nil)
	r := p.Match([]string{"value", "head", "volcano"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}

func TestGlobMatch_wrongPattern_noTraversal_mismatch() {
	p := ParsePattern("**/head/v[ou]l[", nil)
	r := p.Match([]string{"value", "head", "vol["}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_wrongPattern_onTraversal_mismatch() {
	p := ParsePattern("/value/**/v[ou]l[", nil)
	r := p.Match([]string{"value", "head", "vol["}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_issue_923() {
	p := ParsePattern("**/android/**/GeneratedPluginRegistrant.java", nil)
	r := p.Match([]string{"packages", "flutter_tools", "lib", "src", "android", "gradle.dart"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_folderVersusFile() {
	p := ParsePattern("/a*/**", nil)
	r := p.Match([]string{"ab"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = NoMatch, r)
}

func TestGlobMatch_folderVersusFileAgain() {
	p := ParsePattern("/a*/**/a*", nil)
	r := p.Match([]string{"ab", "ab"}, false)
	if actual != expected { t.Errorf("Expected %v, got %v", expected, actual); return }; actual = Exclude, r)
}
