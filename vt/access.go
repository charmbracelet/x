package vt

import (
	"maps"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// The accessors below expose terminal state that a caller needs to reproduce
// the screen elsewhere, for example to snapshot it and replay it into another
// terminal. They are not synchronized; see [SafeEmulator].

// MainScreen returns the main screen, whichever screen is active.
func (e *Emulator) MainScreen() *Screen {
	return &e.scrs[0]
}

// AltScreen returns the alternate screen, whichever screen is active.
func (e *Emulator) AltScreen() *Screen {
	return &e.scrs[1]
}

// ActiveScreen returns the screen output is drawn on.
func (e *Emulator) ActiveScreen() *Screen {
	return e.scr
}

// SavedCursor returns the cursor saved by DECSC on this screen.
func (s *Screen) SavedCursor() Cursor {
	return s.saved
}

// Line returns row y of the screen, or nil outside it. The line is the
// screen's own storage and changes with the next Write.
func (s *Screen) Line(y int) uv.Line {
	return s.buf.Line(y)
}

// PendingWrap reports whether the cursor sits past the last column, so the
// next printed character wraps to the next line first.
func (e *Emulator) PendingWrap() bool {
	return e.atPhantom
}

// MainPendingWrap reports whether the main screen's cursor has a wrap
// pending, whichever screen is active: while the alternate screen is, it is
// the one leaving it restores.
func (e *Emulator) MainPendingWrap() bool {
	if e.scr == &e.scrs[0] {
		return e.atPhantom
	}
	return e.keptMainPhantom()
}

// keptMainPhantom is the pending wrap kept for the main screen while the
// alternate screen is active, if its cursor is still on the last column.
func (e *Emulator) keptMainPhantom() bool {
	return e.mainPhantom && e.scrs[0].cur.X >= e.scrs[0].Width()-1
}

// AtBoundary reports whether the bytes written so far end between sequences:
// the parser holds no partial escape sequence, control string or UTF-8 rune.
func (e *Emulator) AtBoundary() bool {
	return e.parser.State() == parser.GroundState
}

// Partial returns how many of the last bytes written belong to the sequence
// the parser holds unfinished, counted across Writes; zero at a boundary.
func (e *Emulator) Partial() int {
	return e.partial
}

// Mode returns the setting of a mode, or [ansi.ModeNotRecognized] for a mode
// the program never set and the emulator has no default for.
func (e *Emulator) Mode(m ansi.Mode) ansi.ModeSetting {
	return e.modes[m]
}

// Modes returns a copy of every mode setting the emulator holds.
func (e *Emulator) Modes() ansi.Modes {
	return maps.Clone(e.modes)
}

// Charsets reports the final byte of the SCS sequence designating each of
// G0-G3 ('B' for ASCII, '0' for DEC special graphics, 'A' for UK), and which
// of them GL and GR invoke.
func (e *Emulator) Charsets() (designations [4]byte, gl, gr int) {
	for i, d := range e.designations {
		if d == 0 {
			d = 'B'
		}
		designations[i] = d
	}
	return designations, e.gl, e.gr
}
