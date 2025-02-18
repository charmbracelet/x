package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// handleControl handles a control character.
func (t *Terminal) handleControl(r byte) {
	if !t.handlers.handleCc(r) {
		t.logf("unhandled sequence: ControlCode %q", r)
	}
}

// linefeed is the same as [index], except that it respects [ansi.LNM] mode.
func (t *Terminal) linefeed() {
	t.index()
	if t.isModeSet(ansi.LineFeedNewLineMode) {
		t.carriageReturn()
	}
}

// index moves the cursor down one line, scrolling up if necessary. This
// always resets the phantom state i.e. pending wrap state.
func (t *Terminal) index() {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	// TODO: Handle scrollback whenever we add it.
	if y == scroll.Max.Y-1 && x >= scroll.Min.X && x < scroll.Max.X {
		t.scr.ScrollUp(1)
	} else if y < scroll.Max.Y-1 || !cellbuf.Pos(x, y).In(scroll) {
		t.scr.moveCursor(0, 1)
	}
	t.atPhantom = false
}

// horizontalTabSet sets a horizontal tab stop at the current cursor position.
func (t *Terminal) horizontalTabSet() {
	x, _ := t.scr.CursorPosition()
	t.tabstops.Set(x)
}

// reverseIndex moves the cursor up one line, or scrolling down. This does not
// reset the phantom state i.e. pending wrap state.
func (t *Terminal) reverseIndex() {
	x, y := t.scr.CursorPosition()
	scroll := t.scr.ScrollRegion()
	if y == scroll.Min.Y && x >= scroll.Min.X && x < scroll.Max.X {
		t.scr.ScrollDown(1)
	} else {
		t.scr.moveCursor(0, -1)
	}
}
