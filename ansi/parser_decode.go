package ansi

import (
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// DecodeSequence decodes a single ANSI escape sequence or a printable grapheme
// from the given data. It returns the sequence slice, the number of bytes
// read, the cell width, and the new state.
func DecodeSequence[T string | []byte](b T, state byte) (seq T, width int, n int, newState byte) {
	// log.Printf("DecodeSequence(%q, %d)", b, state)
	// defer func() {
	// 	log.Printf("DecodeSequence(%q, %d) -> (%q, %d, %d, %d)", b, state, seq, width, n, newState)
	// }()
	for i := 0; i < len(b); i++ {
		var action byte
		newState, action = parser.Table.Transition(state, b[i])
		switch action {
		case parser.PrintAction:
			return b[:i+1], 1, i + 1, newState
		case parser.ExecuteAction:
			return b[:i+1], 0, i + 1, newState
		case parser.IgnoreAction:
			switch b[i] {
			case DEL:
				if state == parser.GroundState {
					// Special case for DEL, which is ignored in the transition table.
					return b[:i+1], 0, i + 1, newState
				}
			case ST:
				// Special case for ST, which is ignored in the transition table.
				return b[:i+1], 0, i + 1, newState
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
			switch action {
			case parser.DispatchAction:
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
					return b[:i+1], 0, i + 1, newState
				case HasEscPrefix(b):
					// Handle escape sequences
					return b[:i+1], 0, i + 1, newState
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
