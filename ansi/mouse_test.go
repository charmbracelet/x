package ansi

import (
	"fmt"
	"testing"
)

func TestMouseButton(t *testing.T) {
	type test struct {
		name                     string
		btn                      MouseButton
		motion, shift, alt, ctrl bool
		want                     byte
	}

	cases := []test{
		{
			name: "mouse release",
			btn:  MouseNone,
			want: 0b0000_0011,
		},
		{
			name: "mouse release with ctrl",
			btn:  MouseNone,
			ctrl: true,
			want: 0b0001_0011,
		},
		{
			name: "mouse left",
			btn:  MouseLeft,
			want: 0b0000_0000,
		},
		{
			name: "mouse right",
			btn:  MouseRight,
			want: 0b0000_0010,
		},
		{
			name: "mouse wheel up",
			btn:  MouseWheelUp,
			want: 0b0100_0000,
		},
		{
			name: "mouse wheel right",
			btn:  MouseWheelRight,
			want: 0b0100_0011,
		},
		{
			name: "mouse backward",
			btn:  MouseBackward,
			want: 0b1000_0000,
		},
		{
			name: "mouse forward",
			btn:  MouseForward,
			want: 0b1000_0001,
		},
		{
			name: "mouse button 10",
			btn:  MouseButton10,
			want: 0b1000_0010,
		},
		{
			name: "mouse button 11",
			btn:  MouseButton11,
			want: 0b1000_0011,
		},
		{
			name:   "mouse middle with motion",
			btn:    MouseMiddle,
			motion: true,
			want:   0b0010_0001,
		},
		{
			name:  "mouse middle with shift",
			btn:   MouseMiddle,
			shift: true,
			want:  0b0000_0101,
		},
		{
			name:   "mouse middle with motion and alt",
			btn:    MouseMiddle,
			motion: true,
			alt:    true,
			want:   0b0010_1001,
		},
		{
			name:  "mouse right with shift, alt, and ctrl",
			btn:   MouseRight,
			shift: true,
			alt:   true,
			ctrl:  true,
			want:  0b0001_1110,
		},
		{
			name:   "mouse button 10 with motion, shift, alt, and ctrl",
			btn:    MouseButton10,
			motion: true,
			shift:  true,
			alt:    true,
			ctrl:   true,
			want:   0b1011_1110,
		},
		{
			name:   "mouse left with motion, shift, and ctrl",
			btn:    MouseLeft,
			motion: true,
			shift:  true,
			ctrl:   true,
			want:   0b0011_0100,
		},
		{
			name: "invalid mouse button",
			btn:  MouseButton(0xff),
			want: 0b1111_1111,
		},
		{
			name:   "mouse wheel down with motion",
			btn:    MouseWheelDown,
			motion: true,
			want:   0b0110_0001,
		},
		{
			name:  "mouse wheel down with shift and ctrl",
			btn:   MouseWheelDown,
			shift: true,
			ctrl:  true,
			want:  0b0101_0101,
		},
		{
			name: "mouse wheel left with alt",
			btn:  MouseWheelLeft,
			alt:  true,
			want: 0b0100_1010,
		},
		{
			name:   "mouse middle with all modifiers",
			btn:    MouseMiddle,
			motion: true,
			shift:  true,
			alt:    true,
			ctrl:   true,
			want:   0b0011_1101,
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.btn.Button(tc.motion, tc.shift, tc.alt, tc.ctrl)
			if got != tc.want {
				t.Errorf("test %d: got %08b; want %08b", i+1, got, tc.want)
			}
		})
	}
}

func TestMouseSgr(t *testing.T) {
	type test struct {
		name    string
		btn     byte
		x, y    int
		release bool
	}

	cases := []test{
		{
			name: "mouse left",
			btn:  MouseLeft.Button(false, false, false, false),
			x:    0,
			y:    0,
		},
		{
			name: "wheel down",
			btn:  MouseWheelDown.Button(false, false, false, false),
			x:    1,
			y:    10,
		},
		{
			name: "mouse right with shift, alt, and ctrl",
			btn:  MouseRight.Button(false, true, true, true),
			x:    10,
			y:    1,
		},
		{
			name:    "mouse release",
			btn:     MouseNone.Button(false, false, false, false),
			x:       5,
			y:       5,
			release: true,
		},
		{
			name: "mouse button 10 with motion, shift, alt, and ctrl",
			btn:  MouseButton10.Button(true, true, true, true),
			x:    10,
			y:    10,
		},
		{
			name: "mouse wheel up with motion",
			btn:  MouseWheelUp.Button(true, false, false, false),
			x:    15,
			y:    15,
		},
		{
			name: "mouse middle with all modifiers",
			btn:  MouseMiddle.Button(true, true, true, true),
			x:    20,
			y:    20,
		},
		{
			name: "mouse wheel left at max coordinates",
			btn:  MouseWheelLeft.Button(false, false, false, false),
			x:    223,
			y:    223,
		},
		{
			name:    "mouse forward release",
			btn:     MouseForward.Button(false, false, false, false),
			x:       100,
			y:       100,
			release: true,
		},
		{
			name: "mouse backward with shift and ctrl",
			btn:  MouseBackward.Button(false, true, false, true),
			x:    50,
			y:    50,
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := MouseSgr(tc.btn, tc.x, tc.y, tc.release)
			action := 'M'
			if tc.release {
				action = 'm'
			}
			want := fmt.Sprintf("\x1b[<%d;%d;%d%c", tc.btn, tc.x+1, tc.y+1, action)
			if m != want {
				t.Errorf("test %d: got %q; want %q", i+1, m, want)
			}
		})
	}
}
