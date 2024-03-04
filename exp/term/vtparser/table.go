package parser

// Table values are generated like this:
//
//	index:  currentState << indexStateShift | charCode
//	value:  action << transitionActionShift | nextState
const (
	transitionActionShift = 4
	transitionStateMask   = 15
	indexStateShift       = 8
)

// TransitionTable is a DEC ANSI transition table.
// https://vt100.net/emu/dec_ansi_parser
type TransitionTable []byte

// NewTransitionTable returns a new DEC ANSI transition table.
func NewTransitionTable(size int) TransitionTable {
	if size <= 0 {
		size = 4096
	}
	return TransitionTable(make([]byte, size))
}

// SetDefault sets default transition.
func (t TransitionTable) SetDefault(action Action, state State) {
	for i := 0; i < len(t); i++ {
		t[i] = action<<transitionActionShift | state
	}
}

// AddOne adds a transition.
func (t TransitionTable) AddOne(code byte, state State, action Action, next State) {
	idx := int(state)<<indexStateShift | int(code)
	value := action<<transitionActionShift | next
	t[idx] = value

}

// AddMany adds many transitions.
func (t TransitionTable) AddMany(codes []byte, state State, action Action, next State) {
	for _, code := range codes {
		t.AddOne(code, state, action, next)
	}
}

// AddRange adds a range of transitions.
func (t TransitionTable) AddRange(start, end byte, state State, action Action, next State) {
	for i := int(start); i <= int(end); i++ {
		t.AddOne(byte(i), state, action, next)
	}
}

// Transition returns the next state and action for the given state and byte.
func (t TransitionTable) Transition(state State, code byte) (State, Action) {
	index := int(state)<<indexStateShift | int(code)
	value := t[index]
	return State(value & transitionStateMask), Action(value >> transitionActionShift)
}

// byte range macro
func r(start, end byte) []byte {
	var a []byte
	for i := int(start); i <= int(end); i++ {
		a = append(a, byte(i))
	}
	return a
}

var table = GenerateTransitionTable()

// GenerateTransitionTable generates a DEC ANSI transition table compatible
// with the VT500-series of terminals. This implementation includes a few
// modifications that include:
//   - A new Utf8State is introduced to handle UTF8 sequences.
//   - Osc and Dcs data accept UTF8 sequences by extending the printable range
//     to 0xFF and 0xFE respectively.
//   - We don't ignore 0x3A (':') when building Csi and Dcs parameters and
//     instead use it to denote sub-parameters.
//   - TODO: implement APC
func GenerateTransitionTable() TransitionTable {
	table := NewTransitionTable(4096)
	table.SetDefault(NoneAction, GroundState)

	// Anywhere
	for _, state := range r(GroundState, Utf8State) { // TODO: adjust range
		// Anywhere -> Ground
		table.AddMany([]byte{0x18, 0x1a, 0x99, 0x9a}, state, ExecuteAction, GroundState)
		table.AddRange(0x80, 0x8F, state, ExecuteAction, GroundState)
		table.AddRange(0x90, 0x97, state, ExecuteAction, GroundState)
		table.AddOne(0x9C, state, IgnoreAction, GroundState)
		// Anywhere -> Escape
		table.AddOne(0x1B, state, ClearAction, EscapeState)
		// Anywhere -> SosPmApcStringState
		table.AddMany([]byte{0x98, 0x9E, 0x9F}, state, IgnoreAction, SosPmApcStringState)
		// Anywhere -> CsiEntry
		table.AddOne(0x9B, state, NoneAction, CsiEntryState)
		// Anywhere -> DcsEntry
		table.AddOne(0x90, state, NoneAction, DcsEntryState)
		// Anywhere -> OscString
		table.AddOne(0x9D, state, NoneAction, OscStringState)
		// Anywhere -> Utf8
		table.AddRange(0xC2, 0xDF, state, CollectAction, Utf8State) // UTF8 2 byte sequence
		table.AddRange(0xE0, 0xEF, state, CollectAction, Utf8State) // UTF8 3 byte sequence
		table.AddRange(0xF0, 0xF4, state, CollectAction, Utf8State) // UTF8 4 byte sequence
	}

	// Ground
	table.AddRange(0x00, 0x17, GroundState, ExecuteAction, GroundState)
	table.AddOne(0x19, GroundState, ExecuteAction, GroundState)
	table.AddRange(0x1C, 0x1F, GroundState, ExecuteAction, GroundState)
	table.AddRange(0x20, 0x7F, GroundState, PrintAction, GroundState)

	// EscapeIntermediate
	table.AddRange(0x00, 0x17, EscapeIntermediateState, ExecuteAction, EscapeIntermediateState)
	table.AddOne(0x19, EscapeIntermediateState, ExecuteAction, EscapeIntermediateState)
	table.AddRange(0x1C, 0x1F, EscapeIntermediateState, ExecuteAction, EscapeIntermediateState)
	table.AddRange(0x20, 0x2F, EscapeIntermediateState, CollectAction, EscapeIntermediateState)
	table.AddOne(0x7F, EscapeIntermediateState, IgnoreAction, EscapeIntermediateState)
	// EscapeIntermediate -> Ground
	table.AddRange(0x30, 0x7E, EscapeIntermediateState, EscDispatchAction, GroundState)

	// Sos_pm_apc_string
	table.AddRange(0x00, 0x17, SosPmApcStringState, IgnoreAction, SosPmApcStringState)
	table.AddOne(0x19, SosPmApcStringState, IgnoreAction, SosPmApcStringState)
	table.AddRange(0x1C, 0x1F, SosPmApcStringState, IgnoreAction, SosPmApcStringState)
	table.AddRange(0x20, 0x7F, SosPmApcStringState, IgnoreAction, SosPmApcStringState)

	// Escape
	table.AddRange(0x00, 0x17, EscapeState, ExecuteAction, EscapeState)
	table.AddOne(0x19, EscapeState, ExecuteAction, EscapeState)
	table.AddRange(0x1C, 0x1F, EscapeState, ExecuteAction, EscapeState)
	table.AddOne(0x7F, EscapeState, IgnoreAction, EscapeState)
	// Escape -> Ground
	table.AddRange(0x30, 0x4F, EscapeState, EscDispatchAction, GroundState)
	table.AddRange(0x51, 0x57, EscapeState, EscDispatchAction, GroundState)
	table.AddOne(0x59, EscapeState, EscDispatchAction, GroundState)
	table.AddOne(0x5A, EscapeState, EscDispatchAction, GroundState)
	table.AddOne(0x5C, EscapeState, EscDispatchAction, GroundState)
	table.AddRange(0x60, 0x7E, EscapeState, EscDispatchAction, GroundState)
	// Escape -> Escape_intermediate
	table.AddRange(0x20, 0x2F, EscapeState, CollectAction, EscapeIntermediateState)
	// Escape -> Sos_pm_apc_string
	table.AddOne(0x58, EscapeState, NoneAction, SosPmApcStringState)
	table.AddOne(0x5E, EscapeState, NoneAction, SosPmApcStringState)
	table.AddOne(0x5F, EscapeState, NoneAction, SosPmApcStringState)
	// Escape -> Dcs_entry
	table.AddOne(0x50, EscapeState, NoneAction, DcsEntryState)
	// Escape -> Csi_entry
	table.AddOne(0x5B, EscapeState, NoneAction, CsiEntryState)
	// Escape -> Osc_string
	table.AddOne(0x5D, EscapeState, NoneAction, OscStringState)

	// Dcs_entry
	table.AddRange(0x00, 0x17, DcsEntryState, IgnoreAction, DcsEntryState)
	table.AddOne(0x19, DcsEntryState, IgnoreAction, DcsEntryState)
	table.AddRange(0x1C, 0x1F, DcsEntryState, IgnoreAction, DcsEntryState)
	table.AddOne(0x7F, DcsEntryState, IgnoreAction, DcsEntryState)
	// Dcs_entry -> Dcs_intermediate
	table.AddRange(0x20, 0x2F, DcsEntryState, CollectAction, DcsIntermediateState)
	// Dcs_entry -> Dcs_ignore
	// Dcs_entry -> Dcs_param
	table.AddRange(0x30, 0x3B, DcsEntryState, ParamAction, DcsParamState)
	table.AddRange(0x3C, 0x3F, DcsEntryState, CollectAction, DcsParamState)
	// Dcs_entry -> Dcs_passthrough
	table.AddRange(0x40, 0x7E, DcsEntryState, NoneAction, DcsPassthroughState)

	// Dcs_intermediate
	table.AddRange(0x00, 0x17, DcsIntermediateState, IgnoreAction, DcsIntermediateState)
	table.AddOne(0x19, DcsIntermediateState, IgnoreAction, DcsIntermediateState)
	table.AddRange(0x1C, 0x1F, DcsIntermediateState, IgnoreAction, DcsIntermediateState)
	table.AddRange(0x20, 0x2F, DcsIntermediateState, CollectAction, DcsIntermediateState)
	table.AddOne(0x7F, DcsIntermediateState, IgnoreAction, DcsIntermediateState)
	// Dcs_intermediate -> Dcs_ignore
	table.AddRange(0x30, 0x3F, DcsIntermediateState, NoneAction, DcsIgnoreState)
	// Dcs_intermediate -> Dcs_passthrough
	table.AddRange(0x40, 0x7E, DcsIntermediateState, NoneAction, DcsPassthroughState)

	// Dcs_ignore
	table.AddRange(0x00, 0x17, DcsIgnoreState, IgnoreAction, DcsIgnoreState)
	table.AddOne(0x19, DcsIgnoreState, IgnoreAction, DcsIgnoreState)
	table.AddRange(0x1C, 0x1F, DcsIgnoreState, IgnoreAction, DcsIgnoreState)

	// Dcs_param
	table.AddRange(0x00, 0x17, DcsParamState, IgnoreAction, DcsParamState)
	table.AddOne(0x19, DcsParamState, IgnoreAction, DcsParamState)
	table.AddRange(0x1C, 0x1F, DcsParamState, IgnoreAction, DcsParamState)
	table.AddRange(0x30, 0x3B, DcsParamState, ParamAction, DcsParamState)
	table.AddOne(0x7F, DcsParamState, IgnoreAction, DcsParamState)
	// Dcs_param -> Dcs_ignore
	table.AddRange(0x3C, 0x3F, DcsParamState, NoneAction, DcsIgnoreState)
	// Dcs_param -> Dcs_intermediate
	table.AddRange(0x20, 0x2F, DcsParamState, CollectAction, DcsIntermediateState)
	// Dcs_param -> Dcs_passthrough
	table.AddRange(0x40, 0x7E, DcsParamState, NoneAction, DcsPassthroughState)

	// Dcs_passthrough
	table.AddRange(0x00, 0x17, DcsPassthroughState, DcsPutAction, DcsPassthroughState)
	table.AddOne(0x19, DcsPassthroughState, DcsPutAction, DcsPassthroughState)
	table.AddRange(0x1C, 0x1F, DcsPassthroughState, DcsPutAction, DcsPassthroughState)
	table.AddRange(0x20, 0x7E, DcsPassthroughState, DcsPutAction, DcsPassthroughState)
	table.AddOne(0x7F, DcsPassthroughState, IgnoreAction, DcsPassthroughState)
	table.AddRange(0x80, 0xFF, DcsPassthroughState, DcsPutAction, DcsPassthroughState) // Allow Utf8 characters by extending the printable range from 0x7F to 0xFF
	// ST, CAN, SUB, and ESC terminate the sequence
	// table.AddOne(0x1b, DcsPassthroughState, NoneAction, EscapeState)
	table.AddMany([]byte{0x9C, 0x18, 0x1A}, DcsPassthroughState, NoneAction, GroundState)

	// Csi_param
	table.AddRange(0x00, 0x17, CsiParamState, ExecuteAction, CsiParamState)
	table.AddOne(0x19, CsiParamState, ExecuteAction, CsiParamState)
	table.AddRange(0x1C, 0x1F, CsiParamState, ExecuteAction, CsiParamState)
	table.AddRange(0x30, 0x3B, CsiParamState, ParamAction, CsiParamState)
	table.AddOne(0x7F, CsiParamState, IgnoreAction, CsiParamState)
	// Csi_param -> Ground
	table.AddRange(0x40, 0x7E, CsiParamState, CsiDispatchAction, GroundState)
	// Csi_param -> Csi_ignore
	table.AddRange(0x3C, 0x3F, CsiParamState, IgnoreAction, CsiIgnoreState)
	// Csi_param -> Csi_intermediate
	table.AddRange(0x20, 0x2F, CsiParamState, CollectAction, CsiIntermediateState)

	// Csi_ignore
	table.AddRange(0x00, 0x17, CsiIgnoreState, ExecuteAction, CsiIgnoreState)
	table.AddOne(0x19, CsiIgnoreState, ExecuteAction, CsiIgnoreState)
	table.AddRange(0x1C, 0x1F, CsiIgnoreState, ExecuteAction, CsiIgnoreState)
	table.AddRange(0x20, 0x3F, CsiIgnoreState, IgnoreAction, CsiIgnoreState)
	table.AddOne(0x7F, CsiIgnoreState, IgnoreAction, CsiIgnoreState)
	// Csi_ignore -> Ground
	table.AddRange(0x40, 0x7E, CsiIgnoreState, NoneAction, GroundState)

	// Csi_intermediate
	table.AddRange(0x00, 0x17, CsiIntermediateState, ExecuteAction, CsiIntermediateState)
	table.AddOne(0x19, CsiIntermediateState, ExecuteAction, CsiIntermediateState)
	table.AddRange(0x1C, 0x1F, CsiIntermediateState, ExecuteAction, CsiIntermediateState)
	table.AddRange(0x20, 0x2F, CsiIntermediateState, CollectAction, CsiIntermediateState)
	table.AddOne(0x7F, CsiIntermediateState, IgnoreAction, CsiIntermediateState)
	// Csi_intermediate -> Ground
	table.AddRange(0x40, 0x7E, CsiIntermediateState, CsiDispatchAction, GroundState)
	// Csi_intermediate -> Csi_ignore
	table.AddRange(0x30, 0x3F, CsiIntermediateState, NoneAction, CsiIgnoreState)

	// Csi_entry
	table.AddRange(0x00, 0x17, CsiEntryState, ExecuteAction, CsiEntryState)
	table.AddOne(0x19, CsiEntryState, ExecuteAction, CsiEntryState)
	table.AddRange(0x1C, 0x1F, CsiEntryState, ExecuteAction, CsiEntryState)
	table.AddOne(0x7F, CsiEntryState, IgnoreAction, CsiEntryState)
	// Csi_entry -> Ground
	table.AddRange(0x40, 0x7E, CsiEntryState, CsiDispatchAction, GroundState)
	// Csi_entry -> Csi_intermediate
	table.AddRange(0x20, 0x2F, CsiEntryState, CollectAction, CsiIntermediateState)
	// Csi_entry -> Csi_param
	table.AddRange(0x30, 0x3B, CsiEntryState, ParamAction, CsiParamState)
	table.AddRange(0x3C, 0x3F, CsiEntryState, CollectAction, CsiParamState)

	// Osc_string
	table.AddRange(0x00, 0x06, OscStringState, IgnoreAction, OscStringState)
	table.AddRange(0x08, 0x17, OscStringState, IgnoreAction, OscStringState)
	table.AddOne(0x19, OscStringState, IgnoreAction, OscStringState)
	table.AddRange(0x1C, 0x1F, OscStringState, IgnoreAction, OscStringState)
	table.AddRange(0x20, 0xFF, OscStringState, OscPutAction, OscStringState) // Allow Utf8 characters by extending the printable range from 0x7F to 0xFF

	// ST, CAN, SUB, ESC, and BEL terminate the sequence
	table.AddOne(0x1b, OscStringState, NoneAction, EscapeState)
	table.AddMany([]byte{0x9c, 0x18, 0x1a, 0x07}, OscStringState, NoneAction, GroundState)

	return table
}
