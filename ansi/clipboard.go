package ansi

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// Clipboard names.
const (
	SystemClipboard  = 'c'
	PrimaryClipboard = 'p'
)

// WriteSetClipboard writes the sequence to the given writer.
//
// Set the clipboard named c to the data d. Use [SystemClipboard] or
// [PrimaryClipboard] for c.
//
//	OSC 52 ; Pc ; Pd ST
//	OSC 52 ; Pc ; Pd BEL
//
// Where Pc is the clipboard name and Pd is the base64 encoded data.
// Empty data or invalid base64 data will reset the clipboard.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func WriteSetClipboard(w io.Writer, c byte, d string) (int, error) {
	if len(d) == 0 {
		return fmt.Fprintf(w, "\x1b]52;%c;\x07", c)
	}
	return fmt.Fprintf(w, "\x1b]52;%c;%s\x07", c, base64.StdEncoding.EncodeToString([]byte(d)))
}

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
	var b strings.Builder
	WriteSetClipboard(&b, c, d)
	return b.String()
}

// WriteSetSystemClipboard writes the sequence to set the system clipboard to
// the given writer.
//
// This is equivalent to [WriteSetClipboard](w, SystemClipboard, d).
func WriteSetSystemClipboard(w io.Writer, d string) (int, error) {
	return WriteSetClipboard(w, SystemClipboard, d)
}

// SetSystemClipboard returns a sequence for setting the system clipboard.
//
// This is equivalent to [SetClipboard](SystemClipboard, d).
func SetSystemClipboard(d string) string {
	return SetClipboard(SystemClipboard, d)
}

// WriteSetPrimaryClipboard writes the sequence to set the primary clipboard to
// the given writer.
//
// This is equivalent to WriteSetClipboard(w, PrimaryClipboard, d).
func WriteSetPrimaryClipboard(w io.Writer, d string) (int, error) {
	return WriteSetClipboard(w, PrimaryClipboard, d)
}

// SetPrimaryClipboard returns a sequence for setting the primary clipboard.
//
// This is equivalent to SetClipboard(PrimaryClipboard, d).
func SetPrimaryClipboard(d string) string {
	return SetClipboard(PrimaryClipboard, d)
}

// WriteResetClipboard writes the sequence to reset the clipboard to the given
// writer.
//
// This is equivalent to WriteSetClipboard(w, c, "").
func WriteResetClipboard(w io.Writer, c byte) (int, error) {
	return WriteSetClipboard(w, c, "")
}

// ResetClipboard returns a sequence for resetting the clipboard.
//
// This is equivalent to SetClipboard(c, "").
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func ResetClipboard(c byte) string {
	return SetClipboard(c, "")
}

// WriteRequestClipboard writes the sequence to request the clipboard to the
// given writer.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func WriteRequestClipboard(w io.Writer, c byte) (int, error) {
	return fmt.Fprintf(w, "\x1b]52;%c;?\x07", c)
}

// RequestClipboard returns a sequence for requesting the clipboard.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func RequestClipboard(c byte) string {
	return "\x1b]52;" + string(c) + ";?\x07"
}
