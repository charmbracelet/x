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
type UnknownEvent struct {
	Any any
}

var _ Event = UnknownEvent{}

// String implements Event.
func (e UnknownEvent) String() string {
	var s string
	switch v := e.Any.(type) {
	case string:
		s = v
	case fmt.Stringer:
		s = v.String()
	default:
		s = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("unknown event: %q", s)
}

// Type implements Event.
func (UnknownEvent) Type() string {
	return "Unknown"
}
