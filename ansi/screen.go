package ansi

import "strconv"

// EraseDisplay (ED) clears the display or parts of the display. A screen is
// the shown part of the terminal display excluding the scrollback buffer.
// Possible values:
//
// Default is 0.
//
//	 0: Clear from cursor to end of screen.
//	 1: Clear from cursor to beginning of the screen.
//	 2: Clear entire screen (and moves cursor to upper left on DOS).
//	 3: Clear entire display which delete all lines saved in the scrollback buffer (xterm).
//
//	CSI <n> J
//
// See: https://vt100.net/docs/vt510-rm/ED.html
func EraseDisplay(n int) string {
	var s string
	if n > 0 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "J"
}

// ED is an alias for [EraseDisplay].
func ED(n int) string {
	return EraseDisplay(n)
}

// EraseDisplay constants.
// These are the possible values for the EraseDisplay function.
const (
	EraseScreenBelow   = "\x1b[J"
	EraseScreenAbove   = "\x1b[1J"
	EraseEntireScreen  = "\x1b[2J"
	EraseEntireDisplay = "\x1b[3J"
)

// EraseLine (EL) clears the current line or parts of the line. Possible values:
//
//	0: Clear from cursor to end of line.
//	1: Clear from cursor to beginning of the line.
//	2: Clear entire line.
//
// The cursor position is not affected.
//
//	CSI <n> K
//
// See: https://vt100.net/docs/vt510-rm/EL.html
func EraseLine(n int) string {
	var s string
	if n > 0 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "K"
}

// EL is an alias for [EraseLine].
func EL(n int) string {
	return EraseLine(n)
}

// EraseLine constants.
// These are the possible values for the EraseLine function.
const (
	EraseLineRight  = "\x1b[K"
	EraseLineLeft   = "\x1b[1K"
	EraseEntireLine = "\x1b[2K"
)

// ScrollUp (SU) scrolls the screen up n lines. New lines are added at the
// bottom of the screen.
//
//	CSI Pn S
//
// See: https://vt100.net/docs/vt510-rm/SU.html
func ScrollUp(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "S"
}

// PanDown is an alias for [ScrollUp].
func PanDown(n int) string {
	return ScrollUp(n)
}

// SU is an alias for [ScrollUp].
func SU(n int) string {
	return ScrollUp(n)
}

// ScrollDown (SD) scrolls the screen down n lines. New lines are added at the
// top of the screen.
//
//	CSI Pn T
//
// See: https://vt100.net/docs/vt510-rm/SD.html
func ScrollDown(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "T"
}

// PanUp is an alias for [ScrollDown].
func PanUp(n int) string {
	return ScrollDown(n)
}

// SD is an alias for [ScrollDown].
func SD(n int) string {
	return ScrollDown(n)
}

// InsertLine (IL) inserts n blank lines at the current cursor position.
// Existing lines are moved down.
//
//	CSI Pn L
//
// See: https://vt100.net/docs/vt510-rm/IL.html
func InsertLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "L"
}

// IL is an alias for [InsertLine].
func IL(n int) string {
	return InsertLine(n)
}

// DeleteLine (DL) deletes n lines at the current cursor position. Existing
// lines are moved up.
//
//	CSI Pn M
//
// See: https://vt100.net/docs/vt510-rm/DL.html
func DeleteLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "M"
}

// DL is an alias for [DeleteLine].
func DL(n int) string {
	return DeleteLine(n)
}

// SetTopBottomMargins (DECSTBM) sets the top and bottom margins for the scrolling
// region. The default is the entire screen.
//
// Default is 1,1.
//
//	CSI Pt ; Pb r
//
// See: https://vt100.net/docs/vt510-rm/DECSTBM.html
func SetTopBottomMargins(top, bot int) string {
	var t, b string
	if top > 0 {
		t = strconv.Itoa(top)
	}
	if bot > 0 {
		b = strconv.Itoa(bot)
	}
	return "\x1b[" + t + ";" + b + "r"
}

// DECSTBM is an alias for [SetTopBottomMargins].
func DECSTBM(top, bot int) string {
	return SetTopBottomMargins(top, bot)
}

// SetScrollingRegion (DECSTBM) sets the top and bottom margins for the scrolling
// region. The default is the entire screen.
//
//	CSI <top> ; <bottom> r
//
// See: https://vt100.net/docs/vt510-rm/DECSTBM.html
// Deprecated: use [SetTopBottomMargins] instead.
func SetScrollingRegion(t, b int) string {
	if t < 0 {
		t = 0
	}
	if b < 0 {
		b = 0
	}
	return "\x1b[" + strconv.Itoa(t) + ";" + strconv.Itoa(b) + "r"
}

// DeleteCharacter (DCH) deletes n characters at the current cursor position.
// As the characters are deleted, the remaining characters move to the left and
// the cursor remains at the same position.
//
// Default is 1.
//
//	CSI Pn P
//
// See: https://vt100.net/docs/vt510-rm/DCH.html
func DeleteCharacter(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "P"
}

// DCH is an alias for [DeleteCharacter].
func DCH(n int) string {
	return DeleteCharacter(n)
}

// SetTabEvery8Columns (DECST8C) sets the tab stops at every 8 columns.
//
//	CSI ? 5 W
//
// See: https://vt100.net/docs/vt510-rm/DECST8C.html
const (
	SetTabEvery8Columns = "\x1b[?5W"
	DECST8C             = SetTabEvery8Columns
)
