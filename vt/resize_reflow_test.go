package vt

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func reflowRowText(e *Emulator, y int) string {
	var b strings.Builder
	for x := 0; x < e.Width(); x++ {
		if c := e.CellAt(x, y); c != nil && c.Content != "" {
			b.WriteString(c.Content)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.TrimRight(b.String(), " ")
}

// Growing taller must pull lines back down from scrollback so existing
// content stays anchored to the BOTTOM (xterm behavior), instead of
// padding blanks at the bottom and leaving the old frame at the top —
// the duplicated-text-on-stretch bug for bottom-anchored TUIs.
func TestResizeGrowPullsScrollbackBottomAnchored(t *testing.T) {
	e := NewEmulator(20, 4)
	for i := 1; i <= 8; i++ {
		e.WriteString(fmt.Sprintf("L%d\r\n", i))
	}
	beforeSB := e.ScrollbackLen()
	if beforeSB < 3 {
		t.Fatalf("need >=3 scrollback lines for the test, got %d", beforeSB)
	}
	var before []string
	for y := 0; y < 4; y++ {
		before = append(before, reflowRowText(e, y))
	}
	curBefore := e.CursorPosition().Y

	e.Resize(20, 7) // grow by 3

	// Old visible content stayed anchored to the bottom (rows 3..6).
	for y := 0; y < 4; y++ {
		if got := reflowRowText(e, y+3); got != before[y] {
			t.Errorf("bottom row %d = %q, want %q (content not bottom-anchored)", y+3, got, before[y])
		}
	}
	// Scrollback shrank by the 3 pulled lines.
	if got := e.ScrollbackLen(); got != beforeSB-3 {
		t.Errorf("scrollback len = %d, want %d", got, beforeSB-3)
	}
	// The revealed top row is real pulled history, not blank.
	if reflowRowText(e, 0) == "" {
		t.Error("top row empty after grow; expected pulled scrollback content")
	}
	// Cursor moved down with the content.
	if got := e.CursorPosition().Y; got != curBefore+3 {
		t.Errorf("cursor Y = %d, want %d (should follow content down)", got, curBefore+3)
	}
}

// Shrinking pushes the displaced TOP lines into scrollback (recoverable)
// and keeps the bottom rows + cursor.
func TestResizeShrinkPushesToScrollback(t *testing.T) {
	e := NewEmulator(20, 6)
	for i := 1; i <= 6; i++ {
		e.WriteString(fmt.Sprintf("R%d", i))
		if i < 6 {
			e.WriteString("\r\n")
		}
	}
	// Cursor is on the last written row near the bottom.
	beforeSB := e.ScrollbackLen()
	bottom := reflowRowText(e, e.Height()-1)

	e.Resize(20, 3) // shrink by 3

	if got := e.ScrollbackLen(); got <= beforeSB {
		t.Errorf("scrollback len = %d, want > %d (top lines pushed)", got, beforeSB)
	}
	// The bottom row survived the shrink.
	if got := reflowRowText(e, e.Height()-1); got != bottom {
		t.Errorf("bottom row after shrink = %q, want %q", got, bottom)
	}
}

// Pop is the inverse of Push for the newest line, across the wrapped ring.
func TestScrollbackPop(t *testing.T) {
	sb := NewScrollback(3) // small so it wraps
	mk := func(s string) uv.Line { return uv.Line{{Content: s, Width: 1}} }
	for _, s := range []string{"a", "b", "c", "d", "e"} { // wraps: keeps c,d,e
		sb.Push(mk(s))
	}
	if sb.Len() != 3 {
		t.Fatalf("len = %d, want 3", sb.Len())
	}
	got := []string{}
	for sb.Len() > 0 {
		got = append(got, sb.Pop()[0].Content)
	}
	want := []string{"e", "d", "c"} // newest-first
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("pop order = %v, want %v", got, want)
	}
	if sb.Pop() != nil {
		t.Error("Pop on empty should be nil")
	}
}
