package sys_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi/style"
	"github.com/charmbracelet/x/exp/term/ansi/sys"
)

func TestSetForegroundColorNil(t *testing.T) {
	s := sys.SetForegroundColor(nil)
	if s != "\x1b]10;\x07" {
		t.Errorf("Unexpected string for SetForegroundColor: got %q", s)
	}
}

func TestStringImplementations(t *testing.T) {
	foregroundColor := sys.SetForegroundColor(style.BrightMagenta)
	backgroundColor := sys.SetBackgroundColor(style.ExtendedColor(255))
	cursorColor := sys.SetCursorColor(style.TrueColor(0xffeeaa))

	if foregroundColor != "\x1b]10;#ff00ff\x07" {
		t.Errorf("Unexpected string for SetForegroundColor: got %q",
			foregroundColor)
	}
	if backgroundColor != "\x1b]11;#eeeeee\x07" {
		t.Errorf("Unexpected string for SetBackgroundColor: got %q",
			backgroundColor)
	}
	if cursorColor != "\x1b]12;#ffeeaa\x07" {
		t.Errorf("Unexpected string for SetCursorColor: got %q",
			cursorColor)
	}
}
