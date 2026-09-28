package vt

import (
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// TestDrawFullFrame verifies that Draw repaints the entire screen every
// time, not just lines written since the last draw. Hosts that rebuild the
// frame on each render pass (like Bubble Tea) lose content otherwise.
func TestDrawFullFrame(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	if _, err := e.WriteString("hello"); err != nil {
		t.Fatal(err)
	}

	render := func() string {
		scr := uv.NewScreenBuffer(20, 5)
		e.Draw(scr, scr.Bounds())
		return scr.Render()
	}

	first := render()
	if !strings.Contains(first, "hello") {
		t.Fatalf("first draw missing content:\n%q", first)
	}

	// A second draw of the same emulator must contain the same content,
	// even though nothing was written in between.
	second := render()
	if !strings.Contains(second, "hello") {
		t.Fatalf("second draw lost content:\n%q", second)
	}
}

// TestCursorHidden verifies the cursor visibility state tracks DECSCUSR /
// DECTCEM sequences so hosts can hide the cursor when the child does.
func TestCursorHidden(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	if e.CursorHidden() {
		t.Fatal("cursor should be visible by default")
	}

	if _, err := e.WriteString("\x1b[?25l"); err != nil { // hide cursor
		t.Fatal(err)
	}
	if !e.CursorHidden() {
		t.Fatal("cursor should be hidden after DECTCEM reset")
	}

	if _, err := e.WriteString("\x1b[?25h"); err != nil { // show cursor
		t.Fatal(err)
	}
	if e.CursorHidden() {
		t.Fatal("cursor should be visible after DECTCEM set")
	}
}

// TestCursorStyle verifies the cursor style and blink state track DECSCUSR
// sequences.
func TestCursorStyle(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	style, blink := e.CursorStyle()
	if style != CursorBlock || !blink {
		t.Fatalf("default cursor should be a blinking block, got style=%v blink=%v", style, blink)
	}

	// Set cursor style: steady bar.
	if _, err := e.WriteString("\x1b[6 q"); err != nil {
		t.Fatal(err)
	}
	style, blink = e.CursorStyle()
	if style != CursorBar || blink {
		t.Fatalf("expected steady bar, got style=%v blink=%v", style, blink)
	}
}

// readInputs reads n bytes the emulator produced on its input pipe.
func readInputs(t *testing.T, e *Emulator, n int) string {
	t.Helper()

	type result struct {
		data string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		buf := make([]byte, 64)
		var out []byte
		for len(out) < n {
			read, err := e.Read(buf)
			if err != nil {
				done <- result{string(out), err}
				return
			}
			out = append(out, buf[:read]...)
		}
		done <- result{string(out), nil}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("read input: %v", r.err)
		}
		return r.data
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for terminal input")
		return ""
	}
}

// TestSendKeyPrintableText verifies printable keys send their resolved
// text rather than the unshifted key code.
func TestSendKeyPrintableText(t *testing.T) {
	e := newTestTerminal(t, 20, 5)

	done := make(chan string, 1)
	go func() { done <- readInputs(t, e, 3) }()

	e.SendKey(uv.KeyPressEvent{Code: 'a', Text: "A"})
	e.SendKey(uv.KeyPressEvent{Code: '1', Text: "!"})
	e.SendKey(uv.KeyPressEvent{Code: 'z'})

	select {
	case got := <-done:
		if got != "A!z" {
			t.Fatalf("expected %q, got %q", "A!z", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}

// TestSendKeyApplicationCursorKeys verifies arrow keys are encoded for the
// cursor key mode the child enabled: a full-screen program that turns on
// application cursor keys expects SS3 sequences, and one that does not
// expects CSI ones.
func TestSendKeyApplicationCursorKeys(t *testing.T) {
	e := newTestTerminal(t, 20, 5)
	if _, err := e.WriteString("\x1b[?1h"); err != nil { // application cursor keys
		t.Fatal(err)
	}

	done := make(chan string, 1)
	go func() { done <- readInputs(t, e, 3) }()

	e.SendKey(uv.KeyPressEvent{Code: KeyUp})

	select {
	case got := <-done:
		if got != "\x1bOA" {
			t.Fatalf("expected application cursor encoding %q, got %q", "\x1bOA", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}

	// Without the mode, the same key must use the CSI form.
	e2 := newTestTerminal(t, 20, 5)
	done2 := make(chan string, 1)
	go func() { done2 <- readInputs(t, e2, 3) }()
	e2.SendKey(uv.KeyPressEvent{Code: KeyUp})

	select {
	case got := <-done2:
		if got != "\x1b[A" {
			t.Fatalf("expected normal cursor encoding %q, got %q", "\x1b[A", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}
