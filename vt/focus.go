package vt

import (
	"io"

	"github.com/charmbracelet/x/ansi"
)

// Focus sends the terminal a focus event if focus events mode is enabled.
// This is the opposite of [Blur].
func (t *Emulator) Focus() {
	t.focus(true)
}

// Blur sends the terminal a blur event if focus events mode is enabled.
// This is the opposite of [Focus].
func (t *Emulator) Blur() {
	t.focus(false)
}

func (t *Emulator) focus(focus bool) {
	if mode, ok := t.modes[ansi.FocusEventMode]; ok && mode.IsSet() {
		if focus {
			_, _ = io.WriteString(t.pw, ansi.Focus)
		} else {
			_, _ = io.WriteString(t.pw, ansi.Blur)
		}
	}
}
