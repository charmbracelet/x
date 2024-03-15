//go:build windows
// +build windows

package input

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/erikgeiser/coninput"
	"golang.org/x/sys/windows"
)

// ReadInput reads input events from the terminal.
//
// It reads up to len(e) events into e and returns the number of events read
// and an error, if any.
func (d *Driver) ReadInput(e []Event) (n int, err error) {
	events, err := d.handleConInput(coninput.ReadConsoleInput)
	if errors.Is(err, errNotConInputReader) {
		return d.readInput(e)
	}
	if err != nil {
		return 0, err
	}

	ne := copy(e, events)
	return ne, nil
}

var errNotConInputReader = fmt.Errorf("handleConInput: not a conInputReader")

// PeekInput peeks at input events from the terminal without consuming them.
//
// If the number of events requested is greater than the number of events
// available in the buffer, the number of available events will be returned.
func (d *Driver) PeekInput(n int) ([]Event, error) {
	events, err := d.handleConInput(coninput.PeekConsoleInput)
	if errors.Is(err, errNotConInputReader) {
		return d.peekInput(n)
	}
	if err != nil {
		return nil, err
	}

	if n < len(events) {
		return events[:n], nil
	}

	return events, nil
}

func (d *Driver) handleConInput(
	finput func(windows.Handle, []coninput.InputRecord) (uint32, error),
) ([]Event, error) {
	cc, ok := d.rd.(*conInputReader)
	if !ok {
		return nil, errNotConInputReader
	}

	// read up to 256 events, this is to allow for sequences events reported as
	// key events.
	var events [256]coninput.InputRecord
	_, err := finput(cc.conin, events[:])
	if err != nil {
		return nil, fmt.Errorf("read coninput events: %w", err)
	}

	var evs []Event
	for _, event := range events {
		if e := parseConInputEvent(event, &d.prevMouseState); e != nil {
			evs = append(evs, e)
		}
	}

	return detectConInputQuerySequences(evs), nil
}

// Using ConInput API, Windows Terminal responds to sequence query events with
// KEY_EVENT_RECORDs so we need to collect them and parse them as a single
// sequence.
// Is this a hack?
func detectConInputQuerySequences(events []Event) []Event {
	var newEvents []Event
	start, end := -1, -1

loop:
	for i, e := range events {
		switch e := e.(type) {
		case KeyDownEvent:
			switch e.Rune {
			case ansi.ESC, ansi.CSI, ansi.OSC, ansi.DCS, ansi.APC:
				// start of a sequence
				if start == -1 {
					start = i
				}
			}
		default:
			break loop
		}
		end = i
	}

	if start == -1 || end <= start {
		return events
	}

	var seq []byte
	for i := start; i <= end; i++ {
		switch e := events[i].(type) {
		case KeyDownEvent:
			seq = append(seq, byte(e.Rune))
		}
	}

	n, seqevent := ParseSequence(seq)
	switch seqevent.(type) {
	case UnknownEvent:
		// We're not interested in unknown events
	default:
		if start+n > len(events) {
			return events
		}
		newEvents = events[:start]
		newEvents = append(newEvents, seqevent)
		newEvents = append(newEvents, events[start+n:]...)
		return detectConInputQuerySequences(newEvents)
	}

	return events
}

func parseConInputEvent(event coninput.InputRecord, ps *coninput.ButtonState) Event {
	switch e := event.Unwrap().(type) {
	case coninput.KeyEventRecord:
		return parseWin32InputKeyEvent(e.VirtualKeyCode, e.VirtualScanCode,
			e.Char, e.KeyDown, e.ControlKeyState, e.RepeatCount)

	case coninput.WindowBufferSizeEventRecord:
		return WindowSizeEvent{
			Width:  int(e.Size.X),
			Height: int(e.Size.Y),
		}
	case coninput.MouseEventRecord:
		mevent := mouseEvent(*ps, e)
		*ps = e.ButtonState
		return mevent
	case coninput.FocusEventRecord, coninput.MenuEventRecord:
		// ignore
	}
	return nil
}

func mouseEventButton(p, s coninput.ButtonState) (button MouseButton, isRelease bool) {
	btn := p ^ s
	if btn&s == 0 {
		isRelease = true
	}

	if btn == 0 {
		switch {
		case s&coninput.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseButtonLeft
		case s&coninput.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseButtonMiddle
		case s&coninput.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseButtonRight
		case s&coninput.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseButtonBackward
		case s&coninput.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseButtonForward
		}
		return
	}

	switch {
	case btn == coninput.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseButtonLeft
	case btn == coninput.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseButtonRight
	case btn == coninput.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseButtonMiddle
	case btn == coninput.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseButtonBackward
	case btn == coninput.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseButtonForward
	}

	return
}

func mouseEvent(p coninput.ButtonState, e coninput.MouseEventRecord) (ev Event) {
	var mod Mod
	var isRelease bool
	if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
		mod |= Alt
	}
	if e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED) {
		mod |= Ctrl
	}
	if e.ControlKeyState.Contains(coninput.SHIFT_PRESSED) {
		mod |= Shift
	}
	m := mouse{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		m.Button, isRelease = mouseEventButton(p, e.ButtonState)
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			m.Button = MouseButtonWheelUp
		} else {
			m.Button = MouseButtonWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		if e.WheelDirection > 0 {
			m.Button = MouseButtonWheelRight
		} else {
			m.Button = MouseButtonWheelLeft
		}
	case coninput.MOUSE_MOVED:
		m.Button, _ = mouseEventButton(p, e.ButtonState)
		return MouseMoveEvent(m)
	}

	if isRelease {
		return MouseUpEvent(m)
	}

	return MouseDownEvent(m)
}
