package sys_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi/sys"
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
		{sys.SystemClipboard, "test", "\x1b]52;c;dGVzdA==\x07"},
	}
	for _, tp := range tt {
		cb := sys.SetClipboard(tp.name, tp.data)
		if cb != tp.expect {
			t.Errorf("SetClipboard(%q, %q) = %q, want %q", tp.name, tp.data, cb, tp.expect)
		}
	}
}

func TestClipboardReset(t *testing.T) {
	cb := sys.ResetClipboard(sys.PrimaryClipboard)
	if cb != "\x1b]52;p;\x07" {
		t.Errorf("Unexpected clipboard reset: %q", cb)
	}
}

func TestClipboardRequest(t *testing.T) {
	cb := sys.RequestClipboard(sys.PrimaryClipboard)
	if cb != "\x1b]52;p;?\x07" {
		t.Errorf("Unexpected clipboard request: %q", cb)
	}
}
