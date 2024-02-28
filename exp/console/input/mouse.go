package input

import "fmt"

// MouseButton represents the button that was pressed during a mouse event.
type MouseButton int

// Mouse event buttons
//
// This is based on X11 mouse button codes.
//
//	1 = left button
//	2 = middle button (pressing the scroll wheel)
//	3 = right button
//	4 = turn scroll wheel up
//	5 = turn scroll wheel down
//	6 = push scroll wheel left
//	7 = push scroll wheel right
//	8 = 4th button (aka browser backward button)
//	9 = 5th button (aka browser forward button)
//	10
//	11
//
// Other buttons are not supported.
const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonWheelUp
	MouseButtonWheelDown
	MouseButtonWheelLeft
	MouseButtonWheelRight
	MouseButtonBackward
	MouseButtonForward
	MouseButton10
	MouseButton11
)

var mouseButtons = map[MouseButton]string{
	MouseButtonNone:       "none",
	MouseButtonLeft:       "left",
	MouseButtonMiddle:     "middle",
	MouseButtonRight:      "right",
	MouseButtonWheelUp:    "wheel up",
	MouseButtonWheelDown:  "wheel down",
	MouseButtonWheelLeft:  "wheel left",
	MouseButtonWheelRight: "wheel right",
	MouseButtonBackward:   "backward",
	MouseButtonForward:    "forward",
	MouseButton10:         "button 10",
	MouseButton11:         "button 11",
}

// MouseAction represents the action that occurred during a mouse event.
type MouseAction int

// Mouse event actions.
const (
	MouseActionPress MouseAction = iota
	MouseActionRelease
	MouseActionMotion
)

var mouseActions = map[MouseAction]string{
	MouseActionPress:   "press",
	MouseActionRelease: "release",
	MouseActionMotion:  "motion",
}

// Mouse represents a mouse event.
type MouseEvent struct {
	X, Y   int
	Mod    Mod
	Button MouseButton
	Action MouseAction
}

// IsWheel returns true if the mouse event is a wheel event.
func (m MouseEvent) IsWheel() bool {
	return m.Button >= MouseButtonWheelUp && m.Button <= MouseButtonWheelRight
}

var _ Event = MouseEvent{}

// String implements Event.
func (m MouseEvent) String() (s string) {
	if m.Mod.IsCtrl() {
		s += "ctrl+"
	}
	if m.Mod.IsAlt() {
		s += "alt+"
	}
	if m.Mod.IsShift() {
		s += "shift+"
	}

	if m.Button == MouseButtonNone {
		if m.Action == MouseActionMotion || m.Action == MouseActionRelease {
			s += mouseActions[m.Action]
		} else {
			s += "unknown"
		}
	} else if m.IsWheel() {
		s += mouseButtons[m.Button]
	} else {
		btn := mouseButtons[m.Button]
		if btn != "" {
			s += btn
		}
		act := mouseActions[m.Action]
		if act != "" {
			s += " " + act
		}
	}
	s += fmt.Sprintf(" at (%d, %d)", m.X, m.Y)

	return s
}

// Type implements Event.
func (MouseEvent) Type() string {
	return "Mouse"
}
