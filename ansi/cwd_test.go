package ansi_test

import (
	"io"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNotifyWorkingDirectory_LocalFile(t *testing.T) {
	h := ansi.NotifyWorkingDirectory("localhost", "path", "to", "file")
	if h != "\x1b]7;file://localhost/path/to/file\x07" {
		t.Errorf("Unexpected url: %s", h)
	}
}

func TestNotifyWorkingDirectory_RemoteFile(t *testing.T) {
	h := ansi.NotifyWorkingDirectory("example.com", "path", "to", "file")
	if h != "\x1b]7;file://example.com/path/to/file\x07" {
		t.Errorf("Unexpected url: %s", h)
	}
}

func BenchmarkNotifyWorkingDirectory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, ansi.NotifyWorkingDirectory("localhost", "path", "to", "file"))
	}
}

func BenchmarkWriteNotifyWorkingDirectory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ansi.WriteNotifyWorkingDirectory(io.Discard, "localhost", "path", "to", "file")
	}
}
