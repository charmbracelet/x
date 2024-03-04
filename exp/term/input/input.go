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
type Event interface{}

// UnknownEvent represents an unknown event.
type UnknownEvent string

// String implements fmt.Stringer.
func (e UnknownEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// WindowSizeEvent represents a window resize event.
type WindowSizeEvent struct {
	Width, Height int
}

// String implements fmt.Stringer.
func (e WindowSizeEvent) String() string {
	return fmt.Sprintf("resize: %dx%d", e.Width, e.Height)
}
