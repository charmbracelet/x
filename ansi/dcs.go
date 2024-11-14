package ansi

import "strings"

// DcsSequence represents a Device Control String (DCS) escape sequence.
//
// The DCS sequence is used to send device control strings to the terminal. The
// sequence starts with the C1 control code character DCS (0x9B) or ESC P in
// 7-bit environments, followed by parameter bytes, intermediate bytes, a
// command byte, followed by data bytes, and ends with the C1 control code
// character ST (0x9C) or ESC \ in 7-bit environments.
//
// This follows the parameter string format.
// See ECMA-48 § 5.4.1
type DcsSequence struct {
	// Params contains the raw parameters of the sequence.
	// This is a slice of integers, where each integer is a 32-bit integer
	// containing the parameter value in the lower 31 bits and a flag in the
	// most significant bit indicating whether there are more sub-parameters.
	Params []Parameter

	// Data contains the string raw data of the sequence.
	// This is the data between the final byte and the escape sequence terminator.
	Data []byte

	// Cmd contains the raw command of the sequence.
	// The command is a 32-bit integer containing the DCS command byte in the
	// lower 8 bits, the private marker in the next 8 bits, and the intermediate
	// byte in the next 8 bits.
	//
	//  DCS > 0 ; 1 $ r <data> ST
	//
	// Is represented as:
	//
	//  'r' | '>' << 8 | '$' << 16
	Cmd Command
}

var _ Sequence = DcsSequence{}

// Clone returns a deep copy of the DCS sequence.
func (s DcsSequence) Clone() Sequence {
	return DcsSequence{
		Params: append([]Parameter(nil), s.Params...),
		Data:   append([]byte(nil), s.Data...),
		Cmd:    s.Cmd,
	}
}

// Split returns a slice of data split by the semicolon.
func (s DcsSequence) Split() []string {
	return strings.Split(string(s.Data), ";")
}

// Marker returns the marker byte of the DCS sequence.
// This is always gonna be one of the following '<' '=' '>' '?' and in the
// range of 0x3C-0x3F.
// Zero is returned if the sequence does not have a marker.
func (s DcsSequence) Marker() int {
	return s.Cmd.Marker()
}

// Intermediate returns the intermediate byte of the DCS sequence.
// An intermediate byte is in the range of 0x20-0x2F. This includes these
// characters from ' ', '!', '"', '#', '$', '%', '&', ”', '(', ')', '*', '+',
// ',', '-', '.', '/'.
// Zero is returned if the sequence does not have an intermediate byte.
func (s DcsSequence) Intermediate() int {
	return s.Cmd.Intermediate()
}

// Command returns the command byte of the CSI sequence.
func (s DcsSequence) Command() int {
	return s.Cmd.Command()
}
