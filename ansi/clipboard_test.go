package ansi_test

import (
	"io"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestClipboardNewClipboard(t *testing.T) {
	tt := []struct {
		name   byte
		data   string
		expect string
	}{
		{'c', "Hello Test", "\x1b]52;c;SGVsbG8gVGVzdA==\x07"},
		{'p', "Ansi Test", "\x1b]52;p;QW5zaSBUZXN0\x07"},
		{'c', "", "\x1b]52;c;\x07"},
		{'p', "?", "\x1b]52;p;Pw==\x07"},
		{ansi.SystemClipboard, "test", "\x1b]52;c;dGVzdA==\x07"},
	}
	for _, tp := range tt {
		cb := ansi.SetClipboard(tp.name, tp.data)
		if cb != tp.expect {
			t.Errorf("SetClipboard(%q, %q) = %q, want %q", tp.name, tp.data, cb, tp.expect)
		}
	}
}

func TestClipboardReset(t *testing.T) {
	cb := ansi.ResetClipboard(ansi.PrimaryClipboard)
	if cb != "\x1b]52;p;\x07" {
		t.Errorf("Unexpected clipboard reset: %q", cb)
	}
}

func TestClipboardRequest(t *testing.T) {
	cb := ansi.RequestClipboard(ansi.PrimaryClipboard)
	if cb != "\x1b]52;p;?\x07" {
		t.Errorf("Unexpected clipboard request: %q", cb)
	}
}

func BenchmarkWriteSetClipboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ansi.WriteSetClipboard(io.Discard, ansi.SystemClipboard, "Benchmark Test Data")
	}
}

func BenchmarkSetClipboard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, ansi.SetClipboard(ansi.SystemClipboard, "Benchmark Test Data"))
	}
}
