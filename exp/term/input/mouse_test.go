package input

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi"
)

func TestMouseEvent_String(t *testing.T) {
	tt := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			name:     "unknown",
			event:    MouseDownEvent{Button: MouseButton(0xff)},
			expected: "unknown",
		},
		{
			name:     "left",
			event:    MouseDownEvent{Button: MouseLeft},
			expected: "left",
		},
		{
			name:     "right",
			event:    MouseDownEvent{Button: MouseRight},
			expected: "right",
		},
		{
			name:     "middle",
			event:    MouseDownEvent{Button: MouseMiddle},
			expected: "middle",
		},
		{
			name:     "release",
			event:    MouseUpEvent{Button: MouseNone},
			expected: "",
		},
		{
			name:     "wheelup",
			event:    MouseWheelEvent{Button: MouseWheelUp},
			expected: "wheelup",
		},
		{
			name:     "wheeldown",
			event:    MouseWheelEvent{Button: MouseWheelDown},
			expected: "wheeldown",
		},
		{
			name:     "wheelleft",
			event:    MouseWheelEvent{Button: MouseWheelLeft},
			expected: "wheelleft",
		},
		{
			name:     "wheelright",
			event:    MouseWheelEvent{Button: MouseWheelRight},
			expected: "wheelright",
		},
		{
			name:     "motion",
			event:    MouseMotionEvent{Button: MouseNone},
			expected: "motion",
		},
		{
			name:     "shift+left",
			event:    MouseUpEvent{Button: MouseLeft, Mod: Shift},
			expected: "shift+left",
		},
		{
			name: "shift+left", event: MouseDownEvent{Button: MouseLeft, Mod: Shift},
			expected: "shift+left",
		},
		{
			name:     "ctrl+shift+left",
			event:    MouseDownEvent{Button: MouseLeft, Mod: Ctrl | Shift},
			expected: "ctrl+shift+left",
		},
		{
			name:     "alt+left",
			event:    MouseDownEvent{Button: MouseLeft, Mod: Alt},
			expected: "alt+left",
		},
		{
			name:     "ctrl+left",
			event:    MouseDownEvent{Button: MouseLeft, Mod: Ctrl},
			expected: "ctrl+left",
		},
		{
			name:     "ctrl+alt+left",
			event:    MouseDownEvent{Button: MouseLeft, Mod: Alt | Ctrl},
			expected: "ctrl+alt+left",
		},
		{
			name:     "ctrl+alt+shift+left",
			event:    MouseDownEvent{Button: MouseLeft, Mod: Alt | Ctrl | Shift},
			expected: "ctrl+alt+shift+left",
		},
		{
			name:     "ignore coordinates",
			event:    MouseDownEvent{X: 100, Y: 200, Button: MouseLeft},
			expected: "left",
		},
		{
			name:     "broken type",
			event:    MouseDownEvent{Button: MouseButton(120)},
			expected: "unknown",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := fmt.Sprint(tc.event)

			if tc.expected != actual {
				t.Fatalf("expected %q but got %q",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseX10MouseDownEvent(t *testing.T) {
	encode := func(b byte, x, y int) []byte {
		return []byte{
			'\x1b',
			'[',
			'M',
			byte(32) + b,
			byte(x + 32 + 1),
			byte(y + 32 + 1),
		}
	}

	tt := []struct {
		name     string
		buf      []byte
		expected Event
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0b0000_0000, 0, 0),
			expected: MouseDownEvent{X: 0, Y: 0, Button: MouseLeft},
		},
		{
			name:     "max position",
			buf:      encode(0b0000_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: MouseDownEvent{X: 222, Y: 222, Button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0b0000_0000, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(0b0000_0001, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(0b0010_0001, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(0b0000_0010, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "right in motion",
			buf:      encode(0b0010_0010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(0b0010_0011, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(0b0100_0000, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(0b0100_0001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(0b0100_0010, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(0b0100_0011, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelRight},
		},
		{
			name:     "release",
			buf:      encode(0b0000_0011, 32, 16),
			expected: MouseUpEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "backward",
			buf:      encode(0b1000_0000, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(0b1000_0001, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseForward},
		},
		{
			name:     "button 10",
			buf:      encode(0b1000_0010, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseExtra1},
		},
		{
			name:     "button 11",
			buf:      encode(0b1000_0011, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseExtra2},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(0b0000_1010, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Alt, Button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(0b0001_0010, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Ctrl, Button: MouseRight},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "alt+right in motion",
			buf:      encode(0b0010_1010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Mod: Alt, Button: MouseRight},
		},
		{
			name:     "ctrl+right in motion",
			buf:      encode(0b0011_0010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Mod: Ctrl, Button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(0b0001_1010, 32, 16),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Alt | Ctrl, Button: MouseRight},
		},
		{
			name:     "ctrl+wheel up",
			buf:      encode(0b0101_0000, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Ctrl, Button: MouseWheelUp},
		},
		{
			name:     "alt+wheel down",
			buf:      encode(0b0100_1001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Alt, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel down",
			buf:      encode(0b0101_1001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Alt | Ctrl, Button: MouseWheelDown},
		},
		// Overflow position.
		{
			name:     "overflow position",
			buf:      encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: MouseMotionEvent{X: -6, Y: -33, Button: MouseLeft},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := parseX10MouseEvent(tc.buf)

			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseSGRMouseEvent(t *testing.T) {
	encode := func(b, x, y int, r bool) []byte {
		re := 'M'
		if r {
			re = 'm'
		}
		return []byte(fmt.Sprintf("\x1b[<%d;%d;%d%c", b, x+1, y+1, re))
	}

	tt := []struct {
		name     string
		buf      []byte
		expected Event
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0, 0, 0, false),
			expected: MouseDownEvent{X: 0, Y: 0, Button: MouseLeft},
		},
		{
			name:     "225 position",
			buf:      encode(0, 225, 225, false),
			expected: MouseDownEvent{X: 225, Y: 225, Button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(32, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left",
			buf:      encode(0, 32, 16, true),
			expected: MouseUpEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(33, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, true),
			expected: MouseUpEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, true),
			expected: MouseUpEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(35, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(64, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(65, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(66, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(67, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelRight},
		},
		{
			name:     "backward",
			buf:      encode(128, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "backward in motion",
			buf:      encode(160, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(129, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Button: MouseForward},
		},
		{
			name:     "forward in motion",
			buf:      encode(161, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseForward},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(10, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Alt, Button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(18, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Ctrl, Button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(26, 32, 16, false),
			expected: MouseDownEvent{X: 32, Y: 16, Mod: Alt | Ctrl, Button: MouseRight},
		},
		{
			name:     "alt+wheel",
			buf:      encode(73, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Alt, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+wheel",
			buf:      encode(81, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Ctrl, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel",
			buf:      encode(89, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Alt | Ctrl, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+shift+wheel",
			buf:      encode(93, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: Alt | Shift | Ctrl, Button: MouseWheelDown},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			final := tc.buf[len(tc.buf)-1]
			params := ansi.Params(tc.buf)
			actual := parseSGRMouseEvent(params, final)
			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}
