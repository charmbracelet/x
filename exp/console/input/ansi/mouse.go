package ansi

import (
	"regexp"
	"strconv"

	"github.com/charmbracelet/x/exp/console/input"
)

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
func parseSGRMouseEvent(buf []byte) input.MouseEvent {
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
	if m.Action != input.MouseActionMotion && !m.IsWheel() && release {
		m.Action = input.MouseActionRelease
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
func parseX10MouseEvent(buf []byte) input.MouseEvent {
	v := buf[3:6]
	m := parseMouseButton(int(v[0]), false)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = int(v[1]) - x10MouseByteOffset - 1
	m.Y = int(v[2]) - x10MouseByteOffset - 1

	return m
}

// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseMouseButton(b int, isSGR bool) input.MouseEvent {
	var m input.MouseEvent
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
		m.Button = input.MouseButtonBackward + input.MouseButton(e&bitsMask)
	} else if e&bitWheel != 0 {
		m.Button = input.MouseButtonWheelUp + input.MouseButton(e&bitsMask)
	} else {
		m.Button = input.MouseButtonLeft + input.MouseButton(e&bitsMask)
		// X10 reports a button release as 0b0000_0011 (3)
		if e&bitsMask == bitsMask {
			m.Action = input.MouseActionRelease
			m.Button = input.MouseButtonNone
		}
	}

	// Motion bit doesn't get reported for wheel events.
	if e&bitMotion != 0 && !m.IsWheel() {
		m.Action = input.MouseActionMotion
	}

	// Modifiers
	if e&bitAlt != 0 {
		m.Mod |= input.Alt
	}
	if e&bitCtrl != 0 {
		m.Mod |= input.Ctrl
	}
	if e&bitShift != 0 {
		m.Mod |= input.Shift
	}

	return m
}
