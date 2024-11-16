package ansi

import "strconv"

// Status represents a terminal status report.
type Status interface {
	// Status returns the status report identifier.
	Status() int
}

// ANSIStatus represents an ANSI terminal status report.
type ANSIStatus int //nolint:revive

// Status returns the status report identifier.
func (s ANSIStatus) Status() int {
	return int(s)
}

// DECStatus represents a DEC terminal status report.
type DECStatus int

// Status returns the status report identifier.
func (s DECStatus) Status() int {
	return int(s)
}

// DeviceStatusReport (DSR) is a control sequence that reports the terminal's
// status.
// The terminal responds with a DSR sequence.
//
//	CSI Ps n
//	CSI ? Ps n
//
// See also https://vt100.net/docs/vt510-rm/DSR.html
func DeviceStatusReport(status Status) string {
	switch status := status.(type) {
	case ANSIStatus:
		return "\x1b[" + strconv.Itoa(status.Status()) + "n"
	case DECStatus:
		return "\x1b[?" + strconv.Itoa(status.Status()) + "n"
	}
	return ""
}

// DSR is an alias for [DeviceStatusReport].
func DSR(status Status) string {
	return DeviceStatusReport(status)
}

// CursorPositionReport (CPR) is a control sequence that reports the cursor's
// position.
//
//	CSI Pl ; Pc R
//
// Where Pl is the line number and Pc is the column number.
//
// See also https://vt100.net/docs/vt510-rm/CPR.html
func CursorPositionReport(line, column int) string {
	if line < 1 {
		line = 1
	}
	if column < 1 {
		column = 1
	}
	return "\x1b[" + strconv.Itoa(line) + ";" + strconv.Itoa(column) + "R"
}

// CPR is an alias for [CursorPositionReport].
func CPR(line, column int) string {
	return CursorPositionReport(line, column)
}

// ExtendedCursorPositionReport (DECXCPR) is a control sequence that reports the
// cursor's position along with the page number (optional).
//
//	CSI ? Pl ; Pc ; Pv R
//
// Where Pl is the line number, Pc is the column number, and Pv is the page
// number.
//
// If the page number is zero, the returned sequence won't include the page
// number.
//
// See also https://vt100.net/docs/vt510-rm/DECXCPR.html
func ExtendedCursorPositionReport(line, column, page int) string {
	if line < 1 {
		line = 1
	}
	if column < 1 {
		column = 1
	}
	if page < 1 {
		return "\x1b[?" + strconv.Itoa(line) + ";" + strconv.Itoa(column) + "R"
	}
	return "\x1b[?" + strconv.Itoa(line) + ";" + strconv.Itoa(column) + ";" + strconv.Itoa(page) + "R"
}

// DECXCPR is an alias for [ExtendedCursorPositionReport].
func DECXCPR(line, column, page int) string {
	return ExtendedCursorPositionReport(line, column, page)
}
