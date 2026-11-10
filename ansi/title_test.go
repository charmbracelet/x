package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSetIconNameWindowTitle(t *testing.T) {
	if ansi.SetIconNameWindowTitle("hello") != "\x1b]0;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]0;hello\x07", ansi.SetIconNameWindowTitle("hello"))
	}
}

func TestSetIconName(t *testing.T) {
	if ansi.SetIconName("hello") != "\x1b]1;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]1;hello\x07", ansi.SetIconName("hello"))
	}
}

func TestSetWindowTitle(t *testing.T) {
	if ansi.SetWindowTitle("hello") != "\x1b]2;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]2;hello\x07", ansi.SetWindowTitle("hello"))
	}
}

func TestDECSWT(t *testing.T) {
	want := "\x1b]21;hello\x1b\\"
	if got := ansi.DECSWT("hello"); got != want {
		t.Errorf("expected: %q, got: %q", want, got)
	}
}

func TestDECSIN(t *testing.T) {
	want := "\x1b]2L;hello\x1b\\"
	if got := ansi.DECSIN("hello"); got != want {
		t.Errorf("expected: %q, got: %q", want, got)
	}
}
