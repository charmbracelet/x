package input

import (
	"fmt"

	"github.com/charmbracelet/x/exp/term/ansi"
)

// UnknownCsiEvent represents an unknown CSI sequence event.
type UnknownCsiEvent struct {
	ansi.CsiSequence
}

// String implements fmt.Stringer.
func (e UnknownCsiEvent) String() string {
	return fmt.Sprintf("%q", e.CsiSequence)
}

// UnknownOscEvent represents an unknown OSC sequence event.
type UnknownOscEvent struct {
	ansi.OscSequence
}

// String implements fmt.Stringer.
func (e UnknownOscEvent) String() string {
	return fmt.Sprintf("%q", e.OscSequence)
}

// UnknownDcsEvent represents an unknown DCS sequence event.
type UnknownDcsEvent string

// String implements fmt.Stringer.
func (e UnknownDcsEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// UnknownApcEvent represents an unknown APC sequence event.
type UnknownApcEvent string

// String implements fmt.Stringer.
func (e UnknownApcEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}
