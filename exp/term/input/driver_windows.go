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

	var ps coninput.ButtonState // keep track of previous mouse state
	var evs []Event
	for _, event := range events {
		if e := parseConInputEvent(event, &ps); e != nil {
			evs = append(evs, e...)
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
		case KeyEvent:
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
		case KeyEvent:
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

func parseConInputEvent(event coninput.InputRecord, ps *coninput.ButtonState) []Event {
	switch e := event.Unwrap().(type) {
	case coninput.KeyEventRecord:
		isCtrl := e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED)
		k, ok := vkKeyEvent[e.VirtualKeyCode]
		if !ok && isCtrl {
			k = vkCtrlRune(keyType(e), e)
		} else if !ok {
			k = KeyEvent{Rune: e.Char}
		}
		if isCtrl {
			k.Mod |= Ctrl
		}
		if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
			k.Mod |= Alt
		}
		if e.ControlKeyState.Contains(coninput.SHIFT_PRESSED) {
			k.Mod |= Shift
		}

		// XXX: the following keys when set mean that the key is ON, not that
		// it was pressed. We should probably ignore them.
		if e.ControlKeyState.Contains(coninput.NUMLOCK_ON|coninput.CAPSLOCK_ON|coninput.SCROLLLOCK_ON) && k.Rune == 0 && k.Sym == 0 {
			return nil
		}

		if e.RepeatCount > 1 {
			k.Action = KeyRepeat
		} else if !e.KeyDown {
			k.Action = KeyRelease
		}

		var kevents []Event
		for i := 0; i < int(e.RepeatCount); i++ {
			kevents = append(kevents, k)
		}

		return kevents

	case coninput.WindowBufferSizeEventRecord:
		return []Event{WindowSizeEvent{
			Width:  int(e.Size.X),
			Height: int(e.Size.Y),
		}}
	case coninput.MouseEventRecord:
		mevent := mouseEvent(*ps, e)
		*ps = e.ButtonState
		return []Event{mevent}
	case coninput.FocusEventRecord, coninput.MenuEventRecord:
		// ignore
	}
	return nil
}

func mouseEventButton(p, s coninput.ButtonState) (button MouseButton, action MouseAction) {
	btn := p ^ s
	action = MouseActionPress
	if btn&s == 0 {
		action = MouseActionRelease
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

	return button, action
}

func mouseEvent(p coninput.ButtonState, e coninput.MouseEventRecord) MouseEvent {
	var mod Mod
	if e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
		mod |= Alt
	}
	if e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED) {
		mod |= Ctrl
	}
	if e.ControlKeyState.Contains(coninput.SHIFT_PRESSED) {
		mod |= Shift
	}
	ev := MouseEvent{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		ev.Button, ev.Action = mouseEventButton(p, e.ButtonState)
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelUp
		} else {
			ev.Button = MouseButtonWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelRight
		} else {
			ev.Button = MouseButtonWheelLeft
		}
	case coninput.MOUSE_MOVED:
		ev.Button, _ = mouseEventButton(p, e.ButtonState)
		ev.Action = MouseActionMotion
	}

	return ev
}

var vkKeyEvent = map[coninput.VirtualKeyCode]KeyEvent{
	coninput.VK_RETURN:    {Sym: KeyEnter},
	coninput.VK_BACK:      {Sym: KeyBackspace},
	coninput.VK_TAB:       {Sym: KeyTab},
	coninput.VK_ESCAPE:    {Sym: KeyEscape},
	coninput.VK_SPACE:     {Sym: KeySpace, Rune: ' '},
	coninput.VK_UP:        {Sym: KeyUp},
	coninput.VK_DOWN:      {Sym: KeyDown},
	coninput.VK_RIGHT:     {Sym: KeyRight},
	coninput.VK_LEFT:      {Sym: KeyLeft},
	coninput.VK_HOME:      {Sym: KeyHome},
	coninput.VK_END:       {Sym: KeyEnd},
	coninput.VK_PRIOR:     {Sym: KeyPgUp},
	coninput.VK_NEXT:      {Sym: KeyPgDown},
	coninput.VK_DELETE:    {Sym: KeyDelete},
	coninput.VK_SELECT:    {Sym: KeySelect},
	coninput.VK_SNAPSHOT:  {Sym: KeyPrintScreen},
	coninput.VK_INSERT:    {Sym: KeyInsert},
	coninput.VK_LWIN:      {Sym: KeyLeftSuper},
	coninput.VK_RWIN:      {Sym: KeyRightSuper},
	coninput.VK_APPS:      {Sym: KeyMenu},
	coninput.VK_NUMPAD0:   {Sym: KeyKp0},
	coninput.VK_NUMPAD1:   {Sym: KeyKp1},
	coninput.VK_NUMPAD2:   {Sym: KeyKp2},
	coninput.VK_NUMPAD3:   {Sym: KeyKp3},
	coninput.VK_NUMPAD4:   {Sym: KeyKp4},
	coninput.VK_NUMPAD5:   {Sym: KeyKp5},
	coninput.VK_NUMPAD6:   {Sym: KeyKp6},
	coninput.VK_NUMPAD7:   {Sym: KeyKp7},
	coninput.VK_NUMPAD8:   {Sym: KeyKp8},
	coninput.VK_NUMPAD9:   {Sym: KeyKp9},
	coninput.VK_MULTIPLY:  {Sym: KeyKpMul},
	coninput.VK_ADD:       {Sym: KeyKpPlus},
	coninput.VK_SEPARATOR: {Sym: KeyKpComma},
	coninput.VK_SUBTRACT:  {Sym: KeyKpMinus},
	coninput.VK_DECIMAL:   {Sym: KeyKpPeriod},
	coninput.VK_DIVIDE:    {Sym: KeyKpDiv},
	coninput.VK_F1:        {Sym: KeyF1},
	coninput.VK_F2:        {Sym: KeyF2},
	coninput.VK_F3:        {Sym: KeyF3},
	coninput.VK_F4:        {Sym: KeyF4},
	coninput.VK_F5:        {Sym: KeyF5},
	coninput.VK_F6:        {Sym: KeyF6},
	coninput.VK_F7:        {Sym: KeyF7},
	coninput.VK_F8:        {Sym: KeyF8},
	coninput.VK_F9:        {Sym: KeyF9},
	coninput.VK_F10:       {Sym: KeyF10},
	coninput.VK_F11:       {Sym: KeyF11},
	coninput.VK_F12:       {Sym: KeyF12},
	coninput.VK_F13:       {Sym: KeyF13},
	coninput.VK_F14:       {Sym: KeyF14},
	coninput.VK_F15:       {Sym: KeyF15},
	coninput.VK_F16:       {Sym: KeyF16},
	coninput.VK_F17:       {Sym: KeyF17},
	coninput.VK_F18:       {Sym: KeyF18},
	coninput.VK_F19:       {Sym: KeyF19},
	coninput.VK_F20:       {Sym: KeyF20},
	coninput.VK_F21:       {Sym: KeyF21},
	coninput.VK_F22:       {Sym: KeyF22},
	coninput.VK_F23:       {Sym: KeyF23},
	coninput.VK_F24:       {Sym: KeyF24},
	coninput.VK_NUMLOCK:   {Sym: KeyNumLock},
	coninput.VK_SCROLL:    {Sym: KeyScrollLock},
	coninput.VK_LSHIFT:    {Sym: KeyLeftShift},
	coninput.VK_RSHIFT:    {Sym: KeyRightShift},
	coninput.VK_LCONTROL:  {Sym: KeyLeftCtrl},
	coninput.VK_RCONTROL:  {Sym: KeyRightCtrl},
	coninput.VK_LMENU:     {Sym: KeyLeftAlt},
	coninput.VK_RMENU:     {Sym: KeyRightAlt},
	coninput.VK_OEM_4:     {Rune: '['},
	// TODO: add more keys
}

func vkCtrlRune(k KeyEvent, e coninput.KeyEventRecord) KeyEvent {
	switch e.Char {
	case '@':
		k.Rune = '@'
	case '\x01':
		k.Rune = 'a'
	case '\x02':
		k.Rune = 'b'
	case '\x03':
		k.Rune = 'c'
	case '\x04':
		k.Rune = 'd'
	case '\x05':
		k.Rune = 'e'
	case '\x06':
		k.Rune = 'f'
	case '\a':
		k.Rune = 'g'
	case '\b':
		k.Rune = 'h'
	case '\t':
		k.Rune = 'i'
	case '\n':
		k.Rune = 'j'
	case '\v':
		k.Rune = 'k'
	case '\f':
		k.Rune = 'l'
	case '\r':
		k.Rune = 'm'
	case '\x0e':
		k.Rune = 'n'
	case '\x0f':
		k.Rune = 'o'
	case '\x10':
		k.Rune = 'p'
	case '\x11':
		k.Rune = 'q'
	case '\x12':
		k.Rune = 'r'
	case '\x13':
		k.Rune = 's'
	case '\x14':
		k.Rune = 't'
	case '\x15':
		k.Rune = 'u'
	case '\x16':
		k.Rune = 'v'
	case '\x17':
		k.Rune = 'w'
	case '\x18':
		k.Rune = 'x'
	case '\x19':
		k.Rune = 'y'
	case '\x1a':
		k.Rune = 'z'
	case '\x1b':
		k.Rune = ']'
	case '\x1c':
		k.Rune = '\\'
	case '\x1f':
		k.Rune = '_'
	}

	switch e.VirtualKeyCode {
	case coninput.VK_OEM_4:
		k.Rune = '['
	}

	return k
}

func keyType(e coninput.KeyEventRecord) KeyEvent {
	code := e.VirtualKeyCode

	switch code {
	case coninput.VK_RETURN:
		return KeyEvent{Sym: KeyEnter}
	case coninput.VK_BACK:
		return KeyEvent{Sym: KeyBackspace}
	case coninput.VK_TAB:
		return KeyEvent{Sym: KeyTab}
	case coninput.VK_SPACE:
		return KeyEvent{Sym: KeySpace, Rune: ' '}
	case coninput.VK_ESCAPE:
		return KeyEvent{Sym: KeyEscape}
	case coninput.VK_UP:
		return KeyEvent{Sym: KeyUp}
	case coninput.VK_DOWN:
		return KeyEvent{Sym: KeyDown}
	case coninput.VK_RIGHT:
		return KeyEvent{Sym: KeyRight}
	case coninput.VK_LEFT:
		return KeyEvent{Sym: KeyLeft}
	case coninput.VK_HOME:
		return KeyEvent{Sym: KeyHome}
	case coninput.VK_END:
		return KeyEvent{Sym: KeyEnd}
	case coninput.VK_PRIOR:
		return KeyEvent{Sym: KeyPgUp}
	case coninput.VK_NEXT:
		return KeyEvent{Sym: KeyPgDown}
	case coninput.VK_DELETE:
		return KeyEvent{Sym: KeyDelete}
	default:
		if e.ControlKeyState&(coninput.LEFT_CTRL_PRESSED|coninput.RIGHT_CTRL_PRESSED) == 0 {
			return KeyEvent{Rune: e.Char}
		}

		k := KeyEvent{Mod: Ctrl}
		switch e.Char {
		case '@':
			k.Rune = '@'
		case '\x01':
			k.Rune = 'a'
		case '\x02':
			k.Rune = 'b'
		case '\x03':
			k.Rune = 'c'
		case '\x04':
			k.Rune = 'd'
		case '\x05':
			k.Rune = 'e'
		case '\x06':
			k.Rune = 'f'
		case '\a':
			k.Rune = 'g'
		case '\b':
			k.Rune = 'h'
		case '\t':
			k.Rune = 'i'
		case '\n':
			k.Rune = 'j'
		case '\v':
			k.Rune = 'k'
		case '\f':
			k.Rune = 'l'
		case '\r':
			k.Rune = 'm'
		case '\x0e':
			k.Rune = 'n'
		case '\x0f':
			k.Rune = 'o'
		case '\x10':
			k.Rune = 'p'
		case '\x11':
			k.Rune = 'q'
		case '\x12':
			k.Rune = 'r'
		case '\x13':
			k.Rune = 's'
		case '\x14':
			k.Rune = 't'
		case '\x15':
			k.Rune = 'u'
		case '\x16':
			k.Rune = 'v'
		case '\x17':
			k.Rune = 'w'
		case '\x18':
			k.Rune = 'x'
		case '\x19':
			k.Rune = 'y'
		case '\x1a':
			k.Rune = 'z'
		case '\x1b':
			k.Rune = ']'
		case '\x1c':
			k.Rune = '\\'
		case '\x1f':
			k.Rune = '_'
		}

		switch code {
		case coninput.VK_OEM_4:
			k.Rune = '['
		}

		return k
	}
}
