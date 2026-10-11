package vt

import "testing"

func TestRestoreCursorKeepsVisibilityAndShape(t *testing.T) {
	e := newTestTerminal(t, 5, 3)
	_, _ = e.WriteString("\x1b7\x1b[?25l\x1b[4 q\x1b8")
	c := e.scr.Cursor()
	if !c.Hidden || c.Style != CursorUnderline || !c.Steady {
		t.Errorf("cursor after DECRC: got hidden=%v style=%v steady=%v, want a hidden steady underline", c.Hidden, c.Style, c.Steady)
	}
}

func TestRestoreCursorClearsPendingWrap(t *testing.T) {
	for _, restore := range []string{"\x1b8", "\x1b[?1048l"} {
		e := newTestTerminal(t, 10, 3)
		_, _ = e.WriteString("\x1b7\x1b[9Cx" + restore + "y")
		if e.atPhantom {
			t.Errorf("%q: wrap still pending after restoring the cursor", restore)
		}
		if got := rowText(e, 0); got != "y        x" {
			t.Errorf("%q: row 0: got = %q, want = %q", restore, got, "y        x")
		}
	}
}
