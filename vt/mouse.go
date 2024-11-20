package vt

import (
	"github.com/charmbracelet/x/ansi"
)

// MouseButton represents the button that was pressed during a mouse message.
type MouseButton byte

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
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
	MouseWheelLeft
	MouseWheelRight
	MouseBackward
	MouseForward
	MouseExtra1
	MouseExtra2
)

// Mouse represents a mouse event.
type Mouse interface {
	Mouse() mouse
}

// mouse represents a mouse message. Use [Mouse] to represent all mouse
// messages.
//
// The X and Y coordinates are zero-based, with (0,0) being the upper left
// corner of the terminal.
type mouse struct {
	X, Y   int
	Button MouseButton
	Mod    KeyMod
}

// MouseClick represents a mouse click event.
type MouseClick mouse

// Mouse returns the mouse event.
func (m MouseClick) Mouse() mouse {
	return mouse(m)
}

// MouseRelease represents a mouse release event.
type MouseRelease mouse

// Mouse returns the mouse event.
func (m MouseRelease) Mouse() mouse {
	return mouse(m)
}

// MouseWheel represents a mouse wheel event.
type MouseWheel mouse

// Mouse returns the mouse event.
func (m MouseWheel) Mouse() mouse {
	return mouse(m)
}

// MouseMotion represents a mouse motion event.
type MouseMotion mouse

// Mouse returns the mouse event.
func (m MouseMotion) Mouse() mouse {
	return mouse(m)
}

// SendMouse sends a mouse event to the terminal.
// TODO: Support [Utf8ExtMouseMode], [UrxvtExtMouseMode], and
// [SgrPixelExtMouseMode].
func (t *Terminal) SendMouse(m Mouse) {
	var (
		enc  ansi.Mode
		mode ansi.Mode
	)

	for _, m := range []ansi.DECMode{
		ansi.X10MouseMode,         // Button press
		ansi.NormalMouseMode,      // Button press/release
		ansi.HighlightMouseMode,   // Button press/release/hilight
		ansi.ButtonEventMouseMode, // Button press/release/cell motion
		ansi.AnyEventMouseMode,    // Button press/release/all motion
	} {
		if t.isModeSet(m) {
			mode = m
		}
	}

	if mode == nil {
		return
	}

	for _, e := range []ansi.DECMode{
		// ansi.Utf8ExtMouseMode,
		ansi.SgrExtMouseMode,
		// ansi.UrxvtExtMouseMode,
		// ansi.SgrPixelExtMouseMode,
	} {
		if t.isModeSet(e) {
			enc = e
		}
	}

	// mouse bit shifts
	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000
		bitAdd    = 0b1000_0000 // additional buttons 8-11

		bitsMask = 0b0000_0011
	)

	var b byte
	var release bool
	if _, ok := m.(MouseRelease); ok {
		release = true
	}

	// Encode button
	mouse := m.Mouse()
	if release && enc == nil {
		// X10 mouse encoding reports release as a b == 3
		b = bitsMask
	} else if mouse.Button >= MouseLeft && mouse.Button <= MouseRight {
		b = byte(mouse.Button) - byte(MouseLeft)
	} else if mouse.Button >= MouseWheelUp && mouse.Button <= MouseWheelRight {
		b = byte(mouse.Button) - byte(MouseWheelUp)
		b |= bitWheel
	} else if mouse.Button >= MouseBackward && mouse.Button <= MouseExtra2 {
		b = byte(mouse.Button) - byte(MouseBackward)
		b |= bitAdd
	}

	switch m.(type) {
	case MouseMotion:
		switch {
		case mouse.Button == MouseNone && mode == ansi.AnyEventMouseMode:
			b = bitsMask
			fallthrough
		case mouse.Button > MouseNone && mode == ansi.ButtonEventMouseMode:
			b |= bitMotion
		default:
			// No motion events
			return
		}
	}

	// Encode modifiers
	if mouse.Mod&ModShift != 0 {
		b |= bitShift
	}
	if mouse.Mod&ModAlt != 0 {
		b |= bitAlt
	}
	if mouse.Mod&ModCtrl != 0 {
		b |= bitCtrl
	}

	switch enc {
	// TODO: Support [ansi.HighlightMouseMode].
	// TODO: Support [ansi.Utf8ExtMouseMode], [ansi.UrxvtExtMouseMode], and
	// [ansi.SgrPixelExtMouseMode].
	case nil: // X10 mouse encoding
		t.buf.WriteString(ansi.MouseX10(b, mouse.X, mouse.Y))
	case ansi.SgrExtMouseMode: // SGR mouse encoding
		t.buf.WriteString(ansi.MouseSgr(b, mouse.X, mouse.Y, release))
	}
}
