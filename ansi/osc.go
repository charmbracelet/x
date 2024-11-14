package ansi

import (
	"strings"
)

// OscSequence represents an OSC sequence.
//
// The sequence starts with a OSC sequence, OSC (0x9D) in a 8-bit environment
// or ESC ] (0x1B 0x5D) in a 7-bit environment, followed by positive integer identifier,
// then by arbitrary data terminated by a ST (0x9C) in a 8-bit environment,
// ESC \ (0x1B 0x5C) in a 7-bit environment, or BEL (0x07) for backwards compatibility.
//
//	OSC Ps ; Pt ST
//	OSC Ps ; Pt BEL
//
// See ECMA-48 § 5.7.
type OscSequence struct {
	// Data contains the raw data of the sequence including the identifier
	// command.
	Data []byte

	// Cmd contains the raw command of the sequence.
	Cmd int
}

var _ Sequence = OscSequence{}

// Clone returns a deep copy of the OSC sequence.
func (o OscSequence) Clone() Sequence {
	return OscSequence{
		Data: append([]byte(nil), o.Data...),
		Cmd:  o.Cmd,
	}
}

// Split returns a slice of data split by the semicolon with the first element
// being the identifier command.
func (o OscSequence) Split() []string {
	return strings.Split(string(o.Data), ";")
}

// Command returns the OSC command. This is always gonna be a positive integer
// that identifies the OSC sequence.
func (o OscSequence) Command() int {
	return o.Cmd
}
