package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleMode() {
	cmd := t.parser.Cmd()
	for _, p := range t.parser.Params() {
		param := p.Param(-1)
		if param == -1 {
			// Missing parameter, ignore
			continue
		}

		setting := ModeReset
		if cmd.Command() == 'h' {
			setting = ModeSet
		}

		t.logf("setting mode %v to %v", param, setting)
		var mode ansi.Mode = ansi.ANSIMode(param)
		if cmd.Marker() == '?' {
			mode = ansi.DECMode(param)
		}

		t.setMode(mode, setting)
	}
}

// setMode sets the mode to the given value.
func (t *Terminal) setMode(mode ansi.Mode, setting ModeSetting) {
	t.modes[mode] = setting
	switch mode {
	case ansi.CursorEnableMode:
		t.scr.cur.Hidden = setting.IsReset()
	case ansi.DECMode(1047): // Alternate Screen Buffer
		if setting == ModeSet {
			t.scr = &t.scrs[1]
		} else {
			t.scr = &t.scrs[0]
		}
	case ansi.AltScreenBufferMode:
		if setting == ModeSet {
			t.scr = &t.scrs[1]
			t.scr.Clear()
			if t.Damage != nil {
				t.Damage(ScreenDamage{t.scr.Width(), t.scr.Height()})
			}
		} else {
			t.scr = &t.scrs[0]
		}
	}
}

// isModeSet returns true if the mode is set.
func (t *Terminal) isModeSet(mode ansi.Mode) bool {
	m, ok := t.modes[mode]
	return ok && m.IsSet()
}
