//nolint:gosec
package windows

import (
	"encoding/binary"

	"golang.org/x/sys/windows"
)

// FocusEventRecord corresponds to the FocusEventRecord structure from the
// Windows console API.
// https://docs.microsoft.com/en-us/windows/console/focus-event-record-str
type FocusEventRecord struct {
	// SetFocus is reserved and should not be used.
	SetFocus bool
}

// MenuEventRecord corresponds to the MenuEventRecord structure from the
// Windows console API.
// https://docs.microsoft.com/en-us/windows/console/menu-event-record-str
type MenuEventRecord struct {
	CommandID uint32
}

// MouseEventRecord corresponds to the MouseEventRecord structure from the
// Windows console API.
// https://docs.microsoft.com/en-us/windows/console/mouse-event-record-str
type MouseEventRecord struct {
	// MousePosition contains the location of the cursor, in terms of the
	// console screen buffer's character-cell coordinates.
	MousePositon windows.Coord

	// ButtonState holds the status of the mouse buttons.
	ButtonState uint32

	// ControlKeyState holds the state of the control keys.
	ControlKeyState uint32

	// EventFlags specify the type of mouse event.
	EventFlags uint32
}

// WindowBufferSizeRecord corresponds to the WindowBufferSizeRecord structure
// from the Windows console API.
// https://docs.microsoft.com/en-us/windows/console/window-buffer-size-record-str
type WindowBufferSizeRecord struct {
	// Size contains the size of the console screen buffer, in character cell
	// columns and rows.
	Size windows.Coord
}

// FocusEvent returns the event as a FOCUS_EVENT_RECORD.
func (ir InputRecord) FocusEvent() FocusEventRecord {
	return FocusEventRecord{SetFocus: ir.Event[0] > 0}
}

// KeyEvent returns the event as a KEY_EVENT_RECORD.
func (ir InputRecord) KeyEvent() KeyEventRecord {
	return KeyEventRecord{
		KeyDown:         binary.LittleEndian.Uint32(ir.Event[0:4]) > 0,
		RepeatCount:     binary.LittleEndian.Uint16(ir.Event[4:6]),
		VirtualKeyCode:  binary.LittleEndian.Uint16(ir.Event[6:8]),
		VirtualScanCode: binary.LittleEndian.Uint16(ir.Event[8:10]),
		Char:            rune(binary.LittleEndian.Uint16(ir.Event[10:12])),
		ControlKeyState: binary.LittleEndian.Uint32(ir.Event[12:16]),
	}
}

// MouseEvent returns the event as a MOUSE_EVENT_RECORD.
func (ir InputRecord) MouseEvent() MouseEventRecord {
	return MouseEventRecord{
		MousePositon: windows.Coord{
			X: int16(binary.LittleEndian.Uint16(ir.Event[0:2])),
			Y: int16(binary.LittleEndian.Uint16(ir.Event[2:4])),
		},
		ButtonState:     binary.LittleEndian.Uint32(ir.Event[4:8]),
		ControlKeyState: binary.LittleEndian.Uint32(ir.Event[8:12]),
		EventFlags:      binary.LittleEndian.Uint32(ir.Event[12:16]),
	}
}

// WindowBufferSizeEvent returns the event as a WINDOW_BUFFER_SIZE_RECORD.
func (ir InputRecord) WindowBufferSizeEvent() WindowBufferSizeRecord {
	return WindowBufferSizeRecord{
		Size: windows.Coord{
			X: int16(binary.LittleEndian.Uint16(ir.Event[0:2])),
			Y: int16(binary.LittleEndian.Uint16(ir.Event[2:4])),
		},
	}
}

// MenuEvent returns the event as a MENU_EVENT_RECORD.
func (ir InputRecord) MenuEvent() MenuEventRecord {
	return MenuEventRecord{
		CommandID: binary.LittleEndian.Uint32(ir.Event[0:4]),
	}
}
