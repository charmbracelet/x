package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/exp/term/ansi"
)

func TestReset(t *testing.T) {
	var s ansi.Style
	if s.String() != "\x1b[m" {
		t.Errorf("Unexpected reset sequence: %s", ansi.ResetStyle)
	}
}

func TestBold(t *testing.T) {
	var s ansi.Style
	s = s.Bold()
	if s.String() != "\x1b[1m" {
		t.Errorf("Unexpected bold sequence: %s", s)
	}
}

func TestDefaultBackground(t *testing.T) {
	var s ansi.Style
	s = s.DefaultBackgroundColor()
	if s.String() != "\x1b[49m" {
		t.Errorf("Unexpected default background sequence: %s", s)
	}
}

func TestSequence(t *testing.T) {
	var s ansi.Style
	s = s.Bold().Underline().ForegroundColor(ansi.ExtendedColor(255))
	if s.String() != "\x1b[1;4;38;5;255m" {
		t.Errorf("Unexpected sequence: %s", s)
	}
}
