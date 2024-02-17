package ansi

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/input"
)

type UnknownCsiEvent string

var _ input.Event = UnknownCsiEvent("")

// String implements input.Event.
func (s UnknownCsiEvent) String() string {
	return fmt.Sprintf("Unknown CSI sequence: %q", string(s))
}

// Type implements input.Event.
func (s UnknownCsiEvent) Type() string {
	return "CSI"
}

type UnknownSs3Event string

var _ input.Event = UnknownSs3Event("")

// String implements input.Event.
func (s UnknownSs3Event) String() string {
	return fmt.Sprintf("Unknown SS3 sequence: %q", string(s))
}

// Type implements input.Event.
func (UnknownSs3Event) Type() string {
	return "SS3"
}

type UnknownOscEvent string

var _ input.Event = UnknownOscEvent("")

// String implements input.Event.
func (s UnknownOscEvent) String() string {
	return fmt.Sprintf("Unknown OSC sequence: %q", string(s))
}

// Type implements input.Event.
func (UnknownOscEvent) Type() string {
	return "OSC"
}
