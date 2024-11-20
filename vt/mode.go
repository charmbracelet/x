package vt

import "github.com/charmbracelet/x/ansi"

// resetModes resets all modes to their default values.
func (t *Terminal) resetModes() {
	t.modes = map[ansi.Mode]ansi.ModeSetting{
		// Recognized modes and their default values.
		ansi.CursorKeysMode:          ansi.ModeReset,
		ansi.OriginMode:              ansi.ModeReset,
		ansi.AutoWrapMode:            ansi.ModeSet,
		ansi.X10MouseMode:            ansi.ModeReset,
		ansi.LineFeedNewLineMode:     ansi.ModeReset,
		ansi.TextCursorEnableMode:    ansi.ModeSet,
		ansi.NumericKeypadMode:       ansi.ModeReset,
		ansi.LeftRightMarginMode:     ansi.ModeReset,
		ansi.NormalMouseMode:         ansi.ModeReset,
		ansi.HighlightMouseMode:      ansi.ModeReset,
		ansi.ButtonEventMouseMode:    ansi.ModeReset,
		ansi.AnyEventMouseMode:       ansi.ModeReset,
		ansi.FocusEventMode:          ansi.ModeReset,
		ansi.SgrExtMouseMode:         ansi.ModeReset,
		ansi.AltScreenMode:           ansi.ModeReset,
		ansi.SaveCursorMode:          ansi.ModeReset,
		ansi.AltScreenSaveCursorMode: ansi.ModeReset,
		ansi.BracketedPasteMode:      ansi.ModeReset,
	}

	// Set mode effects.
	for mode, setting := range t.modes {
		t.setMode(mode, setting)
	}
}
