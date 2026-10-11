package vt

import (
	"strings"
	"testing"
)

func TestShiftOutShiftIn(t *testing.T) {
	e := newTestTerminal(t, 10, 1)
	// G1 is DEC special graphics, where "q" is a horizontal line.
	_, _ = e.WriteString("\x1b)0\x0eq\x0fq")
	if got := rowText(e, 0); got != "─q" {
		t.Errorf("row 0: got = %q, want = %q", got, "─q")
	}
}

func TestBackTabClearsPendingWrap(t *testing.T) {
	e := newTestTerminal(t, 20, 2)
	_, _ = e.WriteString(strings.Repeat("x", 20) + "\x1b[Z")
	if e.atPhantom {
		t.Error("wrap still pending after CBT moved the cursor left")
	}
	if x, _ := e.scr.CursorPosition(); x != 16 {
		t.Errorf("cursor x: got = %d, want = 16", x)
	}
	_, _ = e.WriteString("y")
	if got := rowText(e, 0); got != strings.Repeat("x", 16)+"y"+"xxx" {
		t.Errorf("row 0: got = %q, want the y at column 16", got)
	}
}
