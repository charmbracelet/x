package vt

import "testing"

// These tests cover the scroll-margin bounds invariant: an explicit DECSTBM
// bottom or DECSLRM right margin beyond the current screen size must be
// clamped (matching xterm), and the stored scroll region must never exceed
// the buffer bounds. A child TUI can legitimately emit margins computed from
// a stale size while resizes race through the PTY.

func TestDECSTBMClampsExplicitBottomToHeight(t *testing.T) {
	em := NewEmulator(80, 63)

	// Explicit bottom (64) beyond the screen height (63).
	em.Write([]byte("\x1b[1;64r"))

	if got := em.scr.scroll.Max.Y; got != 63 {
		t.Errorf("scroll.Max.Y = %d, want 63 (clamped to height)", got)
	}
}

func TestDECSTBMKeepsValidExplicitBottom(t *testing.T) {
	em := NewEmulator(80, 24)

	em.Write([]byte("\x1b[2;20r"))

	if got := em.scr.scroll.Min.Y; got != 1 {
		t.Errorf("scroll.Min.Y = %d, want 1", got)
	}
	if got := em.scr.scroll.Max.Y; got != 20 {
		t.Errorf("scroll.Max.Y = %d, want 20", got)
	}
}

func TestDECSLRMClampsExplicitRightToWidth(t *testing.T) {
	em := NewEmulator(80, 24)

	// DECSLRM requires mode 69 (DECLRMM); explicit right (100) beyond the
	// screen width (80).
	em.Write([]byte("\x1b[?69h\x1b[1;100s"))

	if got := em.scr.scroll.Max.X; got != 80 {
		t.Errorf("scroll.Max.X = %d, want 80 (clamped to width)", got)
	}
}

// TestReverseIndexAfterStaleMarginsDoesNotPanic reproduces the production
// crash path: a DECSTBM computed against a taller screen, followed by a
// reverse index at the top of the scroll region, used to leave an
// out-of-bounds scroll region and panic inside InsertLine.
func TestReverseIndexAfterStaleMarginsDoesNotPanic(t *testing.T) {
	em := NewEmulator(80, 64)
	em.Write([]byte("\x1b[1;64r")) // valid at 80x64

	em.Resize(80, 63) // browser re-fit shrinks the screen

	// Child still believes the screen is 64 rows tall.
	em.Write([]byte("\x1b[1;64r\x1b[1;1H\x1bM\x1bM\x1bM"))

	if got, want := em.scr.scroll.Max.Y, 63; got != want {
		t.Errorf("scroll.Max.Y = %d, want %d", got, want)
	}
}
