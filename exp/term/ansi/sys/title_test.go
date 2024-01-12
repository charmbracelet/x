package sys_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi/sys"
)

func TestSetIconNameWindowTitle(t *testing.T) {
	if sys.SetIconNameWindowTitle("hello") != "\x1b]0;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]0;hello\x07", sys.SetIconNameWindowTitle("hello"))
	}
}

func TestSetIconName(t *testing.T) {
	if sys.SetIconName("hello") != "\x1b]1;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]1;hello\x07", sys.SetIconName("hello"))
	}
}

func TestSetWindowTitle(t *testing.T) {
	if sys.SetWindowTitle("hello") != "\x1b]2;hello\x07" {
		t.Errorf("expected: %q, got: %q", "\x1b]2;hello\x07", sys.SetWindowTitle("hello"))
	}
}
