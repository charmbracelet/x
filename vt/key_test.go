package vt

import (
	"io"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// sendKey writes into the emulator's pipe, which blocks until a reader
// arrives, so the read has to be waiting first.
func sendKey(t *testing.T, e *Emulator, k uv.KeyEvent) string {
	t.Helper()
	out := make(chan string, 1)
	go func() {
		buf := make([]byte, 32)
		n, err := e.Read(buf)
		if err != nil && err != io.EOF {
			out <- ""
			return
		}
		out <- string(buf[:n])
	}()
	go e.SendKey(k)
	select {
	case s := <-out:
		return s
	case <-time.After(time.Second):
		t.Fatal("SendKey wrote nothing within a second")
		return ""
	}
}

// Home and End follow DECCKM, like the arrow keys beside them. terminfo
// pairs smkx (\E[?1h, which is DECCKM) with khome=\EOH and kend=\EOF, so
// an ncurses application that called keypad(true) matches incoming bytes
// against the SS3 forms and never recognizes the CSI ones.
func TestSendKeyCursorKeysFollowDECCKM(t *testing.T) {
	for _, tc := range []struct {
		name             string
		key              uv.KeyPressEvent
		normal, applicat string
	}{
		{"up", uv.KeyPressEvent{Code: KeyUp}, "\x1b[A", "\x1bOA"},
		{"down", uv.KeyPressEvent{Code: KeyDown}, "\x1b[B", "\x1bOB"},
		{"right", uv.KeyPressEvent{Code: KeyRight}, "\x1b[C", "\x1bOC"},
		{"left", uv.KeyPressEvent{Code: KeyLeft}, "\x1b[D", "\x1bOD"},
		{"home", uv.KeyPressEvent{Code: KeyHome}, "\x1b[H", "\x1bOH"},
		{"end", uv.KeyPressEvent{Code: KeyEnd}, "\x1b[F", "\x1bOF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(80, 24)
			if got := sendKey(t, e, tc.key); got != tc.normal {
				t.Errorf("DECCKM reset: got %q, want %q", got, tc.normal)
			}
			e2 := NewEmulator(80, 24)
			e2.WriteString("\x1b[?1h")
			if got := sendKey(t, e2, tc.key); got != tc.applicat {
				t.Errorf("DECCKM set: got %q, want %q", got, tc.applicat)
			}
		})
	}
}
