package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (e *Emulator) handleMode(params ansi.Params, set, isAnsi bool) {
	for _, p := range params {
		param := p.Param(-1)
		if param == -1 {
			// Missing parameter, ignore
			continue
		}

		var mode ansi.Mode = ansi.DECMode(param)
		if isAnsi {
			mode = ansi.ANSIMode(param)
		}

		setting := e.modes[mode]
		if setting == ansi.ModePermanentlyReset || setting == ansi.ModePermanentlySet {
			// Permanently set modes are ignored.
			continue
		}

		setting = ansi.ModeReset
		if set {
			setting = ansi.ModeSet
		}

		e.setMode(mode, setting)
	}
}

// setAltScreenMode sets the alternate screen mode.
func (e *Emulator) setAltScreenMode(on bool) {
	if (on && e.scr == &e.scrs[1]) || (!on && e.scr == &e.scrs[0]) {
		// Already in alternate screen mode, or normal screen, do nothing.
		return
	}
	if on {
		// The pending wrap belongs to the main screen's cursor, which the
		// switch back restores: keep it for then.
		e.mainPhantom = e.atPhantom
		e.scr = &e.scrs[1]
		e.scrs[1].cur = e.scrs[0].cur
		e.scr.Clear()
		e.scr.ClearTouched()
		e.setCursor(0, 0)
	} else {
		e.scr = &e.scrs[0]
		// The main cursor comes back with its pending wrap, as xterm's
		// ?1049l restores it with the cursor; the alternate screen's does
		// not follow. A resize meanwhile may have moved the cursor off
		// the last column, where no wrap is pending.
		e.atPhantom = e.keptMainPhantom()
		e.mainPhantom = false
	}
	if e.cb.AltScreen != nil {
		e.cb.AltScreen(on)
	}
	if e.cb.CursorVisibility != nil {
		e.cb.CursorVisibility(!e.scr.cur.Hidden)
	}
}

// saveCursor saves the cursor position.
func (e *Emulator) saveCursor() {
	e.scr.SaveCursor()
}

// restoreCursor restores the cursor position. The saved slot holds no
// pending wrap, so the restored cursor has none.
func (e *Emulator) restoreCursor() {
	e.scr.RestoreCursor()
	e.atPhantom = false
}

// setMode sets the mode to the given value.
func (e *Emulator) setMode(mode ansi.Mode, setting ansi.ModeSetting) {
	e.logf("setting mode %T(%v) to %v", mode, mode, setting)
	e.modes[mode] = setting
	switch mode {
	case ansi.ModeInsertReplace:
		e.insertMode = setting.IsSet()
	case ansi.ModeOrigin:
		// Setting or resetting DECOM homes the cursor, to the scroll
		// region's corner or the screen's.
		e.setCursorPosition(0, 0)
	case ansi.ModeTextCursorEnable:
		e.scr.setCursorHidden(!setting.IsSet())
	case ansi.ModeAltScreen:
		e.setAltScreenMode(setting.IsSet())
	case ansi.ModeSaveCursor:
		if setting.IsSet() {
			e.saveCursor()
		} else {
			e.restoreCursor()
		}
	case ansi.ModeAltScreenSaveCursor: // Alternate Screen Save Cursor (1047 & 1048)
		// Save primary screen cursor position
		// Switch to alternate screen
		// Doesn't support scrollback
		if setting.IsSet() {
			e.saveCursor()
		}
		e.setAltScreenMode(setting.IsSet())
	case ansi.ModeInBandResize:
		if setting.IsSet() {
			e.reply(ansi.InBandResize(e.Height(), e.Width(), 0, 0))
		}
	}
	if setting.IsSet() {
		if e.cb.EnableMode != nil {
			e.cb.EnableMode(mode)
		}
	} else if setting.IsReset() {
		if e.cb.DisableMode != nil {
			e.cb.DisableMode(mode)
		}
	}
}

// isModeSet returns true if the mode is set.
func (e *Emulator) isModeSet(mode ansi.Mode) bool {
	m, ok := e.modes[mode]
	return ok && m.IsSet()
}
