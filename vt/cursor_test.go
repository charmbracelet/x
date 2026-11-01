package vt

import (
	"testing"
)

func TestCursorStyleCallback(t *testing.T) {
	tests := []struct {
		name      string
		sequence  string
		wantStyle CursorStyle
		wantBlink bool
	}{
		{
			name:      "default (0) blinking block",
			sequence:  "\x1b[0 q",
			wantStyle: CursorBlock,
			wantBlink: true,
		},
		{
			name:      "explicit (1) blinking block",
			sequence:  "\x1b[1 q",
			wantStyle: CursorBlock,
			wantBlink: true,
		},
		{
			name:      "steady block (2)",
			sequence:  "\x1b[2 q",
			wantStyle: CursorBlock,
			wantBlink: false,
		},
		{
			name:      "blinking underline (3)",
			sequence:  "\x1b[3 q",
			wantStyle: CursorUnderline,
			wantBlink: true,
		},
		{
			name:      "steady underline (4)",
			sequence:  "\x1b[4 q",
			wantStyle: CursorUnderline,
			wantBlink: false,
		},
		{
			name:      "blinking bar (5)",
			sequence:  "\x1b[5 q",
			wantStyle: CursorBar,
			wantBlink: true,
		},
		{
			name:      "steady bar (6)",
			sequence:  "\x1b[6 q",
			wantStyle: CursorBar,
			wantBlink: false,
		},
		{
			name:      "missing param defaults to blinking block",
			sequence:  "\x1b[ q",
			wantStyle: CursorBlock,
			wantBlink: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(80, 24)

			// First change away from default block/blink so that default sequence triggers callback
			if tc.wantStyle == CursorBlock && tc.wantBlink {
				_, _ = e.WriteString("\x1b[4 q")
			}

			type call struct {
				style CursorStyle
				blink bool
			}
			var received []call

			e.SetCallbacks(Callbacks{
				CursorStyle: func(style CursorStyle, blink bool) {
					received = append(received, call{style: style, blink: blink})
				},
			})

			_, err := e.WriteString(tc.sequence)
			if err != nil {
				t.Fatalf("unexpected write error: %v", err)
			}

			if len(received) != 1 {
				t.Fatalf("expected 1 callback call, got %d: %+v", len(received), received)
			}

			got := received[0]
			if got.style != tc.wantStyle {
				t.Errorf("style mismatch: got %v, want %v", got.style, tc.wantStyle)
			}
			if got.blink != tc.wantBlink {
				t.Errorf("blink mismatch: got %v, want %v", got.blink, tc.wantBlink)
			}

			if e.scr.cur.Style != tc.wantStyle {
				t.Errorf("internal screen cursor style mismatch: got %v, want %v", e.scr.cur.Style, tc.wantStyle)
			}
			if e.scr.cur.Steady != !tc.wantBlink {
				t.Errorf("internal screen cursor steady mismatch: got %v, want %v", e.scr.cur.Steady, !tc.wantBlink)
			}
		})
	}
}

func TestCursorStyleCallbackDeduplication(t *testing.T) {
	e := NewEmulator(80, 24)

	var callCount int
	e.SetCallbacks(Callbacks{
		CursorStyle: func(style CursorStyle, blink bool) {
			callCount++
		},
	})

	// Changing to steady underline should invoke callback once
	_, _ = e.WriteString("\x1b[4 q")
	if callCount != 1 {
		t.Fatalf("expected 1 callback call after initial change, got %d", callCount)
	}

	// Sending identical sequence again should not trigger callback
	_, _ = e.WriteString("\x1b[4 q")
	if callCount != 1 {
		t.Fatalf("expected callback count to remain 1, got %d", callCount)
	}

	// Changing blink mode to blinking underline should trigger callback
	_, _ = e.WriteString("\x1b[3 q")
	if callCount != 2 {
		t.Fatalf("expected 2 callback calls after blink toggle, got %d", callCount)
	}
}
