package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestSendKeyModified(t *testing.T) {
	tests := []struct {
		name string
		key  uv.KeyPressEvent
		want string
	}{
		{"ctrl+left", uv.KeyPressEvent{Code: KeyLeft, Mod: ModCtrl}, "\x1b[1;5D"},
		{"shift+up", uv.KeyPressEvent{Code: KeyUp, Mod: ModShift}, "\x1b[1;2A"},
		{"alt+right", uv.KeyPressEvent{Code: KeyRight, Mod: ModAlt}, "\x1b[1;3C"},
		{"ctrl+shift+end", uv.KeyPressEvent{Code: KeyEnd, Mod: ModCtrl | ModShift}, "\x1b[1;6F"},
		{"ctrl+delete", uv.KeyPressEvent{Code: KeyDelete, Mod: ModCtrl}, "\x1b[3;5~"},
		{"shift+pgup", uv.KeyPressEvent{Code: KeyPgUp, Mod: ModShift}, "\x1b[5;2~"},
		{"ctrl+f1", uv.KeyPressEvent{Code: KeyF1, Mod: ModCtrl}, "\x1b[1;5P"},
		{"alt+f5", uv.KeyPressEvent{Code: KeyF5, Mod: ModAlt}, "\x1b[15;3~"},
		{"shift+enter", uv.KeyPressEvent{Code: KeyEnter, Mod: ModShift}, "\r"},
		{"ctrl+enter", uv.KeyPressEvent{Code: KeyEnter, Mod: ModCtrl}, "\r"},
		{"alt+enter", uv.KeyPressEvent{Code: KeyEnter, Mod: ModAlt}, "\x1b\r"},
		{"ctrl+backspace", uv.KeyPressEvent{Code: KeyBackspace, Mod: ModCtrl}, "\x08"},
		{"ctrl+shift+c", uv.KeyPressEvent{Code: 'C', Mod: ModCtrl | ModShift}, "\x03"},
		{"ctrl+alt+c", uv.KeyPressEvent{Code: 'c', Mod: ModCtrl | ModAlt}, "\x1b\x03"},
		{"shift+tab", uv.KeyPressEvent{Code: KeyTab, Mod: ModShift}, "\x1b[Z"},
		{"alt+kp8", uv.KeyPressEvent{Code: KeyKp8, Mod: ModAlt}, "\x1b8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestTerminal(t, 20, 5)
			done := make(chan string, 1)
			go func() { done <- readInputs(t, e, len(tt.want)) }()
			e.SendKey(tt.key)
			if got := <-done; got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSendKeyModifiedIgnoresCursorMode(t *testing.T) {
	e := newTestTerminal(t, 20, 5)
	if _, err := e.WriteString("\x1b[?1h"); err != nil {
		t.Fatal(err)
	}
	done := make(chan string, 1)
	go func() { done <- readInputs(t, e, 6) }()
	e.SendKey(uv.KeyPressEvent{Code: KeyLeft, Mod: ModCtrl})
	if got := <-done; got != "\x1b[1;5D" {
		t.Fatalf("expected %q, got %q", "\x1b[1;5D", got)
	}
}
