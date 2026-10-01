package nerdfont

import (
	"strings"
	"testing"
)

func TestGlyph(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	t.Setenv(EnvVar, "true")
	if got := Glyph("\ue0a0", "x"); got != "\ue0a0" {
		t.Errorf("Glyph() with support = %q, want %q", got, "\ue0a0")
	}

	t.Setenv(EnvVar, "false")
	Reset()
	if got := Glyph("\ue0a0", ""); got != "" {
		t.Errorf("Glyph() without support = %q, want the empty fallback", got)
	}
	if got := Glyph("\ue0a0", "git"); got != "git" {
		t.Errorf("Glyph() without support = %q, want %q", got, "git")
	}
}

func TestSupportedIsMemoized(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	t.Setenv(EnvVar, "true")
	if !Supported() {
		t.Fatal("Supported() = false, want true with NERDFONT=true")
	}
	if reason := Reason(); !strings.Contains(reason, "override") {
		t.Errorf("Reason() = %q, want it to mention the override", reason)
	}

	// The memoized result must survive environment changes so that render
	// paths never re-probe.
	t.Setenv(EnvVar, "false")
	if !Supported() {
		t.Error("Supported() = false after the memo was primed, want true")
	}

	Reset()
	if Supported() {
		t.Error("Supported() = true after Reset, want false with NERDFONT=false")
	}
}
