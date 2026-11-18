package vt

import "testing"

// A scroll region must never reach past the screen. xterm clamps the bottom
// margin to the number of rows and the right margin to the number of
// columns, and ignores the sequence when the top margin ends up at or past
// the bottom one.
func TestScrollRegionMarginsAreClampedToScreen(t *testing.T) {
	e := newTestTerminal(t, 40, 27)

	// A bottom margin past the last row.
	e.Write([]byte("\x1b[1;29r"))
	if region := e.scr.ScrollRegion(); region.Max.Y > e.Height() {
		t.Fatalf("scroll region %v extends past the screen height %d", region, e.Height())
	}

	// A right margin past the last column.
	e.Write([]byte("\x1b[?69h"))
	e.Write([]byte("\x1b[1;99s"))
	if region := e.scr.ScrollRegion(); region.Max.X > e.Width() {
		t.Fatalf("scroll region %v extends past the screen width %d", region, e.Width())
	}
}

// Scrolling inside a region that reaches past the screen must not panic.
func TestScrollInsideOutOfBoundsRegionDoesNotPanic(t *testing.T) {
	e := newTestTerminal(t, 40, 27)

	e.Write([]byte("\x1b[1;29r"))
	e.Write([]byte("top\x1b[H\x1bM"))
}
