package style_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi/style"
)

func TestReset(t *testing.T) {
	if style.Sequence() != "\x1b[m" {
		t.Errorf("Unexpected reset sequence: %s", style.ResetSequence)
	}
}

func TestBold(t *testing.T) {
	if style.Sequence(style.Bold) != "\x1b[1m" {
		t.Errorf("Unexpected bold sequence: %s", style.Sequence(style.Bold))
	}
}

func TestDefaultBackground(t *testing.T) {
	if style.Sequence(style.DefaultBackgroundColor) != "\x1b[49m" {
		t.Errorf("Unexpected default background sequence: %s", style.Sequence(style.DefaultBackgroundColor))
	}
}

func TestSequence(t *testing.T) {
	if style.Sequence(
		style.Bold,
		style.Underline,
		style.ForegrondColor(style.ExtendedColor(255)),
	) != "\x1b[1;4;38;5;255m" {
		t.Errorf("Unexpected sequence: %s", style.Sequence(style.Bold, style.Underline, style.ForegrondColor(style.ExtendedColor(255))))
	}
}
