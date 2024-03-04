package parser

import (
	"io"
	"math"
	"unicode/utf8"
)

// DefaultMaxIntermediates is the maximum number of intermediates bytes allowed.
const DefaultMaxIntermediates = 2

// DefaultMaxOscBytes is the maximum number of bytes allowed in an Osc parameter.
const DefaultMaxOscBytes = 1024

// DefaultMaxOscParameters is the default maximum number of Osc parameters allowed.
const DefaultMaxOscParameters = 16

// Handler is an interface for parsing.
type Handler interface {
	// Print is called when a print action is performed.
	// The rune is a utf8 encoded rune.
	Print(r rune)

	// Execute is called when an execute action is performed.
	// This is usually a control character.
	Execute(code byte)

	// EscDispatch is called when an esc dispatch action is performed.
	EscDispatch(intermediates []byte, final rune, ignore bool)

	// CsiDispatch is called when a csi dispatch action is performed.
	CsiDispatch(prefix string, params [][]uint16, intermediates []byte, final rune, ignore bool)

	// OscDispatch is called when an osc dispatch action is performed.
	OscDispatch(params [][]byte, bellTerminated bool)

	// DcsHook is called when a hook action is performed.
	DcsHook(prefix string, params [][]uint16, intermediates []byte, final rune, ignore bool)

	// DcsPut is called when a put action is performed.
	DcsPut(code byte)

	// DcsUnhook is called when an unhook action is performed.
	DcsUnhook()
}

type noopHandler struct{}

func (noopHandler) Print(rune)                                         {}
func (noopHandler) Execute(byte)                                       {}
func (noopHandler) DcsPut(byte)                                        {}
func (noopHandler) DcsUnhook()                                         {}
func (noopHandler) DcsHook(string, [][]uint16, []byte, rune, bool)     {}
func (noopHandler) OscDispatch([][]byte, bool)                         {}
func (noopHandler) CsiDispatch(string, [][]uint16, []byte, rune, bool) {}
func (noopHandler) EscDispatch([]byte, rune, bool)                     {}

// Parser represents a state machine.
type Parser struct {
	state           State
	intermediates   []byte
	intermediateIdx int
	params          *params
	param           uint16
	oscRaw          []byte
	oscParams       [][2]int
	oscNumParams    int
	ignoring        bool
	utf8Idx         int
	utf8Raw         [utf8.UTFMax]byte

	handler Handler

	// MaxParameters is the maximum number of parameters allowed.
	// Defaults to 32.
	MaxParameters int

	// MaxIntermediates is the maximum number of intermediates bytes allowed.
	// This is also used for private parameters such as (<=>?).
	// Defaults to 2.
	MaxIntermediates int

	// MaxOscBytes is the maximum number of bytes allowed in an Osc parameter.
	// Defaults to 1024.
	MaxOscBytes int

	// MaxOscParameters is the maximum number of Osc parameters allowed.
	// Defaults to 16.
	MaxOscParameters int
}

// New returns a new DEC ANSI compatible sequence parser.
func New(
	handler Handler,
) *Parser {
	if handler == nil {
		handler = noopHandler{}
	}
	p := &Parser{
		state:            GroundState,
		handler:          handler,
		MaxParameters:    DefaultMaxParameters,
		MaxIntermediates: DefaultMaxIntermediates,
		MaxOscBytes:      DefaultMaxOscBytes,
		MaxOscParameters: DefaultMaxOscParameters,
	}

	return p
}

func (p *Parser) init() {
	maxIntermediates := p.MaxIntermediates
	if maxIntermediates <= 0 {
		maxIntermediates = DefaultMaxIntermediates
	}
	maxOscBytes := p.MaxOscBytes
	if maxOscBytes <= 0 {
		maxOscBytes = DefaultMaxOscBytes
	}
	maxOscParams := p.MaxOscParameters
	if maxOscParams <= 0 {
		maxOscParams = DefaultMaxOscParameters
	}
	p.intermediates = make([]byte, maxIntermediates)
	p.oscRaw = make([]byte, 0, maxOscBytes)
	p.oscParams = make([][2]int, maxOscParams)
	p.params = newParams(p.MaxParameters)
}

// Parse parses the given reader until eof.
func (p *Parser) Parse(r io.Reader) error {
	p.init()
	buf := [1]byte{}
	for {
		_, err := r.Read(buf[:])
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		p.advance(buf[0])
	}
}

func (p *Parser) advanceUtf8(code byte) {
	// Collect the byte into the array
	p.collectUtf8(code)
	len := utf8ByteLen(p.utf8Raw[0])
	if len == -1 {
		// We panic here because the first byte comes from the state machine,
		// if this panics, it means there is a bug in the state machine!
		panic("invalid rune") // unreachable
	}

	if p.utf8Idx < len {
		return
	}

	// We have enough bytes to decode the rune
	bts := p.utf8Raw[:len]
	r, _ := utf8.DecodeRune(bts)
	p.handler.Print(r)
	p.state = GroundState
	p.clearUtf8()
}

// advance advances the state machine.
func (p *Parser) advance(code byte) {
	if p.state == Utf8State {
		p.advanceUtf8(code)
	} else {
		state, action := table.Transition(p.state, code)
		p.performStateChange(state, action, code)
	}
}

// getIntermediates returns a copy of the intermediates
func (p *Parser) getIntermediates() (string, []byte) {
	intr := append([]byte{}, p.intermediates[:p.intermediateIdx]...)
	prefix := getPrefix(intr)
	if prefix != "" {
		intr = intr[1:]
	}
	return prefix, intr
}

func (p *Parser) getOscParams() [][]byte {
	params := make([][]byte, 0, DefaultMaxOscParameters)

	for i := 0; i < p.oscNumParams; i++ {
		indices := p.oscParams[i]
		param := p.oscRaw[indices[0]:indices[1]]
		params = append(params, param)
	}

	return params[:p.oscNumParams]
}

// StateName returns the current state name
func (p *Parser) StateName() string {
	return stateNames[p.state]
}

// State returns the current state
func (p *Parser) State() State {
	return p.state
}

func (p *Parser) performStateChange(state State, action Action, code byte) {
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
	switch action {
	case NoneAction:
		break

	case IgnoreAction:
		break

	case PrintAction:
		p.handler.Print(rune(code))

	case ExecuteAction:
		p.handler.Execute(code)

	case DcsHookAction:
		if p.params.IsFull() {
			p.ignoring = true
		} else {
			p.params.Push(p.param)
		}

		prefix, intr := p.getIntermediates()
		p.handler.DcsHook(
			prefix,
			p.params.Params(),
			intr,
			rune(code),
			p.ignoring,
		)

	case DcsPutAction:
		p.handler.DcsPut(code)

	case OscStartAction:
		p.oscRaw = make([]byte, 0)
		p.oscNumParams = 0

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

		p.handler.OscDispatch(
			p.getOscParams(),
			code == 0x07,
		)

	case DcsUnhookAction:
		p.handler.DcsUnhook()

	case CsiDispatchAction:
		if p.params.IsFull() {
			p.ignoring = true
		} else {
			p.params.Push(p.param)
		}

		prefix, intr := p.getIntermediates()
		p.handler.CsiDispatch(
			prefix,
			p.params.Params(),
			intr,
			rune(code),
			p.ignoring,
		)

	case EscDispatchAction:
		_, intr := p.getIntermediates()
		p.handler.EscDispatch(
			intr,
			rune(code),
			p.ignoring,
		)

	case CollectAction:
		if utf8ByteLen(code) > 1 {
			p.collectUtf8(code)
		} else {
			p.collect(code)
		}

	case ParamAction:
		if p.params.IsFull() {
			p.ignoring = true
			return
		}

		if code == ';' {
			p.params.Push(p.param)
			p.param = 0
		} else if code == ':' {
			p.params.Extend(p.param)
			p.param = 0
		} else {
			p.param = smulu16(p.param, 10)
			p.param = saddu16(p.param, uint16((code - '0')))
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
	if p.intermediateIdx == DefaultMaxIntermediates {
		p.ignoring = true
	} else {
		p.intermediates[p.intermediateIdx] = code
		p.intermediateIdx++
	}
}

func (p *Parser) clearUtf8() {
	p.utf8Idx = 0
}

func (p *Parser) clear() {
	// Reset everything on ESC/CSI/DCS entry
	p.intermediateIdx = 0
	p.ignoring = false
	p.param = 0

	p.params.Clear()
}

func saddu16(a, b uint16) uint16 {
	if b > 0 && a > math.MaxUint16-b {
		return math.MaxUint16
	}

	return a + b
}

func smulu16(a, b uint16) uint16 {
	if a > 0 && b > 0 && a > math.MaxUint16/b {
		return math.MaxUint16
	}

	return a * b
}

// getPrefix extracts the prefix from the intermediates
// A prefix is the first byte of the intermediates
// and consists of a single byte between 0x3C and 0x3F
func getPrefix(intr []byte) string {
	if len(intr) == 0 {
		return ""
	}

	if intr[0] < 0x3C || intr[0] > 0x3F {
		return ""
	}

	return string(intr[0])
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
	} else {
		return -1
	}
}
