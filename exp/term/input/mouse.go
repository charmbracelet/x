package input

import (
	"fmt"
	"regexp"
	"strconv"
)

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

// String implements fmt.Stringer.
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

var mouseSGRRegex = regexp.MustCompile(`(\d+);(\d+);(\d+)([Mm])`)

// Parse SGR-encoded mouse events; SGR extended mouse events. SGR mouse events
// look like:
//
//	ESC [ < Cb ; Cx ; Cy (M or m)
//
// where:
//
//	Cb is the encoded button code
//	Cx is the x-coordinate of the mouse
//	Cy is the y-coordinate of the mouse
//	M is for button press, m is for button release
//
// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseSGRMouseEvent(buf []byte) MouseEvent {
	str := string(buf[3:])
	matches := mouseSGRRegex.FindStringSubmatch(str)
	if len(matches) != 5 {
		// Unreachable, we already checked the regex in `detectOneMsg`.
		panic("invalid mouse event")
	}

	b, _ := strconv.Atoi(matches[1])
	px := matches[2]
	py := matches[3]
	release := matches[4] == "m"
	m := parseMouseButton(b, true)

	// Wheel buttons don't have release events
	// Motion can be reported as a release event in some terminals (Windows Terminal)
	if m.Action != MouseActionMotion && !m.IsWheel() && release {
		m.Action = MouseActionRelease
	}

	x, _ := strconv.Atoi(px)
	y, _ := strconv.Atoi(py)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = x - 1
	m.Y = y - 1

	return m
}

const x10MouseByteOffset = 32

// Parse X10-encoded mouse events; the simplest kind. The last release of X10
// was December 1986, by the way. The original X10 mouse protocol limits the Cx
// and Cy coordinates to 223 (=255-032).
//
// X10 mouse events look like:
//
//	ESC [M Cb Cx Cy
//
// See: http://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
func parseX10MouseEvent(buf []byte) MouseEvent {
	v := buf[3:6]
	m := parseMouseButton(int(v[0]), false)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = int(v[1]) - x10MouseByteOffset - 1
	m.Y = int(v[2]) - x10MouseByteOffset - 1

	return m
}

// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseMouseButton(b int, isSGR bool) MouseEvent {
	var m MouseEvent
	e := b
	if !isSGR {
		e -= x10MouseByteOffset
	}

	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000
		bitAdd    = 0b1000_0000 // additional buttons 8-11

		bitsMask = 0b0000_0011
	)

	if e&bitAdd != 0 {
		m.Button = MouseButtonBackward + MouseButton(e&bitsMask)
	} else if e&bitWheel != 0 {
		m.Button = MouseButtonWheelUp + MouseButton(e&bitsMask)
	} else {
		m.Button = MouseButtonLeft + MouseButton(e&bitsMask)
		// X10 reports a button release as 0b0000_0011 (3)
		if e&bitsMask == bitsMask {
			m.Action = MouseActionRelease
			m.Button = MouseButtonNone
		}
	}

	// Motion bit doesn't get reported for wheel events.
	if e&bitMotion != 0 && !m.IsWheel() {
		m.Action = MouseActionMotion
	}

	// Modifiers
	if e&bitAlt != 0 {
		m.Mod |= Alt
	}
	if e&bitCtrl != 0 {
		m.Mod |= Ctrl
	}
	if e&bitShift != 0 {
		m.Mod |= Shift
	}

	return m
}
