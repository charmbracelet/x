package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleMode() {
	if t.parser.ParamsLen == 0 {
		return
	}

	cmd := ansi.Cmd(t.parser.Cmd)
	mode := ansi.Param(t.parser.Params[0]).Param(-1)
	setting := ModeReset
	if cmd.Command() == 'h' {
		setting = ModeSet
	}

	t.logf("setting mode %v to %v", mode, setting)
	if cmd.Marker() == '?' {
		mode := ansi.DECMode(mode)
		t.pmodes[mode] = setting
		switch mode {
		case ansi.CursorEnableMode:
			t.scr.cur.Hidden = setting.IsReset()
		case 1047: // Alternate Screen Buffer
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
	} else {
		mode := ansi.ANSIMode(mode)
		t.modes[mode] = setting
	}
}
