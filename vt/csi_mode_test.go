package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
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

func TestSoftReset(t *testing.T) {
	e := newTestTerminal(t, 20, 10)
	_, _ = e.WriteString("\x1b[?25l\x1b[3;8r\x1b[4h\x1b[?6h\x1b[?7l\x1b[?1h\x1b=\x1b(0\x1b[1;31m\x1b[5;5H\x1b7\x1b[6;7H\x1b[!p")
	for _, m := range []ansi.Mode{ansi.ModeInsertReplace, ansi.ModeOrigin, ansi.ModeCursorKeys, ansi.ModeNumericKeypad} {
		if e.isModeSet(m) {
			t.Errorf("mode %v still set", m)
		}
	}
	if !e.isModeSet(ansi.ModeAutoWrap) {
		t.Error("DECAWM not set")
	}
	if !e.isModeSet(ansi.ModeTextCursorEnable) || e.scr.Cursor().Hidden {
		t.Error("cursor still hidden")
	}
	if r := e.scr.ScrollRegion(); r != uv.Rect(0, 0, 20, 10) {
		t.Errorf("scroll region: got = %v, want the whole screen", r)
	}
	c := e.scr.Cursor()
	// CUP 6;7 under DECOM in the 3;8 region put it at (6,7).
	if c.X != 6 || c.Y != 7 {
		t.Errorf("cursor: got = (%d,%d), want = (6,7): DECSTR does not move it", c.X, c.Y)
	}
	if !c.Pen.IsZero() {
		t.Errorf("pen: got = %+v, want reset", c.Pen)
	}
	if s := e.scr.saved; s.X != 0 || s.Y != 0 || !s.Pen.IsZero() {
		t.Errorf("saved cursor: got = (%d,%d) pen %+v, want home with no pen", s.X, s.Y, s.Pen)
	}
	if e.charsets[0] != nil || e.gl != 0 {
		t.Errorf("charsets: got G0 %v GL %d, want ASCII in G0 and GL", e.charsets[0], e.gl)
	}
}
