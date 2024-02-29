package input

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/ansi"
)

// UnknownCsiEvent represents an unknown CSI sequence event.
type UnknownCsiEvent struct {
	ansi.CsiSequence
}

var _ Event = UnknownCsiEvent{}

// String implements Event.
func (e UnknownCsiEvent) String() string {
	return fmt.Sprintf("unknown CSI sequence: %q", e.CsiSequence)
}

// Type implements Event.
func (UnknownCsiEvent) Type() string {
	return "Unknown CSI Sequence"
}

// UnknownOscEvent represents an unknown OSC sequence event.
type UnknownOscEvent struct {
	ansi.OscSequence
}

var _ Event = UnknownOscEvent{}

// String implements Event.
func (e UnknownOscEvent) String() string {
	return fmt.Sprintf("unknown OSC sequence: %q", e.OscSequence)
}

// Type implements Event.
func (UnknownOscEvent) Type() string {
	return "Unknown OSC Sequence"
}
