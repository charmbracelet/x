package vt

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestAccessors(t *testing.T) {
	e := newTestTerminal(t, 5, 3)
	_, _ = e.WriteString("\x1b[2;3H\x1b7\x1b[1;1H\x1b)012345")
	if !e.PendingWrap() {
		t.Error("PendingWrap: got = false after filling the row")
	}
	if s := e.MainScreen().SavedCursor(); s.X != 2 || s.Y != 1 {
		t.Errorf("SavedCursor: got = (%d,%d), want = (2,1)", s.X, s.Y)
	}
	if g, gl, _ := e.Charsets(); g != [4]byte{'B', '0', 'B', 'B'} || gl != 0 {
		t.Errorf("Charsets: got = %q GL %d, want G1 special graphics and G0 in GL", g, gl)
	}
	if l := e.MainScreen().Line(0); l == nil || l.String() != "12345" {
		t.Errorf("Line(0): got = %q, want = %q", l.String(), "12345")
	}
	if l := e.MainScreen().Line(3); l != nil {
		t.Errorf("Line(3): got = %q, want nil", l.String())
	}
	if got := e.Mode(ansi.ModeAutoWrap); got != ansi.ModeSet {
		t.Errorf("Mode(DECAWM): got = %v, want = %v", got, ansi.ModeSet)
	}
	if got := e.Mode(ansi.DECMode(9999)); got != ansi.ModeNotRecognized {
		t.Errorf("Mode(?9999): got = %v, want = %v", got, ansi.ModeNotRecognized)
	}
	modes := e.Modes()
	modes[ansi.ModeAutoWrap] = ansi.ModeReset
	if e.Mode(ansi.ModeAutoWrap) != ansi.ModeSet {
		t.Error("Modes: changing the copy changed the emulator")
	}
	if e.ActiveScreen() != e.MainScreen() {
		t.Error("ActiveScreen is not the main screen")
	}
	_, _ = e.WriteString("\x1b[?1049h")
	if e.ActiveScreen() != e.AltScreen() || e.MainScreen() == e.AltScreen() {
		t.Error("ActiveScreen is not the alternate screen after ?1049h")
	}
	_, _ = e.WriteString("\x1bc")
	if g, _, _ := e.Charsets(); g != [4]byte{'B', 'B', 'B', 'B'} {
		t.Errorf("Charsets after RIS: got = %q, want ASCII", g)
	}
}

func TestAtBoundaryAndPartial(t *testing.T) {
	e := newTestTerminal(t, 5, 3)
	_, _ = e.WriteString("ab\x1b[m")
	if !e.AtBoundary() || e.Partial() != 0 {
		t.Errorf("between sequences: got AtBoundary %v Partial %d, want true and 0", e.AtBoundary(), e.Partial())
	}
	for _, partial := range []string{"\x1b", "\x1b[3", "\x1b]2;t", "\xe2\x9c", "\x1bP1"} {
		e := newTestTerminal(t, 5, 3)
		// Split across two writes, so the count carries over.
		_, _ = e.WriteString("ab\x1b[m" + partial[:1])
		_, _ = e.WriteString(partial[1:])
		if e.AtBoundary() {
			t.Errorf("AtBoundary after %q: got = true", partial)
		}
		if got := e.Partial(); got != len(partial) {
			t.Errorf("Partial after %q: got = %d, want = %d", partial, got, len(partial))
		}
	}
}
