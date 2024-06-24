package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

var passthroughCases = []struct {
	name   string
	seq    string
	limit  int
	screen string
	tmux   string
}{
	{
		name:   "empty",
		seq:    "",
		screen: "\x1bP\x1b\\",
		tmux:   "\x1bPtmux;\x1b\\",
	},
	{
		name:   "short",
		seq:    "hello",
		screen: "\x1bPhello\x1b\\",
		tmux:   "\x1bPtmux;hello\x1b\\",
	},
	{
		name:   "limit",
		seq:    "foobarbaz",
		limit:  3,
		screen: "\x1bPfoo\x1b\\\x1bPbar\x1b\\\x1bPbaz\x1b\\",
		tmux:   "\x1bPtmux;foobarbaz\x1b\\",
	},
	{
		name:   "escaped",
		seq:    "\x1b]52;c;Zm9vYmFy\x07",
		screen: "\x1bP\x1b]52;c;Zm9vYmFy\x07\x1b\\",
		tmux:   "\x1bPtmux;\x1b\x1b]52;c;Zm9vYmFy\x07\x1b\\",
	},
}

func TestScreenPassthrough(t *testing.T) {
	for i, tt := range passthroughCases {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.ScreenPassthrough(tt.seq, tt.limit)
			if got != tt.screen {
				t.Errorf("case: %d, ScreenPassthrough() = %q, want %q", i+1, got, tt.screen)
			}
		})
	}
}

func TestTmuxPassthrough(t *testing.T) {
	for i, tt := range passthroughCases {
		t.Run(tt.name, func(t *testing.T) {
			got := ansi.TmuxPassthrough(tt.seq)
			if got != tt.tmux {
				t.Errorf("case: %d, TmuxPassthrough() = %q, want %q", i+1, got, tt.tmux)
			}
		})
	}
}
