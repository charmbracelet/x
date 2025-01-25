package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleMode(params ansi.Params, set, isAnsi bool) {
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

		setting := t.modes[mode]
		if setting == ansi.ModePermanentlyReset || setting == ansi.ModePermanentlySet {
			// Permanently set modes are ignored.
			continue
		}

		setting = ansi.ModeReset
		if set {
			setting = ansi.ModeSet
		}

		t.setMode(mode, setting)
	}
}

// setAltScreenMode sets the alternate screen mode.
func (t *Terminal) setAltScreenMode(on bool) {
	if on {
		t.scr = &t.scrs[1]
		t.scrs[1].cur = t.scrs[0].cur
		t.scr.Clear()
		t.setCursor(0, 0)
	} else {
		t.scr = &t.scrs[0]
	}
	if t.Callbacks.AltScreen != nil {
		t.Callbacks.AltScreen(on)
	}
}

// saveCursor saves the cursor position.
func (t *Terminal) saveCursor() {
	t.scr.SaveCursor()
}

// restoreCursor restores the cursor position.
func (t *Terminal) restoreCursor() {
	t.scr.RestoreCursor()
}

// setMode sets the mode to the given value.
func (t *Terminal) setMode(mode ansi.Mode, setting ansi.ModeSetting) {
	t.logf("setting mode %T(%v) to %v", mode, mode, setting)
	t.modes[mode] = setting
	switch mode {
	case ansi.TextCursorEnableMode:
		t.scr.setCursorHidden(!setting.IsSet())
	case ansi.AltScreenMode:
		t.setAltScreenMode(setting.IsSet())
	case ansi.SaveCursorMode:
		if setting.IsSet() {
			t.saveCursor()
		} else {
			t.restoreCursor()
		}
	case ansi.AltScreenSaveCursorMode: // Alternate Screen Save Cursor (1047 & 1048)
		// Save primary screen cursor position
		// Switch to alternate screen
		// Doesn't support scrollback
		if setting.IsSet() {
			t.saveCursor()
		}
		t.setAltScreenMode(setting.IsSet())
	}
}

// isModeSet returns true if the mode is set.
func (t *Terminal) isModeSet(mode ansi.Mode) bool {
	m, ok := t.modes[mode]
	return ok && m.IsSet()
}
