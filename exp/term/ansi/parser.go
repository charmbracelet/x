package ansi

import (
	"math"
	"unicode/utf8"

	. "github.com/charmbracelet/x/exp/term/ansi/parser"
)

const (
	// maxIntermediates is the maximum number of intermediates bytes allowed.
	maxIntermediates = 2

	// maxBufferSize is the maximum number of bytes allowed in an Osc or Dcs sequence.
	// nolint: unused
	maxBufferSize = 1024

	// maxOscParameters is the default maximum number of Osc parameters allowed.
	maxOscParameters = 16

	// maxParameters is the maximum number of parameters allowed.
	maxParameters = 32
)

// Parser represents a DEC ANSI compatible sequence parser.
//
// It uses a state machine to parse ANSI escape sequences and control
// characters. The parser is designed to be used with a terminal emulator or
// similar application that needs to parse ANSI escape sequences and control
// characters.
// See [parser] for more information.
//
//go:generate go run ./gen.go
type Parser struct {
	// Print a rune to the output.
	Print func(r rune)

	// Execute a C0 or C1 control code.
	Execute func(b byte)

	// EscDispatch is called when the final byte of an escape sequence is
	// received.
	// The ignore flag indicates that the parser encountered more than one
	// intermediate byte and ignoring subsequent bytes.
	EscDispatch func(inter byte, final byte, ignore bool)

	// CsiDispatch is called when the final byte of a Control Sequence
	// Introducer (CSI) is received.
	// The ignore flag indicates that either the parser encountered more than
	// one marker and one intermediate bytes, or more than the maximum number
	// of parameters, and ignoring subsequent bytes.
	CsiDispatch func(marker byte, params [][]uint, inter byte, final byte, ignore bool)

	// OscDispatch is called to dispatch an Operating System Command (OSC).
	// bellTerminated indicates that the sequence was terminated by a BEL byte.
	OscDispatch func(params [][]byte, bellTerminated bool)

	// DcsDispatch is called to dispatch a Device Control String (DCS).
	// The ignore flag indicates that either the parser encountered more than
	// the maximum number of parameters, or intermediates, and ignoring
	// subsequent bytes.
	DcsDispatch func(marker byte, params [][]uint, inter byte, final byte, data []byte, ignore bool)

	// SosPmApcDispatch is called to dispatch a Start of String (SOS), Privacy
	// Message (PM), or Application Program Command (APC) sequence.
	// The kind byte indicates the type of sequence and can be one of the
	// following:
	//  - SOS: 0x98
	//  - PM: 0x9E
	//  - APC: 0x9F
	SosPmApcDispatch func(kind byte, data []byte)

	// buf holds the bytes of an Osc or Dcs sequence.
	buf []byte

	oscParams [maxOscParameters][2]int

	// params holds the parameters for the current sequence including sub
	// parameters.
	params [maxParameters]uint

	// numSubParams holds the number of sub parameters for each parameter.
	numSubParams [maxParameters]int
	paramsLen    int

	// param holds the current parameter.
	// When we get a DcsHookAction, we store the final byte in param.
	param uint

	subParamsLen int
	oscNumParams int
	intersLen    int

	// utf8Idx & utf8Raw are used to collect utf8 encoded runes.
	utf8Idx int
	utf8Raw [utf8.UTFMax]byte

	// inters holds 2 bytes, the 1st for the private marker and the 2nd for the
	// intermediate byte.
	// ECMA-48 5.4 doesn't specify a limit on the number of private parameter
	// or intermediate bytes, however, in practice, there isn't a need for more
	// than 2.
	inters [maxIntermediates]byte

	state State

	// ignoring is set to true when the number of parameters exceeds the
	// maximum allowed. This is to prevent the parser from consuming too much
	// memory.
	ignoring bool

	// The value here indicates the type of a SOS, PM, or APC sequence.
	sosPmApc byte
}

// Parse parses the given reader until eof.
func (p *Parser) Parse(buf []byte) {
	for i, b := range buf {
		p.Advance(b, i < len(buf)-1)
	}
}

func (p *Parser) advanceUtf8(code byte) {
	// Collect the byte into the array
	p.collectUtf8(code)
	rw := utf8ByteLen(p.utf8Raw[0])
	if rw == -1 {
		// We panic here because the first byte comes from the state machine,
		// if this panics, it means there is a bug in the state machine!
		panic("invalid rune") // unreachable
	}

	if p.utf8Idx < rw {
		return
	}

	// We have enough bytes to decode the rune
	bts := p.utf8Raw[:rw]
	r, _ := utf8.DecodeRune(bts)
	if p.Print != nil {
		p.Print(r)
	}
	p.state = GroundState
	p.clearUtf8()
}

// State returns the current state of the parser.
func (p *Parser) State() State {
	return p.state
}

// StateName returns the name of the current state.
func (p *Parser) StateName() string {
	return StateNames[p.state]
}

// Advance advances the state machine.
func (p *Parser) Advance(code byte, more bool) {
	if p.state == Utf8State {
		p.advanceUtf8(code)
	} else {
		state, action := Table.Transition(p.state, code)
		p.performStateChange(state, action, code, more)
	}
}

func (p *Parser) performEscapeStateChange(code byte, more bool) {
	switch p.state {
	case GroundState:
		if !more {
			// End of input, execute Esc
			p.performAction(ExecuteAction, code)
		}
	case EscapeState:
		// More input mean possible Esc sequence, execute the previous Esc
		p.performAction(ExecuteAction, code)
		if !more {
			// No more input means execute the current Esc
			p.performAction(ExecuteAction, code)
		}
	default:
		if !more {
			// No more input means execute the current Esc
			p.performAction(ExecuteAction, code)
		}
	}
}

func (p *Parser) performStateChange(state State, action Action, code byte, more bool) {
	// Handle Esc execute action
	if code == ESC {
		p.performEscapeStateChange(code, more)
	}

	if p.state != state {
		switch p.state {
		case DcsPassthroughState:
			p.performAction(DcsUnhookAction, code)
		case OscStringState:
			p.performAction(OscEndAction, code)
		case SosPmApcStringState:
			p.performAction(SosPmApcEndAction, code)
		}
	}

	p.performAction(action, code)

	if p.state != state {
		switch state {
		case CsiEntryState, DcsEntryState, EscapeState:
			p.performAction(ClearAction, code)
		case SosPmApcStringState:
			switch code {
			case SOS, 'X':
				p.sosPmApc = SOS
			case PM, '^':
				p.sosPmApc = PM
			case APC, '_':
				p.sosPmApc = APC
			}
			fallthrough
		case OscStringState:
			p.performAction(StartAction, code)
		case DcsPassthroughState:
			p.performAction(DcsHookAction, code)
		}
	}

	p.state = state
}

func (p *Parser) performAction(action Action, code byte) {
	// log.Printf("performing action: %s, code: %q", actionNames[action], code)
	switch action {
	case NoneAction:
		break

	case IgnoreAction:
		break

	case PrintAction:
		if p.Print != nil {
			p.Print(rune(code))
		}

	case ExecuteAction:
		if p.Execute != nil {
			p.Execute(code)
		}

	case EscDispatchAction:
		if p.EscDispatch != nil {
			p.EscDispatch(
				p.inters[1],
				code,
				p.ignoring,
			)
		}

	case SosPmApcEndAction:
		if p.SosPmApcDispatch != nil {
			p.SosPmApcDispatch(p.sosPmApc, p.buf)
		}

	case StartAction:
		p.buf = make([]byte, 0)

	case OscPutAction:
		idx := len(p.buf)
		if code == ';' {
			paramIdx := p.oscNumParams
			switch paramIdx {
			case maxOscParameters:
				return
			case 0:
				p.oscParams[paramIdx] = [2]int{0, idx}
			default:
				prev := p.oscParams[paramIdx-1]
				begin := prev[1]
				p.oscParams[paramIdx] = [2]int{begin, idx}
			}
			p.oscNumParams++
		} else {
			p.buf = append(p.buf, code)
		}

	case OscEndAction:
		paramIdx := p.oscNumParams
		idx := len(p.buf)

		switch paramIdx {
		case maxOscParameters:
			break
		case 0:
			p.oscParams[paramIdx] = [2]int{0, idx}
			p.oscNumParams++
		default:
			prev := p.oscParams[paramIdx-1]
			begin := prev[1]
			p.oscParams[paramIdx] = [2]int{begin, idx}
			p.oscNumParams++
		}

		if p.OscDispatch != nil {
			p.OscDispatch(
				p.getOscParams(),
				code == BEL,
			)
		}

	case DcsHookAction:
		p.buf = make([]byte, 0)
		if p.isParamsFull() {
			p.ignoring = true
		} else if p.param > 0 || p.paramsLen > 0 {
			p.pushParam(p.param)
		}
		p.param = uint(code)

	case PutAction:
		p.buf = append(p.buf, code)

	case DcsUnhookAction:
		if p.DcsDispatch != nil {
			p.DcsDispatch(
				p.inters[0],
				p.getParams(),
				p.inters[1],
				byte(p.param),
				p.buf,
				p.ignoring,
			)
		}

	case CsiDispatchAction:
		if p.isParamsFull() {
			p.ignoring = true
		} else if p.param > 0 || p.paramsLen > 0 {
			p.pushParam(p.param)
		}

		if p.CsiDispatch != nil {
			p.CsiDispatch(
				p.inters[0],
				p.getParams(),
				p.inters[1],
				code,
				p.ignoring,
			)
		}

	case CollectAction:
		if utf8ByteLen(code) > 1 {
			p.collectUtf8(code)
		} else {
			p.collect(code)
		}

	case ParamAction:
		if p.isParamsFull() {
			p.ignoring = true
			return
		}

		switch code {
		case ';':
			p.pushParam(p.param)
			p.param = 0
		case ':':
			p.extendParam(p.param)
			p.param = 0
		default:
			p.param = smulu(p.param, 10)
			p.param = saddu(p.param, uint(code-'0'))
		}

	case ClearAction:
		p.clear()
	}
}

func (p *Parser) collectUtf8(code byte) {
	if p.utf8Idx < utf8.UTFMax {
		p.utf8Raw[p.utf8Idx] = code
		p.utf8Idx++
	}
}

func (p *Parser) collect(code byte) {
	if p.intersLen == maxIntermediates {
		p.ignoring = true
	} else if code >= 0x30 && code <= 0x3F { // private marker
		p.inters[0] = code
	} else {
		p.inters[1] = code
	}
	p.intersLen++
}

func (p *Parser) clearUtf8() {
	p.utf8Idx = 0
}

func (p *Parser) clear() {
	// Reset everything on ESC/CSI/DCS entry
	p.intersLen = 0
	p.ignoring = false
	p.param = 0
	p.paramsLen = 0
	p.subParamsLen = 0
	p.inters[0], p.inters[1] = 0, 0
}

func (p *Parser) getOscParams() [][]byte {
	if p.oscNumParams == 0 {
		return nil
	}

	params := make([][]byte, 0, maxOscParameters)
	for i := 0; i < p.oscNumParams; i++ {
		indices := p.oscParams[i]
		param := p.buf[indices[0]:indices[1]]
		params = append(params, param)
	}

	return params[:p.oscNumParams]
}

func (p *Parser) pushParam(param uint) {
	p.numSubParams[p.paramsLen-p.subParamsLen] = p.subParamsLen + 1
	p.params[p.paramsLen] = param
	p.subParamsLen = 0
	p.paramsLen++
}

func (p *Parser) extendParam(param uint) {
	p.numSubParams[p.paramsLen-p.subParamsLen] = p.subParamsLen + 1
	p.params[p.paramsLen] = param
	p.subParamsLen++
	p.paramsLen++
}

func (p *Parser) isParamsFull() bool {
	return p.paramsLen >= maxParameters
}

func (p *Parser) getParams() [][]uint {
	if p.paramsLen == 0 {
		return nil
	}
	params := make([][]uint, 0)
	for i := 0; i < p.paramsLen; {
		nSubs := p.numSubParams[i]
		subs := p.params[i : i+nSubs]
		i += nSubs
		params = append(params, subs)
	}
	return params
}

func saddu(a, b uint) uint {
	if b > 0 && a > math.MaxUint-b {
		return math.MaxUint
	}

	return a + b
}

func smulu(a, b uint) uint {
	if a > 0 && b > 0 && a > math.MaxUint/b {
		return math.MaxUint
	}

	return a * b
}

func utf8ByteLen(b byte) int {
	if b <= 0b0111_1111 { // 0x00-0x7F
		return 1
	} else if b >= 0b1100_0000 && b <= 0b1101_1111 { // 0xC0-0xDF
		return 2
	} else if b >= 0b1110_0000 && b <= 0b1110_1111 { // 0xE0-0xEF
		return 3
	} else if b >= 0b1111_0000 && b <= 0b1111_0111 { // 0xF0-0xF7
		return 4
	}
	return -1
}
