package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// nextTab moves the cursor to the next tab stop n times. This respects the
// horizontal scrolling region. This performs the same function as [ansi.CHT].
func (e *Emulator) nextTab(n int) {
	x, y := e.scr.CursorPosition()
	scroll := e.scr.ScrollRegion()
	for range n {
		ts := e.tabstops.Next(x)
		// Stop as soon as the cursor does not move: past the last tab stop the
		// position stops changing, and without this the loop keeps going for
		// however many repeats the sequence asked for.
		if ts <= x {
			break
		}
		x = ts
	}

	if x >= scroll.Max.X {
		x = min(scroll.Max.X-1, e.Width()-1)
	}

	// NOTE: We use t.scr.setCursor here because we don't want to reset the
	// phantom state.
	e.scr.setCursor(x, y, false)
}

// prevTab moves the cursor to the previous tab stop n times. This respects the
// horizontal scrolling region when origin mode is set. If the cursor would
// move past the leftmost valid column, the cursor remains at the leftmost
// valid column and the operation completes.
func (e *Emulator) prevTab(n int) {
	x, _ := e.scr.CursorPosition()
	leftmargin := 0
	scroll := e.scr.ScrollRegion()
	if e.isModeSet(ansi.DECOM) {
		leftmargin = scroll.Min.X
	}

	for range n {
		ts := e.tabstops.Prev(x)
		// As in [Emulator.nextTab]: no movement means there is nothing left to
		// count down, whatever the sequence asked for.
		if ts >= x {
			break
		}
		x = ts
	}

	if x < leftmargin {
		x = leftmargin
	}

	// NOTE: We use t.scr.setCursorX here because we don't want to reset the
	// phantom state.
	e.scr.setCursorX(x, false)
}

// moveCursor moves the cursor by the given x and y deltas. If the cursor
// is at phantom, the state will reset and the cursor is back in the screen.
func (e *Emulator) moveCursor(dx, dy int) {
	e.scr.moveCursor(dx, dy)
	e.atPhantom = false
}

// setCursor sets the cursor position. This resets the phantom state.
func (e *Emulator) setCursor(x, y int) {
	e.scr.setCursor(x, y, false)
	e.atPhantom = false
}

// setCursorPosition sets the cursor position. This respects [ansi.DECOM],
// Origin Mode. This performs the same function as [ansi.CUP].
func (e *Emulator) setCursorPosition(x, y int) {
	mode, ok := e.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	e.scr.setCursor(x, y, margins)
	e.atPhantom = false
}

// carriageReturn moves the cursor to the leftmost column. If [ansi.DECOM] is
// set, the cursor is set to the left margin. If not, and the cursor is on or
// to the right of the left margin, the cursor is set to the left margin.
// Otherwise, the cursor is set to the leftmost column of the screen.
// This performs the same function as [ansi.CR].
func (e *Emulator) carriageReturn() {
	mode, ok := e.modes[ansi.DECOM]
	margins := ok && mode.IsSet()
	x, y := e.scr.CursorPosition()
	if margins {
		e.scr.setCursor(0, y, true)
	} else if region := e.scr.ScrollRegion(); uv.Pos(x, y).In(region) {
		e.scr.setCursor(region.Min.X, y, false)
	} else {
		e.scr.setCursor(0, y, false)
	}
	e.atPhantom = false
}

// repeatPreviousCharacter repeats the previous character n times. This is
// equivalent to typing the same character n times. This performs the same as
// [ansi.REP].
func (e *Emulator) repeatPreviousCharacter(n int) {
	if e.lastChar == 0 {
		return
	}
	// A repeat count is a number from the sequence and can name far more
	// characters than the screen could ever show, so it is capped rather than
	// spent. Past a full screen the result repeats every [Emulator.Width]
	// characters — the same cells, the cursor back in the same column, one
	// more identical line of scrollback — so keeping that remainder leaves the
	// screen and the cursor exactly where the whole count would have.
	if w, area := e.Width(), e.Width()*e.Height(); w > 0 && n > area {
		n = area + n%w
	}
	for range n {
		e.handlePrint(e.lastChar)
	}
}
