package ansi

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/console/input"
)

// UnknownCsiEvent represents an unknown CSI sequence event.
type UnknownCsiEvent struct {
	ansi.CsiSequence
}

var _ input.Event = UnknownCsiEvent{}

// String implements input.Event.
func (e UnknownCsiEvent) String() string {
	return fmt.Sprintf("unknown CSI sequence: %q", e.CsiSequence)
}

// Type implements input.Event.
func (UnknownCsiEvent) Type() string {
	return "Unknown CSI Sequence"
}

// UnknownOscEvent represents an unknown OSC sequence event.
type UnknownOscEvent struct {
	ansi.OscSequence
}

var _ input.Event = UnknownOscEvent{}

// String implements input.Event.
func (e UnknownOscEvent) String() string {
	return fmt.Sprintf("unknown OSC sequence: %q", e.OscSequence)
}

// Type implements input.Event.
func (UnknownOscEvent) Type() string {
	return "Unknown OSC Sequence"
}
