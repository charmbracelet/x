package vt

import (
	"io"
	"testing"
)

func TestCursorStyleCallbackReceivesBlink(t *testing.T) {
	e := NewEmulator(10, 3)
	go io.Copy(io.Discard, e)

	type got struct {
		style CursorStyle
		blink bool
	}
	var calls []got
	e.SetCallbacks(Callbacks{CursorStyle: func(style CursorStyle, blink bool) {
		calls = append(calls, got{style, blink})
	}})

	// CSI 5 SP q = blinking bar, CSI 6 SP q = steady bar
	// CSI 1 SP q = blinking block, CSI 2 SP q = steady block
	for _, seq := range []string{"\x1b[5 q", "\x1b[6 q", "\x1b[1 q", "\x1b[2 q"} {
		e.WriteString(seq)
	}

	want := []got{
		{CursorBar, true},
		{CursorBar, false},
		{CursorBlock, true},
		{CursorBlock, false},
	}
	if len(calls) != len(want) {
		t.Fatalf("got %d callbacks, want %d: %#v", len(calls), len(want), calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("call %d: got %#v, want %#v", i, calls[i], want[i])
		}
	}
}
