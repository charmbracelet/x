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
//
// See: https://vt100.net/docs/vt510-rm/ED.html
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
//
// See: https://vt100.net/docs/vt510-rm/EL.html
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
//
// See: https://vt100.net/docs/vt510-rm/SU.html
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
//
// See: https://vt100.net/docs/vt510-rm/SD.html
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
//
// See: https://vt100.net/docs/vt510-rm/IL.html
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
//
// See: https://vt100.net/docs/vt510-rm/DL.html
func DeleteLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "M"
}

// SetScrollingRegion (DECSTBM) sets the top and bottom margins for the scrolling
// region. The default is the entire screen.
//
//	CSI <top>;<bottom>r
//
// See: https://vt100.net/docs/vt510-rm/DECSTBM.html
func SetScrollingRegion(t, b int) string {
	var ts, bs string
	if t > 1 {
		ts = strconv.Itoa(t)
	}
	if b > 1 {
		bs = strconv.Itoa(b)
	}
	return "\x1b" + "[" + ts + ";" + bs + "r"
}
