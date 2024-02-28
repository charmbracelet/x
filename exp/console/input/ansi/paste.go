package ansi

import (
	"github.com/charmbracelet/x/exp/console/input"
)

// PasteStartEvent is an event that is emitted when a terminal enters
// bracketed-paste mode.
type PasteStartEvent struct{}

var _ input.Event = PasteStartEvent{}

// String implements input.Event.
func (e PasteStartEvent) String() string {
	return "paste start"
}

// Type implements input.Event.
func (PasteStartEvent) Type() string {
	return "PasteStart"
}

// PasteEvent is an event that is emitted when a terminal receives pasted text.
type PasteEndEvent struct{}

var _ input.Event = PasteEndEvent{}

// String implements input.Event.
func (e PasteEndEvent) String() string {
	return "paste end"
}

// Type implements input.Event.
func (PasteEndEvent) Type() string {
	return "PasteEnd"
}
