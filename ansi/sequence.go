package ansi

// Sequence represents an ANSI sequence. This can be a control sequence, escape
// sequence, a printable character, etc.
// A Sequence can be one of the following types:
//   - [Rune]
//   - [ControlCode]
//   - [Grapheme]
//   - [EscSequence]
//   - [CsiSequence]
//   - [OscSequence]
//   - [DcsSequence]
//   - [SosSequence]
//   - [PmSequence]
//   - [ApcSequence]
type Sequence interface {
	// Clone returns a deep copy of the sequence.
	Clone() Sequence
}

// Rune represents a printable character.
type Rune rune

var _ Sequence = Rune(0)

// Clone returns a deep copy of the rune.
func (r Rune) Clone() Sequence {
	return r
}

// Grapheme represents a grapheme cluster.
type Grapheme struct {
	Cluster string
	Width   int
}

var _ Sequence = Grapheme{}

// Clone returns a deep copy of the grapheme.
func (g Grapheme) Clone() Sequence {
	return g
}

// ControlCode represents a control code character. This is a character that
// is not printable and is used to control the terminal. This would be a
// character in the C0 or C1 set in the range of 0x00-0x1F and 0x80-0x9F.
type ControlCode byte

var _ Sequence = ControlCode(0)

// Clone returns a deep copy of the control code.
func (c ControlCode) Clone() Sequence {
	return c
}

// EscSequence represents an escape sequence.
type EscSequence Command

var _ Sequence = EscSequence(0)

// Clone returns a deep copy of the escape sequence.
func (e EscSequence) Clone() Sequence {
	return e
}

// Command returns the command byte of the escape sequence.
func (e EscSequence) Command() int {
	return Command(e).Command()
}

// Intermediate returns the intermediate byte of the escape sequence.
func (e EscSequence) Intermediate() int {
	return Command(e).Intermediate()
}

// SosSequence represents a SOS sequence.
type SosSequence struct {
	// Data contains the raw data of the sequence.
	Data []byte
}

var _ Sequence = SosSequence{}

// Clone returns a deep copy of the SOS sequence.
func (s SosSequence) Clone() Sequence {
	return SosSequence{
		Data: append([]byte(nil), s.Data...),
	}
}

// PmSequence represents a PM sequence.
type PmSequence struct {
	// Data contains the raw data of the sequence.
	Data []byte
}

var _ Sequence = PmSequence{}

// Clone returns a deep copy of the PM sequence.
func (p PmSequence) Clone() Sequence {
	return PmSequence{
		Data: append([]byte(nil), p.Data...),
	}
}

// ApcSequence represents an APC sequence.
type ApcSequence struct {
	// Data contains the raw data of the sequence.
	Data []byte
}

var _ Sequence = ApcSequence{}

// Clone returns a deep copy of the APC sequence.
func (a ApcSequence) Clone() Sequence {
	return ApcSequence{
		Data: append([]byte(nil), a.Data...),
	}
}
