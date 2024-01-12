package screen

import "strconv"

// EraseDisplay (ED) clears the screen or parts of the screen. Possible values:
//
//	 0: Clear from cursor to end of screen.
//	 1: Clear from cursor to beginning of the screen.
//	 2: Clear entire screen (and moves cursor to upper left on DOS).
//	 3: Clear entire screen and delete all lines saved in the scrollback buffer.
//
//	CSI <n> J
func EraseDisplay(n int) string {
	if n < 0 {
		n = 0
	}
	return "\x1b" + "[" + strconv.Itoa(n) + "J"
}

// EraseLine (EL) clears the current line or parts of the line. Possible values:
//
//	0: Clear from cursor to end of line.
//	1: Clear from cursor to beginning of the line.
//	2: Clear entire line.
//
// The cursor position is not affected.
//
//	CSI <n> K
func EraseLine(n int) string {
	if n < 0 {
		n = 0
	}
	return "\x1b" + "[" + strconv.Itoa(n) + "K"
}

// ScrollUp (SU) scrolls the screen up n lines. New lines are added at the
// bottom of the screen.
//
//	CSI <n> S
func ScrollUp(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "S"
}

// ScrollDown (SD) scrolls the screen down n lines. New lines are added at the
// top of the screen.
//
//	CSI <n> T
func ScrollDown(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "T"
}

// InsertLine (IL) inserts n blank lines at the current cursor position.
// Existing lines are moved down.
//
//	CSI <n> L
func InsertLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "L"
}

// DeleteLine (DL) deletes n lines at the current cursor position. Existing
// lines are moved up.
//
//	CSI <n> M
func DeleteLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "M"
}
