package input

import "fmt"

// Button represents a mouse button.
type Button int

// Mouse represents a mouse event.
type Mouse struct {
	X, Y int    // position
	Btn  Button // button
}

// MouseEvent represents a mouse event.
type MouseEvent Mouse

var _ Event = MouseEvent{}

// String implements Event.
func (e MouseEvent) String() string {
	return fmt.Sprintf("mouse: x=%d, y=%d, btn=%d", e.X, e.Y, e.Btn)
}

// Type implements Event.
func (MouseEvent) Type() string {
	return "Mouse"
}
