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

func TestResolve(t *testing.T) {
	t.Parallel()

	on, off := true, false
	tests := []struct {
		name     string
		env      map[string]string
		override *bool
		want     bool
	}{
		{
			name:     "environment variable beats the override",
			env:      map[string]string{EnvVar: "false"},
			override: &on,
			want:     false,
		},
		{
			name:     "unrecognized environment value falls through",
			env:      map[string]string{EnvVar: "maybe"},
			override: &on,
			want:     true,
		},
		{
			name:     "empty environment value falls through",
			env:      map[string]string{EnvVar: ""},
			override: &off,
			want:     false,
		},
		{
			name:     "override forces support on",
			env:      map[string]string{},
			override: &on,
			want:     true,
		},
		{
			name:     "override forces support off",
			env:      map[string]string{},
			override: &off,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolve(envFrom(tt.env), tt.override)
			if got.Supported != tt.want {
				t.Errorf("resolve() supported = %v, want %v", got.Supported, tt.want)
			}
			if got.Reason == "" {
				t.Error("resolve() returned an empty reason")
			}
		})
	}
}

func TestSetOverride(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	t.Run("forces support on and off", func(t *testing.T) {
		t.Setenv(EnvVar, "")

		SetOverride(false)
		if Supported() {
			t.Error("Supported() = true, want the override to force support off")
		}

		SetOverride(true)
		if !Supported() {
			t.Error("Supported() = false, want the override to force support on")
		}
	})

	t.Run("clear restores automatic detection", func(t *testing.T) {
		t.Setenv(EnvVar, "false")

		SetOverride(true)
		ClearOverride()
		if Supported() {
			t.Error("Supported() = true, want the environment variable to decide after clearing")
		}
	})

	t.Run("environment variable takes precedence", func(t *testing.T) {
		t.Setenv(EnvVar, "true")

		SetOverride(false)
		if !Supported() {
			t.Error("Supported() = false, want NERDFONT to win over the override")
		}
	})

	t.Run("detect honors the override", func(t *testing.T) {
		t.Setenv(EnvVar, "")

		SetOverride(true)
		if got := Detect(); !got.Supported {
			t.Errorf("Detect() = %+v, want support forced on", got)
		}
	})
}
