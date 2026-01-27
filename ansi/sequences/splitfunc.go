package sequences

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/stringish"
)

// ParamFunc is a callback function type for parsed parameters.
// i is the parameter index, val is the parameter value where a negative value
// indicates a missing parameter, hasMore indicates if there are
// sub-parameters.
type ParamFunc = func(i int, val int, hasMore bool)

// CmdFunc is a callback function for parsed commands escape sequences. The typ represents
// the type of sequence ([ansi.CSI] or [ansi.DCS]), prefix and intermed are
// intermediate bytes, and final is the final byte of the sequence.
type CmdFunc = func(typ byte, prefix byte, intermed byte, final byte)

// DataFunc is a callback function for parsed data in string sequences. The typ
// represents the type of sequence ([ansi.OSC], [ansi.APC], [ansi.SOS], or
// [ansi.PM]), data is the parsed data, and cancelled indicates if the sequence
// was cancelled i.e. terminated with [ansi.CAN] or [ansi.SUB].
type DataFunc[T stringish.Interface] = func(typ byte, data T, cancelled bool)

// ExecFunc is a callback function for parsed control characters.
type ExecFunc = func(b byte)

// PrintFunc is a callback function for parsed grapheme clusters.
type PrintFunc[T stringish.Interface] = func(data T, width int)

// state holds the state for the [SplitFunc].
type state[T stringish.Interface] struct {
	state byte
	width int // width of the last token
	// fields for parsing ANSI sequences
	typ                   byte
	param                 int
	paramIdx              int
	prefix, intermed, cmd byte
	dataIdx               int // the start index of data in string sequences
	// callbacks for parsed params and data
	paramFunc ParamFunc
	cmdFunc   CmdFunc
	dataFunc  DataFunc[T]
	execFunc  ExecFunc
	printFunc PrintFunc[T]
}

func newState[T stringish.Interface]() state[T] {
	return state[T]{
		state: ansi.NormalState,
		param: -1, // indicates no parameter parsed yet
	}
}

// splitFunc is a [bufio.SplitFunc]-compatible function that splits ANSI escape
// sequences and grapheme clusters.
func (s *state[T]) splitFunc(data T, atEOF bool) (n int, token T, err error) {
	var empty T
	if atEOF {
		return 0, empty, nil
	}

	for i := 0; i < len(data); i++ {
		c := data[i]

		switch s.state {
		case ansi.NormalState:
			switch c {
			case ansi.ESC:
				s.prefix = 0
				s.intermed = 0
				s.param = -1
				s.paramIdx = 0
				s.cmd = 0
				s.state = ansi.EscapeState
				s.typ = c
				continue
			case ansi.CSI, ansi.DCS:
				s.prefix = 0
				s.intermed = 0
				s.param = -1
				s.paramIdx = 0
				s.cmd = 0
				s.state = ansi.PrefixState
				s.typ = c
				continue
			case ansi.OSC, ansi.APC, ansi.SOS, ansi.PM:
				s.state = ansi.StringState
				s.typ = c
				s.dataIdx = i + 1
				continue
			}

			s.cmd = 0
			s.paramIdx = 0
			if c > ansi.US && c < ansi.DEL {
				// ASCII printable characters
				s.width = 1
				s.state = ansi.NormalState
				return 1, data[i : i+1], nil
			}

			if c <= ansi.US || c == ansi.DEL || c < 0xC0 {
				// C0 & C1 control characters & DEL
				s.state = ansi.NormalState
				s.width = 0
				if s.execFunc != nil {
					s.execFunc(c)
				}
				return 1, data[i : i+1], nil
			}

			if utf8.RuneStart(c) {
				token, s.width = ansi.FirstGraphemeCluster(data, ansi.GraphemeWidth)
				i += len(token)
				s.state = ansi.NormalState
				if s.printFunc != nil {
					s.printFunc(token, s.width)
				}
				return i, data[:i], nil
			}

			// Invalid UTF-8 sequence
			s.state = ansi.NormalState
			s.width = 0
			return i, data[:i], nil
		case ansi.PrefixState:
			if c >= '<' && c <= '?' {
				s.prefix = c
				break
			}

			s.state = ansi.ParamsState
			fallthrough
		case ansi.ParamsState:
			if c >= '0' && c <= '9' {
				if s.param < 0 {
					s.param = 0
				}
				s.param *= 10
				s.param += int(c - '0')
				break
			}

			if c == ';' || c == ':' {
				if s.paramFunc != nil {
					s.paramFunc(s.paramIdx, s.param, c == ':')
				}
				s.paramIdx++
				s.param = -1
				break
			}

			s.state = ansi.IntermedState
			fallthrough
		case ansi.IntermedState:
			if c >= ' ' && c <= '/' {
				s.intermed = c
				break
			}

			if s.param >= 0 {
				// Increment the last parameter
				if s.paramFunc != nil {
					s.paramFunc(s.paramIdx, s.param, false)
				}
				s.paramIdx++
				s.param = -1
			}

			if c >= '@' && c <= '~' {
				// Final byte of CSI/DCS sequence
				s.cmd = c
				if s.cmdFunc != nil {
					s.cmdFunc(s.typ, s.prefix, s.intermed, byte(s.cmd))
				}

				if ansi.HasDcsPrefix(data) {
					// Continue to collect DCS data
					s.state = ansi.StringState
					s.dataIdx = i + 1
					continue
				}

				// End of CSI sequence
				s.state = ansi.NormalState
				s.width = 0
				return i + 1, data[:i+1], nil
			}

			// Invalid CSI/DCS sequence
			s.state = ansi.NormalState
			s.width = 0
			return i, data[:i], nil
		case ansi.EscapeState:
			switch c {
			case '[', 'P':
				s.param = -1
				s.paramIdx = 0
				s.cmd = 0
				s.state = ansi.PrefixState
				if c == '[' {
					s.typ = ansi.CSI
				} else {
					s.typ = ansi.DCS
				}
				continue
			case ']', 'X', '^', '_':
				s.state = ansi.StringState
				s.dataIdx = i + 1
				switch c {
				case ']':
					s.typ = ansi.OSC
				case 'X':
					s.typ = ansi.APC
				case '^':
					s.typ = ansi.PM
				case '_':
					s.typ = ansi.SOS
				}
				continue
			}

			if c >= ' ' && c <= '/' {
				s.intermed = c
				continue
			} else if c >= '0' && c <= '~' {
				s.cmd = c
				s.state = ansi.NormalState
				s.width = 0
				return i + 1, data[:i+1], nil
			}

			// Invalid escape sequence
			s.state = ansi.NormalState
			s.width = 0
			return i, data[:i], nil
		case ansi.StringState:
			switch c {
			case ansi.BEL:
				if ansi.HasOscPrefix(data) {
					s.state = ansi.NormalState
					s.width = 0
					if s.dataFunc != nil {
						s.dataFunc(s.typ, data[s.dataIdx:i], false)
					}
					return i + 1, data[:i+1], nil
				}
			case ansi.CAN, ansi.SUB:
				if s.dataFunc != nil {
					s.dataFunc(s.typ, data[s.dataIdx:i], true)
				}

				// Cancel the sequence
				s.state = ansi.NormalState
				s.width = 0
				return i, data[:i], nil
			case ansi.ST:
				if s.dataFunc != nil {
					s.dataFunc(s.typ, data[s.dataIdx:i], false)
				}

				s.state = ansi.NormalState
				s.width = 0
				return i + 1, data[:i+1], nil
			case ansi.ESC:
				if ansi.HasStPrefix(data[i:]) {
					if s.dataFunc != nil {
						s.dataFunc(s.typ, data[s.dataIdx:i], false)
					}

					// End of string 7-bit (ST)
					s.state = ansi.NormalState
					s.width = 0
					return i + 2, data[:i+2], nil
				}

				if s.dataFunc != nil {
					s.dataFunc(s.typ, data[s.dataIdx:i], true)
				}

				// Otherwise, cancel the sequence
				s.state = ansi.NormalState
				s.width = 0
				return i, data[:i], nil
			}
		}
	}

	return len(data), data, nil
}

// SplitFunc is a [bufio.SplitFunc]-compatible function that splits ANSI escape
// sequences and grapheme clusters for the given stringish type.
func SplitFunc[T stringish.Interface](data T, atEOF bool) (n int, token T, err error) {
	var state state[T]
	return state.splitFunc(data, atEOF)
}
