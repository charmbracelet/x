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

// Issue #1013: shrinking must push top rows into scrollback and keep the
// bottom rows + cursor (xterm behavior), not truncate the bottom.
func TestResizeShrinkPushesTopIntoScrollback(t *testing.T) {
	e := NewEmulator(10, 5)
	e.WriteString("1\r\n2\r\n3\r\n4\r\n5") // cursor on "5"

	if got := e.ScrollbackLen(); got != 0 {
		t.Fatalf("precondition: scrollback len = %d, want 0", got)
	}
	curBefore := e.CursorPosition()

	e.Resize(10, 3)

	if got := e.Height(); got != 3 {
		t.Fatalf("height = %d, want 3", got)
	}
	wantRows := []string{"3", "4", "5"}
	for y, want := range wantRows {
		if got := reflowRowText(e, y); got != want {
			t.Errorf("row %d = %q, want %q", y, got, want)
		}
	}
	if got := e.ScrollbackLen(); got != 2 {
		t.Fatalf("scrollback len = %d, want 2", got)
	}
	sb := e.Scrollback()
	if got := string(sb.Line(0)[0].Content); got != "1" {
		t.Errorf("scrollback[0] = %q, want %q", got, "1")
	}
	if got := string(sb.Line(1)[0].Content); got != "2" {
		t.Errorf("scrollback[1] = %q, want %q", got, "2")
	}
	cur := e.CursorPosition()
	if cur.Y != 2 {
		t.Errorf("cursor Y = %d, want 2 (bottom row)", cur.Y)
	}
	if cur.X != curBefore.X {
		t.Errorf("cursor X = %d, want %d", cur.X, curBefore.X)
	}

	// Growing again should pull the pushed lines back onto the top.
	e.Resize(10, 5)
	wantGrown := []string{"1", "2", "3", "4", "5"}
	for y, want := range wantGrown {
		if got := reflowRowText(e, y); got != want {
			t.Errorf("after grow row %d = %q, want %q", y, got, want)
		}
	}
	if got := e.ScrollbackLen(); got != 0 {
		t.Errorf("after grow scrollback len = %d, want 0", got)
	}
}

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

	for y := 0; y < 4; y++ {
		if got := reflowRowText(e, y+3); got != before[y] {
			t.Errorf("bottom row %d = %q, want %q (content not bottom-anchored)", y+3, got, before[y])
		}
	}
	if got := e.ScrollbackLen(); got != beforeSB-3 {
		t.Errorf("scrollback len = %d, want %d", got, beforeSB-3)
	}
	if reflowRowText(e, 0) == "" {
		t.Error("top row empty after grow; expected pulled scrollback content")
	}
	if got := e.CursorPosition().Y; got != curBefore+3 {
		t.Errorf("cursor Y = %d, want %d (should follow content down)", got, curBefore+3)
	}
}

func TestScrollbackPop(t *testing.T) {
	sb := NewScrollback(10)
	mk := func(s string) uv.Line { return uv.Line{{Content: s, Width: 1}} }
	for _, s := range []string{"a", "b", "c"} {
		sb.Push(mk(s))
	}
	if got := sb.Pop()[0].Content; got != "c" {
		t.Fatalf("Pop() = %q, want %q", got, "c")
	}
	if got := sb.Pop()[0].Content; got != "b" {
		t.Fatalf("Pop() = %q, want %q", got, "b")
	}
	if sb.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", sb.Len())
	}
	if got := sb.Pop()[0].Content; got != "a" {
		t.Fatalf("Pop() = %q, want %q", got, "a")
	}
	if sb.Pop() != nil {
		t.Fatal("Pop() on empty should return nil")
	}
}
