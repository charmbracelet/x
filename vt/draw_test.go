package vt

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// TestDrawFullFrame verifies that Draw repaints the entire screen every
// time, not just lines written since the last draw. Hosts that rebuild the
// frame on each render pass (like Bubble Tea) lose content otherwise.
func TestDrawFullFrame(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	if _, err := e.WriteString("hello"); err != nil {
		t.Fatal(err)
	}

	render := func() string {
		scr := uv.NewScreenBuffer(20, 5)
		e.Draw(scr, scr.Bounds())
		return scr.Render()
	}

	first := render()
	if !strings.Contains(first, "hello") {
		t.Fatalf("first draw missing content:\n%q", first)
	}

	// A second draw of the same emulator must contain the same content,
	// even though nothing was written in between.
	second := render()
	if !strings.Contains(second, "hello") {
		t.Fatalf("second draw lost content:\n%q", second)
	}
}

// TestCursorHidden verifies the cursor visibility state tracks DECSCUSR /
// DECTCEM sequences so hosts can hide the cursor when the child does.
func TestCursorHidden(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	if e.CursorHidden() {
		t.Fatal("cursor should be visible by default")
	}

	if _, err := e.WriteString("\x1b[?25l"); err != nil { // hide cursor
		t.Fatal(err)
	}
	if !e.CursorHidden() {
		t.Fatal("cursor should be hidden after DECTCEM reset")
	}

	if _, err := e.WriteString("\x1b[?25h"); err != nil { // show cursor
		t.Fatal(err)
	}
	if e.CursorHidden() {
		t.Fatal("cursor should be visible after DECTCEM set")
	}
}

// TestCursorStyle verifies the cursor style and blink state track DECSCUSR
// sequences.
func TestCursorStyle(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	style, blink := e.CursorStyle()
	if style != CursorBlock || !blink {
		t.Fatalf("default cursor should be a blinking block, got style=%v blink=%v", style, blink)
	}

	// Set cursor style: steady bar.
	if _, err := e.WriteString("\x1b[6 q"); err != nil {
		t.Fatal(err)
	}
	style, blink = e.CursorStyle()
	if style != CursorBar || blink {
		t.Fatalf("expected steady bar, got style=%v blink=%v", style, blink)
	}
}
