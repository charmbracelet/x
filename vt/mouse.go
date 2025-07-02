package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
)

// MouseButton represents the button that was pressed during a mouse message.
type MouseButton = input.MouseButton

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
	MouseNone       = input.MouseNone
	MouseLeft       = input.MouseLeft
	MouseMiddle     = input.MouseMiddle
	MouseRight      = input.MouseRight
	MouseWheelUp    = input.MouseWheelUp
	MouseWheelDown  = input.MouseWheelDown
	MouseWheelLeft  = input.MouseWheelLeft
	MouseWheelRight = input.MouseWheelRight
	MouseBackward   = input.MouseBackward
	MouseForward    = input.MouseForward
	MouseButton10   = input.MouseButton10
	MouseButton11   = input.MouseButton11
)

// Mouse represents a mouse event.
type Mouse = input.MouseEvent

// MouseClick represents a mouse click event.
type MouseClick = input.MouseClickEvent

// MouseRelease represents a mouse release event.
type MouseRelease = input.MouseReleaseEvent

// MouseWheel represents a mouse wheel event.
type MouseWheel = input.MouseWheelEvent

// MouseMotion represents a mouse motion event.
type MouseMotion = input.MouseMotionEvent

// SendMouse sends a mouse event to the terminal. This can be any kind of mouse
// events such as [MouseClick], [MouseRelease], [MouseWheel], or [MouseMotion].
func (t *Terminal) SendMouse(m Mouse) {
	// XXX: Support [Utf8ExtMouseMode], [UrxvtExtMouseMode], and
	// [SgrPixelExtMouseMode].
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

	// Encode button
	mouse := m.Mouse()
	_, isMotion := m.(MouseMotion)
	_, isRelease := m.(MouseRelease)
	b := ansi.EncodeMouseButton(mouse.Button, isMotion,
		mouse.Mod.Contains(ModShift),
		mouse.Mod.Contains(ModAlt),
		mouse.Mod.Contains(ModCtrl))

	switch enc {
	// XXX: Support [ansi.HighlightMouseMode].
	// XXX: Support [ansi.Utf8ExtMouseMode], [ansi.UrxvtExtMouseMode], and
	// [ansi.SgrPixelExtMouseMode].
	case nil: // X10 mouse encoding
		t.buf.WriteString(ansi.MouseX10(b, mouse.X, mouse.Y))
	case ansi.SgrExtMouseMode: // SGR mouse encoding
		t.buf.WriteString(ansi.MouseSgr(b, mouse.X, mouse.Y, isRelease))
	}
}
