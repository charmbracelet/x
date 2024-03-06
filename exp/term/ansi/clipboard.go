package ansi

import "encoding/base64"

// Clipboard names.
const (
	SystemClipboard  = 'c'
	PrimaryClipboard = 'p'
)

// SetClipboard returns a sequence for manipulating the clipboard.
//
//	OSC 52 ; Pc ; Pd ST
//	OSC 52 ; Pc ; Pd BEL
//
// Where Pc is the clipboard name and Pd is the base64 encoded data.
// Empty data or invalid base64 data will reset the clipboard.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func SetClipboard(c byte, d string) string {
	if d != "" {
		d = base64.StdEncoding.EncodeToString([]byte(d))
	}
	return "\x1b]52;" + string(c) + ";" + d + "\x07"
}

// ResetClipboard returns a sequence for resetting the clipboard.
//
// This is equivalent to SetClipboard(c, "").
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func ResetClipboard(c byte) string {
	return SetClipboard(c, "")
}

// RequestClipboard returns a sequence for requesting the clipboard.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func RequestClipboard(c byte) string {
	return "\x1b]52;" + string(c) + ";?\x07"
}
