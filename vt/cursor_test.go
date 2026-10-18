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

// After a shrink, every way of getting a cursor back must land on the
// screen; one left outside it drops whatever is printed next.
func TestCursorsStayOnShrunkScreen(t *testing.T) {
	tests := []struct {
		name    string
		setup   string
		restore string
	}{
		{name: "DECRC", setup: "\x1b[20;70H\x1b7", restore: "\x1b8"},
		{name: "?1048", setup: "\x1b[20;70H\x1b[?1048h", restore: "\x1b[?1048l"},
		{name: "?1049", setup: "\x1b[20;70H\x1b[?1049h", restore: "\x1b[?1049l"},
		{name: "?1047", setup: "\x1b[20;70H\x1b[?1047h", restore: "\x1b[?1047l"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestTerminal(t, 80, 24)
			_, _ = e.WriteString(tc.setup)
			e.Resize(40, 10)
			_, _ = e.WriteString(tc.restore)
			x, y := e.scr.CursorPosition()
			if x >= 40 || y >= 10 {
				t.Fatalf("cursor after restore: got = (%d,%d), want inside 40x10", x, y)
			}
			_, _ = e.WriteString("Z")
			if c := e.CellAt(x, y); c == nil || c.Content != "Z" {
				t.Errorf("cell (%d,%d) after printing: got = %+v, want Z", x, y, c)
			}
		})
	}
}

// TestInsertDeleteLineClearPendingWrap pins that IL and DL, which move the
// cursor to the left margin, leave no wrap pending, as in xterm.
func TestInsertDeleteLineClearPendingWrap(t *testing.T) {
	for _, seq := range []string{"\x1b[L", "\x1b[M"} {
		e := newTestTerminal(t, 10, 3)
		_, _ = e.WriteString("\x1b[1;10Hx" + seq + "y")
		if pos := e.CursorPosition(); pos.X != 1 || pos.Y != 0 {
			t.Errorf("%q: cursor after one more character: got = %v, want = (1,0)", seq, pos)
		}
	}
}
