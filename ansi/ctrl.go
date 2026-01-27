package ansi

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RequestXTVersion is a control sequence that requests the terminal's XTVERSION. It responds with a DSR sequence identifying the version.
//
//	CSI > Ps q
//	DCS > | text ST
//
// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
//
// Deprecated: use [RequestNameVersion] instead.
const RequestXTVersion = RequestNameVersion

// WritePrimaryDeviceAttributes writes the Primary Device Attributes (DA1)
// control sequence to the given writer.
//
//	CSI c
//	CSI 0 c
//	CSI ? Ps ; ... c
//
// If no attributes are given, or if the attribute is 0, this function returns
// the request sequence. Otherwise, it returns the response sequence.
//
// Common attributes include:
//   - 1	132 columns
//   - 2	Printer port
//   - 4	Sixel
//   - 6	Selective erase
//   - 7	Soft character set (DRCS)
//   - 8	User-defined keys (UDKs)
//   - 9	National replacement character sets (NRCS) (International terminal only)
//   - 12	Yugoslavian (SCS)
//   - 15	Technical character set
//   - 18	Windowing capability
//   - 21	Horizontal scrolling
//   - 23	Greek
//   - 24	Turkish
//   - 42	ISO Latin-2 character set
//   - 44	PCTerm
//   - 45	Soft key map
//   - 46	ASCII emulation
//
// See https://vt100.net/docs/vt510-rm/DA1.html
func WritePrimaryDeviceAttributes(w io.Writer, attrs ...int) (int, error) {
	if len(attrs) == 0 || (len(attrs) == 1 && attrs[0] <= 0) {
		return io.WriteString(w, RequestPrimaryDeviceAttributes)
	}

	// Fast path for 1, 2, 3, or 4 attributes.
	if len(attrs) <= 4 {
		switch len(attrs) {
		case 1:
			return fmt.Fprintf(w, "\x1b[?%dc", attrs[0])
		case 2:
			return fmt.Fprintf(w, "\x1b[?%d;%dc", attrs[0], attrs[1])
		case 3:
			return fmt.Fprintf(w, "\x1b[?%d;%d;%dc", attrs[0], attrs[1], attrs[2])
		case 4:
			return fmt.Fprintf(w, "\x1b[?%d;%d;%d;%dc", attrs[0], attrs[1], attrs[2], attrs[3])
		}
	}

	// General case for more than 4 attributes.

	b := strings.Builder{}

	// 4 for `ESC [ ?` and `c`, length of attrs minus 1 for semicolons, and 2
	// per attribute for digits (rough estimate).
	b.Grow(4 + (len(attrs) - 1) + len(attrs)*2)
	b.WriteString("\x1b[?")
	for i, a := range attrs {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(a))
	}
	b.WriteByte('c')

	return io.WriteString(w, b.String())
}

// PrimaryDeviceAttributes (DA1) is a control sequence that reports the
// terminal's primary device attributes.
//
//	CSI c
//	CSI 0 c
//	CSI ? Ps ; ... c
//
// If no attributes are given, or if the attribute is 0, this function returns
// the request sequence. Otherwise, it returns the response sequence.
//
// Common attributes include:
//   - 1	132 columns
//   - 2	Printer port
//   - 4	Sixel
//   - 6	Selective erase
//   - 7	Soft character set (DRCS)
//   - 8	User-defined keys (UDKs)
//   - 9	National replacement character sets (NRCS) (International terminal only)
//   - 12	Yugoslavian (SCS)
//   - 15	Technical character set
//   - 18	Windowing capability
//   - 21	Horizontal scrolling
//   - 23	Greek
//   - 24	Turkish
//   - 42	ISO Latin-2 character set
//   - 44	PCTerm
//   - 45	Soft key map
//   - 46	ASCII emulation
//
// See https://vt100.net/docs/vt510-rm/DA1.html
func PrimaryDeviceAttributes(attrs ...int) string {
	var b strings.Builder
	WritePrimaryDeviceAttributes(&b, attrs...)
	return b.String()
}

// WriteSecondaryDeviceAttributes writes the Secondary Device Attributes (DA2)
// control sequence to the given writer.
//
//	CSI > c
//	CSI > 0 c
//	CSI > Ps ; ... c
//
// See https://vt100.net/docs/vt510-rm/DA2.html
func WriteSecondaryDeviceAttributes(w io.Writer, attrs ...int) (int, error) {
	if len(attrs) == 0 || (len(attrs) == 1 && attrs[0] <= 0) {
		return io.WriteString(w, RequestSecondaryDeviceAttributes)
	}

	// Fast path for 1, 2, 3, or 4 attributes.
	if len(attrs) <= 4 {
		switch len(attrs) {
		case 1:
			return fmt.Fprintf(w, "\x1b[>%dc", attrs[0])
		case 2:
			return fmt.Fprintf(w, "\x1b[>%d;%dc", attrs[0], attrs[1])
		case 3:
			return fmt.Fprintf(w, "\x1b[>%d;%d;%dc", attrs[0], attrs[1], attrs[2])
		case 4:
			return fmt.Fprintf(w, "\x1b[>%d;%d;%d;%dc", attrs[0], attrs[1], attrs[2], attrs[3])
		}
	}

	// General case for more than 4 attributes.

	b := strings.Builder{}

	// 4 for `ESC [ >` and `c`, length of attrs minus 1 for semicolons, and 2
	// per attribute for digits (rough estimate).
	b.Grow(4 + (len(attrs) - 1) + len(attrs)*2)
	b.WriteString("\x1b[>")
	for i, a := range attrs {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(a))
	}
	b.WriteByte('c')

	return io.WriteString(w, b.String())
}

// SecondaryDeviceAttributes (DA2) is a control sequence that reports the
// terminal's secondary device attributes.
//
//	CSI > c
//	CSI > 0 c
//	CSI > Ps ; ... c
//
// See https://vt100.net/docs/vt510-rm/DA2.html
func SecondaryDeviceAttributes(attrs ...int) string {
	var b strings.Builder
	WriteSecondaryDeviceAttributes(&b, attrs...)
	return b.String()
}

// WriteTertiaryDeviceAttributes writes the Tertiary Device Attributes (DA3)
// control sequence to the given writer.
//
//	CSI = c
//	CSI = 0 c
//	DCS ! | Text ST
//
// Where Text is the unit ID for the terminal.
//
// If no unit ID is given, or if the unit ID is 0, this function returns the
// request sequence. Otherwise, it returns the response sequence.
//
// See https://vt100.net/docs/vt510-rm/DA3.html
func WriteTertiaryDeviceAttributes(w io.Writer, unitID string) (int, error) {
	if len(unitID) == 0 || unitID == "0" {
		return io.WriteString(w, RequestTertiaryDeviceAttributes)
	}

	return io.WriteString(w, "\x1bP!|"+unitID+"\x1b\\")
}

// TertiaryDeviceAttributes (DA3) is a control sequence that reports the
// terminal's tertiary device attributes.
//
//	CSI = c
//	CSI = 0 c
//	DCS ! | Text ST
//
// Where Text is the unit ID for the terminal.
//
// If no unit ID is given, or if the unit ID is 0, this function returns the
// request sequence. Otherwise, it returns the response sequence.
//
// See https://vt100.net/docs/vt510-rm/DA3.html
func TertiaryDeviceAttributes(unitID string) string {
	if len(unitID) == 0 || unitID == "0" {
		return RequestTertiaryDeviceAttributes
	}

	return "\x1bP!|" + unitID + "\x1b\\"
}
