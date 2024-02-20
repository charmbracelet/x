package input

import (
	"fmt"
)

var (
	// ErrUnknownEvent is returned when an unknown event is encountered.
	ErrUnknownEvent = fmt.Errorf("unknown event")

	// ErrEmpty is returned when the event buffer is empty.
	ErrEmpty = fmt.Errorf("empty event buffer")
)

// Event represents a terminal input event.
type Event interface {
	fmt.Stringer

	// Type returns the type of the event.
	Type() string
}

// Driver represents a terminal input driver.
type Driver interface {
	// ReadInput reads input events from the terminal.
	ReadInput() ([]Event, error)

	// PeekInput peeks at input events from the terminal without consuming
	// them.
	PeekInput() ([]Event, error)
}

// UnknownEvent represents an unknown event.
type UnknownEvent string

var _ Event = UnknownEvent("")

// String implements Event.
func (e UnknownEvent) String() string {
	return fmt.Sprintf("unknown event: %q", string(e))
}

// Type implements Event.
func (UnknownEvent) Type() string {
	return "Unknown"
}
