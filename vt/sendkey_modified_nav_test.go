package vt

import (
	"testing"
	"time"
)

func readSendKey(t *testing.T, e *Emulator, key KeyPressEvent) string {
	t.Helper()
	ch := make(chan string, 1)
	errc := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		n, err := e.Read(buf)
		if err != nil && n == 0 {
			errc <- err
			return
		}
		ch <- string(buf[:n])
	}()
	// Give the reader a moment to block on the pipe before writing.
	time.Sleep(10 * time.Millisecond)
	e.SendKey(key)
	select {
	case s := <-ch:
		return s
	case err := <-errc:
		t.Fatalf("read error: %v", err)
		return ""
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for SendKey output")
		return ""
	}
}

func TestSendKeyModifiedCursorKeys(t *testing.T) {
	e := NewEmulator(80, 24)

	cases := []struct {
		name string
		key  KeyPressEvent
		want string
		set  string // optional mode sequence written before SendKey
	}{
		{"up", KeyPressEvent{Code: KeyUp}, "\x1b[A", ""},
		{"ctrl-up", KeyPressEvent{Code: KeyUp, Mod: ModCtrl}, "\x1b[1;5A", ""},
		{"alt-up", KeyPressEvent{Code: KeyUp, Mod: ModAlt}, "\x1b[1;3A", ""},
		{"shift-left", KeyPressEvent{Code: KeyLeft, Mod: ModShift}, "\x1b[1;2D", ""},
		{"ctrl-home", KeyPressEvent{Code: KeyHome, Mod: ModCtrl}, "\x1b[1;5H", ""},
		{"ctrl-pgup", KeyPressEvent{Code: KeyPgUp, Mod: ModCtrl}, "\x1b[5;5~", ""},
		{"up-app", KeyPressEvent{Code: KeyUp}, "\x1bOA", "\x1b[?1h"},
		{"ctrl-up-app", KeyPressEvent{Code: KeyUp, Mod: ModCtrl}, "\x1b[1;5A", "\x1b[?1h"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			term := NewEmulator(80, 24)
			if tc.set != "" {
				term.WriteString(tc.set)
			}
			got := readSendKey(t, term, tc.key)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			_ = term.Close()
		})
	}
	_ = e.Close()
}
