package ansi

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/input"
)

type csiSequence string

var _ input.Event = csiSequence("")

// String implements input.Event.
func (s csiSequence) String() string {
	return fmt.Sprintf("CSI sequence: %q", string(s))
}

// Type implements input.Event.
func (s csiSequence) Type() string {
	return "CSI"
}

type ss3Sequence string

var _ input.Event = ss3Sequence("")

// String implements input.Event.
func (s ss3Sequence) String() string {
	return fmt.Sprintf("SS3 sequence: %q", string(s))
}

// Type implements input.Event.
func (ss3Sequence) Type() string {
	return "SS3"
}

type oscSequence string

var _ input.Event = oscSequence("")

// String implements input.Event.
func (s oscSequence) String() string {
	return fmt.Sprintf("OSC sequence: %q", string(s))
}

// Type implements input.Event.
func (oscSequence) Type() string {
	return "OSC"
}
