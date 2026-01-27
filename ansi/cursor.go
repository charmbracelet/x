package ansi

import (
	"strconv"
)

// RequestCursorPosition is an escape sequence that requests the current cursor
// position.
//
//	CSI 6 n
//
// The terminal will report the cursor position as a CSI sequence in the
// following format:
//
//	CSI Pl ; Pc R
//
// Where Pl is the line number and Pc is the column number.
// See: https://vt100.net/docs/vt510-rm/CPR.html
//
// Deprecated: use [RequestCursorPositionReport] instead.
const RequestCursorPosition = "\x1b[6n"

// RequestExtendedCursorPosition (DECXCPR) is a sequence for requesting the
// cursor position report including the current page number.
//
//	CSI ? 6 n
//
// The terminal will report the cursor position as a CSI sequence in the
// following format:
//
//	CSI ? Pl ; Pc ; Pp R
//
// Where Pl is the line number, Pc is the column number, and Pp is the page
// number.
// See: https://vt100.net/docs/vt510-rm/DECXCPR.html
//
// Deprecated: use [RequestExtendedCursorPositionReport] instead.
const RequestExtendedCursorPosition = "\x1b[?6n"

// CursorUp1 is a sequence for moving the cursor up one cell.
//
// This is equivalent to CursorUp(1).
//
// Deprecated: use [CUU1] instead.
const CursorUp1 = "\x1b[A"

// CursorDown1 is a sequence for moving the cursor down one cell.
//
// This is equivalent to CursorDown(1).
//
// Deprecated: use [CUD1] instead.
const CursorDown1 = "\x1b[B"

// CUF1 is a sequence for moving the cursor right one cell.
const CUF1 = "\x1b[C"

// CursorRight (CUF) returns a sequence for moving the cursor right n cells.
//
//	CSI n C
//
// See: https://vt100.net/docs/vt510-rm/CUF.html
//
// Deprecated: use [CursorForward] instead.
func CursorRight(n int) string {
	return CursorForward(n)
}

// CursorRight1 is a sequence for moving the cursor right one cell.
//
// This is equivalent to CursorRight(1).
//
// Deprecated: use [CUF1] instead.
const CursorRight1 = CUF1

// CUB1 is a sequence for moving the cursor left one cell.
const CUB1 = "\x1b[D"

// CursorLeft (CUB) returns a sequence for moving the cursor left n cells.
//
//	CSI n D
//
// See: https://vt100.net/docs/vt510-rm/CUB.html
//
// Deprecated: use [CursorBackward] instead.
func CursorLeft(n int) string {
	return CursorBackward(n)
}

// CursorLeft1 is a sequence for moving the cursor left one cell.
//
// This is equivalent to CursorLeft(1).
//
// Deprecated: use [CUB1] instead.
const CursorLeft1 = CUB1

// SetCursorPosition (CUP) returns a sequence for setting the cursor to the
// given row and column.
//
//	CSI n ; m H
//
// See: https://vt100.net/docs/vt510-rm/CUP.html
//
// Deprecated: use [CursorPosition] instead.
func SetCursorPosition(col, row int) string {
	if row <= 0 && col <= 0 {
		return HomeCursorPosition
	}

	var r, c string
	if row > 0 {
		r = strconv.Itoa(row)
	}
	if col > 0 {
		c = strconv.Itoa(col)
	}
	return "\x1b[" + r + ";" + c + "H"
}

// HomeCursorPosition is a sequence for moving the cursor to the upper left
// corner of the scrolling region. This is equivalent to `SetCursorPosition(1, 1)`.
//
// Deprecated: use [CursorHomePosition] instead.
const HomeCursorPosition = CursorHomePosition

// MoveCursor (CUP) returns a sequence for setting the cursor to the
// given row and column.
//
//	CSI n ; m H
//
// See: https://vt100.net/docs/vt510-rm/CUP.html
//
// Deprecated: use [CursorPosition] instead.
func MoveCursor(col, row int) string {
	return SetCursorPosition(col, row)
}

// CursorOrigin is a sequence for moving the cursor to the upper left corner of
// the display. This is equivalent to `SetCursorPosition(1, 1)`.
//
// Deprecated: use [CursorHomePosition] instead.
const CursorOrigin = "\x1b[1;1H"

// MoveCursorOrigin is a sequence for moving the cursor to the upper left
// corner of the display. This is equivalent to `SetCursorPosition(1, 1)`.
//
// Deprecated: use [CursorHomePosition] instead.
const MoveCursorOrigin = CursorOrigin

// CursorHorizontalForwardTab (CHT) returns a sequence for moving the cursor to
// the next tab stop n times.
//
// Default is 1.
//
//	CSI n I
//
// See: https://vt100.net/docs/vt510-rm/CHT.html
func CursorHorizontalForwardTab(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "I"
}

// CHT is an alias for [CursorHorizontalForwardTab].
func CHT(n int) string {
	return CursorHorizontalForwardTab(n)
}

// EraseCharacter (ECH) returns a sequence for erasing n characters from the
// screen. This doesn't affect other cell attributes.
//
// Default is 1.
//
//	CSI n X
//
// See: https://vt100.net/docs/vt510-rm/ECH.html
func EraseCharacter(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "X"
}

// ECH is an alias for [EraseCharacter].
func ECH(n int) string {
	return EraseCharacter(n)
}

// CursorBackwardTab (CBT) returns a sequence for moving the cursor to the
// previous tab stop n times.
//
// Default is 1.
//
//	CSI n Z
//
// See: https://vt100.net/docs/vt510-rm/CBT.html
func CursorBackwardTab(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "Z"
}

// CBT is an alias for [CursorBackwardTab].
func CBT(n int) string {
	return CursorBackwardTab(n)
}

// VerticalPositionAbsolute (VPA) returns a sequence for moving the cursor to
// the given row.
//
// Default is 1.
//
//	CSI n d
//
// See: https://vt100.net/docs/vt510-rm/VPA.html
func VerticalPositionAbsolute(row int) string {
	var s string
	if row > 0 {
		s = strconv.Itoa(row)
	}
	return "\x1b[" + s + "d"
}

// VPA is an alias for [VerticalPositionAbsolute].
func VPA(row int) string {
	return VerticalPositionAbsolute(row)
}

// VerticalPositionRelative (VPR) returns a sequence for moving the cursor down
// n rows relative to the current position.
//
// Default is 1.
//
//	CSI n e
//
// See: https://vt100.net/docs/vt510-rm/VPR.html
func VerticalPositionRelative(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "e"
}

// VPR is an alias for [VerticalPositionRelative].
func VPR(n int) string {
	return VerticalPositionRelative(n)
}

// HorizontalVerticalPosition (HVP) returns a sequence for moving the cursor to
// the given row and column.
//
// Default is 1,1.
//
//	CSI n ; m f
//
// This has the same effect as [CursorPosition].
//
// See: https://vt100.net/docs/vt510-rm/HVP.html
func HorizontalVerticalPosition(col, row int) string {
	var r, c string
	if row > 0 {
		r = strconv.Itoa(row)
	}
	if col > 0 {
		c = strconv.Itoa(col)
	}
	return "\x1b[" + r + ";" + c + "f"
}

// HVP is an alias for [HorizontalVerticalPosition].
func HVP(col, row int) string {
	return HorizontalVerticalPosition(col, row)
}

// HorizontalVerticalHomePosition is a sequence for moving the cursor to the
// upper left corner of the scrolling region. This is equivalent to
// `HorizontalVerticalPosition(1, 1)`.
const HorizontalVerticalHomePosition = "\x1b[f"

// SaveCursorPosition (SCP or SCOSC) is a sequence for saving the cursor
// position.
//
//	CSI s
//
// This acts like Save, except the page number where the cursor is located is
// not saved.
//
// See: https://vt100.net/docs/vt510-rm/SCOSC.html
//
// Deprecated: use [SaveCurrentCursorPosition] instead.
const SaveCursorPosition = "\x1b[s"

// RestoreCursorPosition (RCP or SCORC) is a sequence for restoring the cursor
// position.
//
//	CSI u
//
// This acts like Restore, except the cursor stays on the same page where the
// cursor was saved.
//
// See: https://vt100.net/docs/vt510-rm/SCORC.html
//
// Deprecated: use [RestoreCurrentCursorPosition] instead.
const RestoreCursorPosition = "\x1b[u"

// HorizontalPositionAbsolute (HPA) returns a sequence for moving the cursor to
// the given column. This has the same effect as [CUP].
//
// Default is 1.
//
//	CSI n \`
//
// See: https://vt100.net/docs/vt510-rm/HPA.html
func HorizontalPositionAbsolute(col int) string {
	var s string
	if col > 0 {
		s = strconv.Itoa(col)
	}
	return "\x1b[" + s + "`"
}

// HPA is an alias for [HorizontalPositionAbsolute].
func HPA(col int) string {
	return HorizontalPositionAbsolute(col)
}

// HorizontalPositionRelative (HPR) returns a sequence for moving the cursor
// right n columns relative to the current position. This has the same effect
// as [CUP].
//
// Default is 1.
//
//	CSI n a
//
// See: https://vt100.net/docs/vt510-rm/HPR.html
func HorizontalPositionRelative(n int) string {
	var s string
	if n > 0 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "a"
}

// HPR is an alias for [HorizontalPositionRelative].
func HPR(n int) string {
	return HorizontalPositionRelative(n)
}
