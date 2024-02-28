package cursor

import "strconv"

// Save (DECSC) is an escape sequence that saves the current cursor position.
//
// See: https://vt100.net/docs/vt510-rm/DECSC.html
const Save = "\x1b" + "7"

// Restore (DECRC) is an escape sequence that restores the cursor position.
//
// See: https://vt100.net/docs/vt510-rm/DECRC.html
const Restore = "\x1b" + "8"

// Up (CUU) returns a sequence for moving the cursor up n cells.
//
//	CSI n A
//
// See: https://vt100.net/docs/vt510-rm/CUU.html
func Up(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "A"
}

// Down (CUD) returns a sequence for moving the cursor down n cells.
//
//	CSI n B
//
// See: https://vt100.net/docs/vt510-rm/CUD.html
func Down(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "B"
}

// Right (CUF) returns a sequence for moving the cursor right n cells.
//
//	CSI n C
//
// See: https://vt100.net/docs/vt510-rm/CUF.html
func Right(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "C"
}

// Left (CUB) returns a sequence for moving the cursor left n cells.
//
//	CSI n D
//
// See: https://vt100.net/docs/vt510-rm/CUB.html
func Left(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "D"
}

// NextLine (CNL) returns a sequence for moving the cursor to the beginning of
// the next line n times.
//
//	CSI n E
//
// See: https://vt100.net/docs/vt510-rm/CNL.html
func NextLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "E"
}

// PreviousLine (CPL) returns a sequence for moving the cursor to the beginning
// of the previous line n times.
//
//	CSI n F
//
// See: https://vt100.net/docs/vt510-rm/CPL.html
func PreviousLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b" + "[" + s + "F"
}

// Move (CUP) returns a sequence for moving the cursor to the given row and
// column.
//
//	CSI n ; m H
//
// See: https://vt100.net/docs/vt510-rm/CUP.html
func Move(row, col int) string {
	var r, c string
	if row > 1 {
		r = strconv.Itoa(row)
	}
	if col > 1 {
		c = strconv.Itoa(col)
	}
	return "\x1b" + "[" + r + ";" + c + "H"
}

// SavePosition (SCP or SCOSC) is a sequence for saving the cursor position.
//
//	CSI s
//
// This acts like Save, except the page number where the cursor is located is
// not saved.
//
// See: https://vt100.net/docs/vt510-rm/SCOSC.html
const SavePosition = "\x1b" + "[" + "s"

// RestorePosition (RCP or SCORC) is a sequence for restoring the cursor position.
//
//	CSI u
//
// This acts like Restore, except the cursor stays on the same page where the
// cursor was saved.
//
// See: https://vt100.net/docs/vt510-rm/SCORC.html
const RestorePosition = "\x1b" + "[" + "u"
