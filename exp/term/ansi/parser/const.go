package parser

// Action is a DEC ANSI parser action.
type Action = byte

// These are the actions that the parser can take.
const (
	NoneAction Action = iota
	IgnoreAction
	ClearAction
	CollectAction
	CsiDispatchAction
	EscDispatchAction
	ExecuteAction
	DcsHookAction
	OscEndAction
	OscPutAction
	StartAction // Start an Osc or SosPmApc sequence
	ParamAction
	PrintAction
	PutAction // Used to put data into the buffer for Dcs/Sos/Pm/Apc
	DcsUnhookAction
	SosPmApcEndAction
)

// nolint: unused
var ActionNames = []string{
	"NoneAction",
	"IgnoreAction",
	"ClearAction",
	"CollectAction",
	"CsiDispatchAction",
	"EscDispatchAction",
	"ExecuteAction",
	"DcsHookAction",
	"OscEndAction",
	"OscPutAction",
	"StartAction",
	"ParamAction",
	"PrintAction",
	"PutAction",
	"DcsUnhookAction",
	"SosPmApcEndAction",
}

// State is a DEC ANSI parser state.
type State = byte

// These are the states that the parser can be in.
const (
	GroundState State = iota
	CsiEntryState
	CsiIgnoreState
	CsiIntermediateState
	CsiParamState
	DcsEntryState
	DcsIgnoreState
	DcsIntermediateState
	DcsParamState
	DcsPassthroughState
	EscapeState
	EscapeIntermediateState
	OscStringState
	SosPmApcStringState

	// Utf8State is not part of the DEC ANSI standard. It is used to handle
	// UTF-8 sequences.
	Utf8State
)

// nolint: unused
var StateNames = []string{
	"GroundState",
	"CsiEntryState",
	"CsiIgnoreState",
	"CsiIntermediateState",
	"CsiParamState",
	"DcsEntryState",
	"DcsIgnoreState",
	"DcsIntermediateState",
	"DcsParamState",
	"DcsPassthroughState",
	"EscapeState",
	"EscapeIntermediateState",
	"OscStringState",
	"SosPmApcStringState",
	"Utf8State",
}
