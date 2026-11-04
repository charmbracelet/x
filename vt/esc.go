package vt

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
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
	e.scrs[0].Reset()
	e.scrs[1].Reset()
	e.resetTabStops()

	// XXX: Do we reset all modes here? Investigate.
	e.resetModes()

	e.gl, e.gr = 0, 1
	e.gsingle = 0
	e.charsets = [4]CharSet{}
	e.designations = [4]byte{}
	e.atPhantom, e.mainPhantom = false, false
	e.grapheme = e.grapheme[:0]
	e.carry.valid = false
	e.lastChar = 0
}

// softReset performs a soft terminal reset as in [ansi.DECSTR], following
// xterm: the cursor stays where it is, and the modes, margins, pen, charsets
// and saved cursor a program may have left behind go back to their defaults.
func (e *Emulator) softReset() {
	x, y := e.scr.CursorPosition()
	phantom := e.atPhantom
	defer func() {
		// Resetting DECOM homes the cursor; DECSTR leaves it.
		e.scr.setCursor(x, y, false)
		e.atPhantom = phantom
	}()
	e.setMode(ansi.ModeTextCursorEnable, ansi.ModeSet)
	e.setMode(ansi.ModeInsertReplace, ansi.ModeReset)
	e.setMode(ansi.ModeOrigin, ansi.ModeReset)
	e.setMode(ansi.ModeAutoWrap, ansi.ModeSet)
	e.setMode(ansi.ModeCursorKeys, ansi.ModeReset)
	e.setMode(ansi.ModeNumericKeypad, ansi.ModeReset)
	e.scr.setVerticalMargins(0, e.Height())
	e.scr.setHorizontalMargins(0, e.Width())
	e.scr.cur.Pen = uv.Style{}
	e.scr.saved = Cursor{Style: e.scr.cur.Style, Steady: e.scr.cur.Steady}
	e.gl, e.gr = 0, 1
	e.gsingle = 0
	e.charsets = [4]CharSet{}
	e.designations = [4]byte{}
}
