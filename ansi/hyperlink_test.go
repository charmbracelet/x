package ansi_test

import (
	"io"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNewHyperlink_NoParams(t *testing.T) {
	h := ansi.SetHyperlink("https://example.com")
	if h != "\x1b]8;;https://example.com\x07" {
		t.Errorf("Unexpected hyperlink: %s", h)
	}
}

func TestNewHyperlinkParams(t *testing.T) {
	h := ansi.SetHyperlink("https://example.com", "color=blue", "size=12")
	if h != "\x1b]8;color=blue:size=12;https://example.com\x07" {
		t.Errorf("Unexpected hyperlink: %s", h)
	}
}

func TestHyperlinkReset(t *testing.T) {
	h := ansi.SetHyperlink("")
	if h != "\x1b]8;;\x07" {
		t.Errorf("Unexpected hyperlink: %s", h)
	}
}

func BenchmarkSetHyperlink(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, ansi.SetHyperlink("https://example.com", "param1", "param2"))
	}
}

func BenchmarkWriteSetHyperlink(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ansi.WriteSetHyperlink(io.Discard, "https://example.com", "param1", "param2")
	}
}
