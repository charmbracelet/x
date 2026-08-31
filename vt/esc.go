package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// handleEsc handles an escape sequence.
func (e *Emulator) handleEsc(cmd ansi.Cmd) {
	e.flushGrapheme() // Flush any pending grapheme before handling ESC sequences.
	if !e.handlers.handleEsc(int(cmd)) {
		var str string
		if inter := cmd.Intermediate(); inter != 0 {
			str += string(inter) + " "
		}
		if final := cmd.Final(); final != 0 {
			str += string(final)
		}
		e.logf("unhandled sequence: ESC %q", str)
	}
}

// fullReset performs a full terminal reset as in [ansi.RIS].
func (e *Emulator) fullReset() {
	// A reset returns the cursor to a visible, unstyled, home-position state,
	// and nothing along the way reports any of that: [Screen.Reset] assigns a
	// zero Cursor rather than going through the setters that carry the
	// callbacks, and by the time the mode table is restored the values already
	// match. Those three properties are only observable through their
	// callbacks, so a consumer drawing the cursor would keep drawing a hidden
	// one, in the wrong place, in the wrong style, until some later application
	// happened to change each.
	//
	// The callbacks are muted for the reset and each change reported once at
	// the end. Several paths in here touch the cursor — leaving the alternate
	// screen reports visibility unconditionally, for one — and a reset is one
	// event to a consumer, not a burst.
	was, cb := e.scr.cur, e.cb
	e.cb.CursorVisibility, e.cb.CursorPosition, e.cb.CursorStyle = nil, nil, nil

	defer func() {
		// Restored in a defer so a panic in one of the callbacks fired during
		// the reset, recovered by the caller, cannot leave them muted for good.
		e.cb.CursorVisibility = cb.CursorVisibility
		e.cb.CursorPosition = cb.CursorPosition
		e.cb.CursorStyle = cb.CursorStyle

		cur := e.scr.cur
		if cb.CursorVisibility != nil && cur.Hidden != was.Hidden {
			cb.CursorVisibility(!cur.Hidden)
		}
		if cb.CursorPosition != nil && cur.Position != was.Position {
			cb.CursorPosition(was.Position, cur.Position)
		}
		if cb.CursorStyle != nil && (cur.Style != was.Style || cur.Steady != was.Steady) {
			cb.CursorStyle(cur.Style, !cur.Steady)
		}
	}()

	e.scrs[0].Reset()
	e.scrs[1].Reset()
	e.resetTabStops()

	// XXX: Do we reset all modes here? Investigate.
	e.resetModes()

	e.gl, e.gr = 0, 1
	e.gsingle = 0
	e.charsets = [4]CharSet{}
	e.atPhantom = false
	e.grapheme = e.grapheme[:0]
	e.lastChar = 0
	e.lastState = parser.GroundState
}
