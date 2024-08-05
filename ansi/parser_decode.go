package ansi

import (
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// DecodeSequence decodes a single ANSI escape sequence or a printable grapheme
// from the given data. It returns the sequence slice, the number of bytes
// read, the cell width for each sequence, and the new state.
//
// The cell width will always be 0 for control and escape sequences, 1 for
// ASCII printable characters, and the number of cells other Unicode characters
// occupy. It uses the uniseg package to calculate the width of Unicode
// graphemes and characters.
//
// Passing a non-nil [*Parser] as the last argument will allow the decoder to
// collect sequence parameters, data, and commands. The parser cmd will have
// the packed command value that contains intermediate and marker characters.
// In the case of a OSC sequence, the cmd will be the OSC command number. Use
// [Cmd] and [Param] types to unpack command intermediates and markers as well
// as parameters.
//
// Example:
//
//	var state byte // the initial state is always zero [parser.GroundState]
//	p := NewParser(32, 1024) // create a new parser with a 32 params buffer and 1024 data buffer (optional)
//	input := []byte("\x1b[31mHello, World!\x1b[0m")
//	for len(input) > 0 {
//		seq, width, n, newState := DecodeSequence(input, state, p)
//		log.Printf("seq: %q, width: %d", seq, width)
//		state = newState
//		input = input[n:]
//	}
func DecodeSequence[T string | []byte](b T, state byte, p *Parser) (seq T, width int, n int, newState byte) {
	// log.Printf("DecodeSequence(%q, %d)", b, state)
	// defer func() {
	// 	log.Printf("DecodeSequence(%q, %d) -> (%q, %d, %d, %d)", b, state, seq, width, n, newState)
	// }()
	for i := 0; i < len(b); i++ {
		var action byte
		newState, action = parser.Table.Transition(state, b[i])
		if p != nil && state != newState && state == parser.EscapeState {
			// XXX: We need to clear the cmd when we transition from escape
			// state to be able to properly collect any intermediate characters
			// and the command byte in [p].
			p.Cmd = 0
		}
		switch action {
		case parser.PrintAction:
			return b[:i+1], 1, i + 1, newState
		case parser.ExecuteAction:
			return b[:i+1], 0, i + 1, newState
		case parser.DispatchAction:
			// Increment the last parameter
			if p != nil && (p.ParamsLen > 0 && p.ParamsLen < len(p.Params)-1 ||
				p.ParamsLen == 0 && len(p.Params) > 0 && p.Params[0] != parser.MissingParam) {
				p.ParamsLen++
			}

			// Handle ST, CAN, SUB, ESC
			switch {
			case HasOscPrefix(b):
				// Handle BEL terminated OSC
				if b[i] == BEL {
					return b[:i+1], 0, i + 1, newState
				}
				fallthrough
			case HasApcPrefix(b), HasDcsPrefix(b), HasPmPrefix(b), HasSosPrefix(b):
				if i < len(b) && HasStPrefix(b[i:]) {
					// Include ST in the sequence
					if b[i] == ESC {
						return b[:i+2], 0, i + 2, parser.GroundState
					}
					return b[:i+1], 0, i + 1, parser.GroundState
				}
				if b[i] == ESC || b[i] == CAN || b[i] == SUB {
					// Return unterminated sequence
					return b[:i], 0, i, newState
				}
				return b[:i+1], 0, i + 1, newState
			case HasCsiPrefix(b):
				if p != nil {
					p.Cmd |= int(b[i])
				}
				return b[:i+1], 0, i + 1, newState
			case HasEscPrefix(b):
				if p != nil {
					p.Cmd |= int(b[i])
				}
				// Handle escape sequences
				return b[:i+1], 0, i + 1, newState
			}
		case parser.ClearAction:
			if p == nil {
				break
			}
			if len(p.Params) > 0 {
				p.Params[0] = parser.MissingParam
			}
			p.Cmd = 0
			p.ParamsLen = 0
		case parser.MarkerAction:
			if p == nil {
				break
			}
			p.Cmd &^= 0xff << parser.MarkerShift
			p.Cmd |= int(b[i]) << parser.MarkerShift
		case parser.CollectAction:
			if p == nil {
				break
			}
			p.Cmd &^= 0xff << parser.IntermedShift
			p.Cmd |= int(b[i]) << parser.IntermedShift
		case parser.ParamAction:
			if p == nil {
				break
			}

			if p.ParamsLen >= len(p.Params) {
				break
			}

			if b[i] >= '0' && b[i] <= '9' {
				if p.Params[p.ParamsLen] == parser.MissingParam {
					p.Params[p.ParamsLen] = 0
				}

				p.Params[p.ParamsLen] *= 10
				p.Params[p.ParamsLen] += int(b[i] - '0')
			}

			if b[i] == ':' {
				p.Params[p.ParamsLen] |= parser.HasMoreFlag
			}

			if b[i] == ';' || b[i] == ':' {
				p.ParamsLen++
				if p.ParamsLen < len(p.Params) {
					p.Params[p.ParamsLen] = parser.MissingParam
				}
			}
		case parser.StartAction:
			if p == nil {
				break
			}

			p.DataLen = 0
			if state >= parser.DcsEntryState && state <= parser.DcsStringState {
				// Collect the command byte for DCS
				p.Cmd |= int(b[i])
			} else {
				p.Cmd = parser.MissingCommand
			}
		case parser.PutAction:
			if p == nil {
				break
			}

			if state == parser.DcsEntryState && newState == parser.DcsStringState {
				// XXX: This is a special case where we need to start collecting
				// non-string parameterized data i.e. doesn't follow the ECMA-48 §
				// 5.4.1 string parameters format.
				p.Cmd |= int(b[i])
			}

			if p.DataLen >= len(p.Data) {
				break
			}

			p.Data[p.DataLen] = b[i]
			p.DataLen++

			switch state {
			case parser.OscStringState:
				if b[i] == ';' && p.Cmd == parser.MissingCommand {
					// Try to parse the command
					for i := 0; i < len(p.Data); i++ {
						d := p.Data[i]
						if d < '0' || d > '9' {
							break
						}
						if p.Cmd == parser.MissingCommand {
							p.Cmd = 0
						}
						p.Cmd *= 10
						p.Cmd += int(d - '0')
					}
				}
			}
		}
		if state != newState {
			switch newState {
			case parser.Utf8State:
				cluster, _, width, _ := FirstGraphemeCluster(b[i:], -1)
				i += len(cluster)
				return b[:i], width, i, parser.GroundState
			case parser.EscapeState:
				if i < len(b)-1 {
					switch b[i+1] {
					case ESC:
						// Handle double escape
						return b[:i+1], 0, i + 1, parser.GroundState
					}
					if i > 0 && i < len(b) && !HasStPrefix(b[i:]) {
						// Handle unterminated escape sequence
						return b[:i], 0, i, newState
					}
				}
			}
			state = newState
		}
	}
	return b, 0, len(b), newState
}

// HasCsiPrefix returns true if the given byte slice has a CSI prefix.
func HasCsiPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == CSI) ||
		(len(b) > 1 && b[0] == ESC && b[1] == '[')
}

// HasOscPrefix returns true if the given byte slice has an OSC prefix.
func HasOscPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == OSC) ||
		(len(b) > 1 && b[0] == ESC && b[1] == ']')
}

// HasApcPrefix returns true if the given byte slice has an APC prefix.
func HasApcPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == APC) ||
		(len(b) > 1 && b[0] == ESC && b[1] == '_')
}

// HasDcsPrefix returns true if the given byte slice has a DCS prefix.
func HasDcsPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == DCS) ||
		(len(b) > 1 && b[0] == ESC && b[1] == 'P')
}

// HasSosPrefix returns true if the given byte slice has a SOS prefix.
func HasSosPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == SOS) ||
		(len(b) > 1 && b[0] == ESC && b[1] == 'X')
}

// HasPmPrefix returns true if the given byte slice has a PM prefix.
func HasPmPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == PM) ||
		(len(b) > 1 && b[0] == ESC && b[1] == '^')
}

// HasStPrefix returns true if the given byte slice has a ST prefix.
func HasStPrefix[T string | []byte](b T) bool {
	return (len(b) > 0 && b[0] == ST) ||
		(len(b) > 1 && b[0] == ESC && b[1] == '\\')
}

// HasEscPrefix returns true if the given byte slice has an ESC prefix.
func HasEscPrefix[T string | []byte](b T) bool {
	return len(b) > 0 && b[0] == ESC
}

// FirstGraphemeCluster returns the first grapheme cluster in the given string or byte slice.
// This is a syntactic sugar function that wraps
// uniseg.FirstGraphemeClusterInString and uniseg.FirstGraphemeCluster.
func FirstGraphemeCluster[T string | []byte](b T, state int) (T, T, int, int) {
	switch b := any(b).(type) {
	case string:
		cluster, rest, width, newState := uniseg.FirstGraphemeClusterInString(b, state)
		return T(cluster), T(rest), width, newState
	case []byte:
		cluster, rest, width, newState := uniseg.FirstGraphemeCluster(b, state)
		return T(cluster), T(rest), width, newState
	}
	panic("unreachable")
}

// Cmd represents a sequence command. This is used to pack/unpack a sequence
// command with its intermediate and marker characters. Those are commonly
// found in CSI and DCS sequences.
type Cmd int

// Marker returns the marker byte of the CSI sequence.
// This is always gonna be one of the following '<' '=' '>' '?' and in the
// range of 0x3C-0x3F.
// Zero is returned if the sequence does not have a marker.
func (c Cmd) Marker() int {
	return parser.Marker(int(c))
}

// Intermediate returns the intermediate byte of the CSI sequence.
// An intermediate byte is in the range of 0x20-0x2F. This includes these
// characters from ' ', '!', '"', '#', '$', '%', '&', ”', '(', ')', '*', '+',
// ',', '-', '.', '/'.
// Zero is returned if the sequence does not have an intermediate byte.
func (c Cmd) Intermediate() int {
	return parser.Intermediate(int(c))
}

// Command returns the command byte of the CSI sequence.
func (c Cmd) Command() int {
	return parser.Command(int(c))
}

// Param represents a sequence parameter. Sequence parameters with
// sub-parameters are packed with the HasMoreFlag set. This is used to unpack
// the parameters from a CSI and DCS sequences.
type Param int

// Param returns the parameter at the given index.
// It returns -1 if the parameter does not exist.
func (s Param) Param() int {
	return int(s) & parser.ParamMask
}

// HasMore returns true if the parameter has more sub-parameters.
func (s Param) HasMore() bool {
	return int(s)&parser.HasMoreFlag != 0
}
