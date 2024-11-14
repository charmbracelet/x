package ansi

// CsiSequence represents a control sequence introducer (CSI) sequence.
//
// The sequence starts with a CSI sequence, CSI (0x9B) in a 8-bit environment
// or ESC [ (0x1B 0x5B) in a 7-bit environment, followed by any number of
// parameters in the range of 0x30-0x3F, then by any number of intermediate
// byte in the range of 0x20-0x2F, then finally with a single final byte in the
// range of 0x20-0x7E.
//
//	CSI P..P I..I F
//
// See ECMA-48 § 5.4.
type CsiSequence struct {
	// Params contains the raw parameters of the sequence.
	// This is a slice of integers, where each integer is a 32-bit integer
	// containing the parameter value in the lower 31 bits and a flag in the
	// most significant bit indicating whether there are more sub-parameters.
	Params []Parameter

	// Cmd contains the raw command of the sequence.
	// The command is a 32-bit integer containing the CSI command byte in the
	// lower 8 bits, the private marker in the next 8 bits, and the intermediate
	// byte in the next 8 bits.
	//
	//  CSI ? u
	//
	// Is represented as:
	//
	//  'u' | '?' << 8
	Cmd Command
}

var _ Sequence = CsiSequence{}

// Clone returns a deep copy of the CSI sequence.
func (s CsiSequence) Clone() Sequence {
	return CsiSequence{
		Params: append([]Parameter(nil), s.Params...),
		Cmd:    s.Cmd,
	}
}

// Marker returns the marker byte of the CSI sequence.
// This is always gonna be one of the following '<' '=' '>' '?' and in the
// range of 0x3C-0x3F.
// Zero is returned if the sequence does not have a marker.
func (s CsiSequence) Marker() int {
	return s.Cmd.Marker()
}

// Intermediate returns the intermediate byte of the CSI sequence.
// An intermediate byte is in the range of 0x20-0x2F. This includes these
// characters from ' ', '!', '"', '#', '$', '%', '&', ”', '(', ')', '*', '+',
// ',', '-', '.', '/'.
// Zero is returned if the sequence does not have an intermediate byte.
func (s CsiSequence) Intermediate() int {
	return s.Cmd.Intermediate()
}

// Command returns the command byte of the CSI sequence.
func (s CsiSequence) Command() int {
	return s.Cmd.Command()
}
