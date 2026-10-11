package vt

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestInsertMode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "shifts the rest of the line", input: "abcdef\x1b[1;2H\x1b[4hXY", want: "aXYbcd"},
		{name: "reset stops inserting", input: "abcdef\x1b[1;2H\x1b[4hX\x1b[4lY", want: "aXYcde"},
		{name: "wide character shifts by its width", input: "abcdef\x1b[1;1H\x1b[4h日", want: "日abcd"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestTerminal(t, 6, 1)
			_, _ = e.WriteString(tc.input)
			if got := rowText(e, 0); got != tc.want {
				t.Errorf("row 0: got = %q, want = %q", got, tc.want)
			}
		})
	}

	t.Run("reported by DECRQM", func(t *testing.T) {
		e := newTestTerminal(t, 6, 1)
		if got := e.modes[ansi.ModeInsertReplace]; got != ansi.ModeReset {
			t.Errorf("IRM before CSI 4 h: got = %v, want = %v", got, ansi.ModeReset)
		}
		_, _ = e.WriteString("\x1b[4h")
		if got := e.modes[ansi.ModeInsertReplace]; got != ansi.ModeSet {
			t.Errorf("IRM after CSI 4 h: got = %v, want = %v", got, ansi.ModeSet)
		}
	})
}

func TestOriginModeHomesCursor(t *testing.T) {
	e := newTestTerminal(t, 10, 5)
	_, _ = e.WriteString("\x1b[2;4r\x1b[5;5H\x1b[?6h")
	if x, y := e.scr.CursorPosition(); x != 0 || y != 1 {
		t.Errorf("cursor after DECOM set: got = (%d,%d), want the region's corner (0,1)", x, y)
	}
	_, _ = e.WriteString("\x1b[3;3H\x1b[?6l")
	if x, y := e.scr.CursorPosition(); x != 0 || y != 0 {
		t.Errorf("cursor after DECOM reset: got = (%d,%d), want (0,0)", x, y)
	}
}
