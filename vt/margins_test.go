package vt

import "testing"

// TestDECSTBMClampsToScreen: margins past the last row are clamped to
// the screen (xterm behavior). Regression for a daemon-wide panic:
// an app re-sent its 49-row margins right after the screen shrank to
// 48 rows (before it saw SIGWINCH), the scroll region grew taller
// than the buffer, and the next reverse index ran
// InsertLineArea off the end of it.
func TestDECSTBMClampsToScreen(t *testing.T) {
	e := NewEmulator(80, 49)
	e.WriteString("\x1b[1;49r")
	e.Resize(80, 48)
	e.WriteString("\x1b[1;49r") // stale margins for the old height
	if r := e.scr.ScrollRegion(); r.Max.Y != 48 {
		t.Fatalf("scroll region bottom = %d, want clamped to 48", r.Max.Y)
	}
	// Reverse index at the top margin scrolls the region down; this
	// was the panic.
	e.WriteString("\x1b[1;1H\x1bM")

	// A region entirely past the screen is ignored, not applied.
	e.WriteString("\x1b[60;70r")
	if r := e.scr.ScrollRegion(); r.Min.Y != 0 || r.Max.Y != 48 {
		t.Fatalf("out-of-screen margins changed the region: %v", r)
	}
}

func TestDECSLRMClampsToScreen(t *testing.T) {
	e := NewEmulator(80, 24)
	e.WriteString("\x1b[?69h")   // enable left/right margin mode
	e.WriteString("\x1b[1;200s") // right margin past the last column
	if r := e.scr.ScrollRegion(); r.Max.X != 80 {
		t.Fatalf("scroll region right = %d, want clamped to 80", r.Max.X)
	}
	e.WriteString("\x1b[1;1H\x1bM")
}
