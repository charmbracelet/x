package vt

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestReplyFuncNeverBlocksWrite(t *testing.T) {
	e := newTestTerminal(t, 10, 2)
	var got strings.Builder
	e.SetReplyFunc(func(p []byte) { got.Write(p) })

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = e.WriteString("\x1b[5n\x1b[6n\x1b[c\x1b[?2004$p")
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Write blocked with a reply func set")
	}

	want := ansi.DeviceStatusReport(ansi.DECStatusReport(0)) + "\x1b[1;1R\x1b[?62;1;6;22c\x1b[?2004;2$y"
	if got.String() != want {
		t.Errorf("replies: got = %q, want = %q", got.String(), want)
	}
}

func TestReplyFuncDiscards(t *testing.T) {
	e := newTestTerminal(t, 10, 2)
	e.SetReplyFunc(func([]byte) {})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 100 {
			_, _ = e.WriteString("\x1b[c\x1b[6n\x1b]11;?\x07")
		}
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Write blocked with a discarding reply func")
	}
}

func TestRepliesGoToReadWithoutReplyFunc(t *testing.T) {
	e := newTestTerminal(t, 10, 2)
	go func() { _, _ = e.WriteString("\x1b[6n") }()
	buf := make([]byte, 16)
	n, err := e.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if got, want := string(buf[:n]), "\x1b[1;1R"; got != want {
		t.Errorf("Read: got = %q, want = %q", got, want)
	}
}
