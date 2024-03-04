package parser

// Action is a DEC ANSI parser action.
type Action = byte

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
	OscStartAction
	ParamAction
	PrintAction
	DcsPutAction
	DcsUnhookAction
)

// nolint: unused
var actionNames = []string{
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
	"OscStartAction",
	"ParamAction",
	"PrintAction",
	"DcsPutAction",
	"DcsUnhookAction",
}

// State is a DEC ANSI parser state.
type State = byte

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

var stateNames = []string{
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
