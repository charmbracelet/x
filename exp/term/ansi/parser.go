package ansi

import (
	"math"
	"unicode/utf8"
)

// DefaultMaxIntermediates is the maximum number of intermediates bytes allowed.
const DefaultMaxIntermediates = 2

// DefaultMaxOscBytes is the maximum number of bytes allowed in an Osc parameter.
const DefaultMaxOscBytes = 1024

// DefaultMaxOscParameters is the default maximum number of Osc parameters allowed.
const DefaultMaxOscParameters = 16

// DefaultMaxParameters is the maximum number of parameters allowed.
const DefaultMaxParameters = 32

// Handler is an interface for parsing.
type Handler struct {
	// Rune is called when a print action is performed.
	// The rune is a utf8 encoded rune.
	Rune func(r rune)

	// Execute is called when an execute action is performed.
	// This is usually a control character.
	Execute func(b byte)

	// EscHandler is called when an esc dispatch action is performed.
	EscHandler func(inter byte, final byte, ignore bool)

	// CsiHandler is called when a csi dispatch action is performed.
	CsiHandler func(marker byte, params [][]uint, inter byte, final byte, ignore bool)

	// OscHandler is called when an osc dispatch action is performed.
	OscHandler func(params [][]byte, bellTerminated bool)

	// DcsHandler is called when a dcs dispatch action is performed.
	DcsHandler func(marker byte, params [][]uint, inter byte, final byte, data []byte, ignore bool)
}

// Parser represents a state machine.
type Parser struct {
	handler Handler

	oscParams [][2]int

	oscRaw []byte

	// params holds the parameters for the current sequence including sub
	// parameters.
	params [32]uint

	// numSubParams holds the number of sub parameters for each parameter.
	numSubParams [32]int
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
	inters [2]byte

	state State

	// ignoring is set to true when the number of parameters exceeds the
	// maximum allowed. This is to prevent the parser from consuming too much
	// memory.
	ignoring bool
}

// New returns a new DEC ANSI compatible sequence parser.
func New(
	handler Handler,
) *Parser {
	p := &Parser{
		state:     GroundState,
		handler:   handler,
		oscRaw:    make([]byte, 0, DefaultMaxOscBytes),
		oscParams: make([][2]int, DefaultMaxOscParameters),
	}

	return p
}

// Parse parses the given reader until eof.
func (p *Parser) Parse(buf []byte) {
	for i, b := range buf {
		p.advance(b, i < len(buf)-1)
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
	if p.handler.Rune != nil {
		p.handler.Rune(r)
	}
	p.state = GroundState
	p.clearUtf8()
}

// advance advances the state machine.
func (p *Parser) advance(code byte, more bool) {
	if p.state == Utf8State {
		p.advanceUtf8(code)
	} else {
		state, action := table.Transition(p.state, code)
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
			// TODO: implement
		}
	}

	p.performAction(action, code)

	if p.state != state {
		switch state {
		case CsiEntryState, DcsEntryState, EscapeState:
			p.performAction(ClearAction, code)
		case OscStringState:
			p.performAction(OscStartAction, code)
		case DcsPassthroughState:
			p.performAction(DcsHookAction, code)
		case SosPmApcStringState:
			// TODO: Implement
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
		if p.handler.Rune != nil {
			p.handler.Rune(rune(code))
		}

	case ExecuteAction:
		if p.handler.Execute != nil {
			p.handler.Execute(code)
		}

	case EscDispatchAction:
		if p.handler.EscHandler != nil {
			p.handler.EscHandler(
				p.inters[1],
				code,
				p.ignoring,
			)
		}

	case OscStartAction:
		p.oscRaw = make([]byte, 0)

	case OscPutAction:
		idx := len(p.oscRaw)
		if code == ';' {
			paramIdx := p.oscNumParams
			switch paramIdx {
			case DefaultMaxOscParameters:
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
			p.oscRaw = append(p.oscRaw, code)
		}

	case OscEndAction:
		paramIdx := p.oscNumParams
		idx := len(p.oscRaw)

		switch paramIdx {
		case DefaultMaxOscParameters:
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

		if p.handler.OscHandler != nil {
			p.handler.OscHandler(
				p.getOscParams(),
				code == BEL,
			)
		}

	case DcsHookAction:
		p.oscRaw = make([]byte, 0)
		if p.isParamsFull() {
			p.ignoring = true
		} else {
			p.pushParam(p.param)
		}
		p.param = uint(code)

	case DcsPutAction:
		p.oscRaw = append(p.oscRaw, code)

	case DcsUnhookAction:
		if p.handler.DcsHandler != nil {
			p.handler.DcsHandler(
				p.inters[0],
				p.getParams(),
				p.inters[1],
				byte(p.param),
				p.oscRaw,
				p.ignoring,
			)
		}

	case CsiDispatchAction:
		if p.isParamsFull() {
			p.ignoring = true
		} else {
			p.pushParam(p.param)
		}

		if p.handler.CsiHandler != nil {
			p.handler.CsiHandler(
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
	if p.intersLen == DefaultMaxIntermediates {
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

	params := make([][]byte, 0, DefaultMaxOscParameters)
	for i := 0; i < p.oscNumParams; i++ {
		indices := p.oscParams[i]
		param := p.oscRaw[indices[0]:indices[1]]
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
	return p.paramsLen >= DefaultMaxParameters
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
