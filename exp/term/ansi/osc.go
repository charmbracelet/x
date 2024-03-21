package ansi

import (
	"strconv"
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
	// Data contains the raw data of the sequence.
	Data []byte

	// Cmd contains the raw command of the sequence.
	Cmd int
}

// NewOscSequence returns a new OscSequence.
// This is synonym for OscSequence{}.
func NewOscSequence() OscSequence {
	return OscSequence{}
}

// Command returns the command of the OSC sequence.
func (s OscSequence) Command() int {
	return s.Cmd
}

// String returns the string representation of the OSC sequence.
// To be more compatible with different terminal, this will always return a
// 7-bit formatted sequence, terminated by BEL.
func (s OscSequence) String() string {
	var b strings.Builder
	b.WriteString("\x1b]")
	b.WriteString(strconv.Itoa(s.Cmd))
	b.WriteByte(';')
	b.Write(s.Data)
	b.WriteByte(BEL)
	return b.String()
}

// Bytes returns the byte representation of the OSC sequence.
// To be more compatible with different terminal, this will always return a
// 7-bit formatted sequence, terminated by BEL.
func (s OscSequence) Bytes() []byte {
	return []byte(s.String())
}
