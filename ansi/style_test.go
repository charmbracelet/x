package ansi_test

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestReset(t *testing.T) {
	var s ansi.Style
	if s.String() != "\x1b[m" {
		t.Errorf("Unexpected reset sequence: %q", ansi.ResetStyle)
	}
}

func TestBold(t *testing.T) {
	var s ansi.Style
	s = s.Bold()
	if s.String() != "\x1b[1m" {
		t.Errorf("Unexpected bold sequence: %q", s)
	}
}

func TestDefaultBackground(t *testing.T) {
	var s ansi.Style
	s = s.DefaultBackgroundColor()
	if s.String() != "\x1b[49m" {
		t.Errorf("Unexpected default background sequence: %q", s)
	}
}

func TestSequence(t *testing.T) {
	var s ansi.Style
	s = s.Bold().Underline().ForegroundColor(ansi.ExtendedColor(255))
	if s.String() != "\x1b[1;4;38;5;255m" {
		t.Errorf("Unexpected sequence: %q", s)
	}
}

func TestColorColor(t *testing.T) {
	var s ansi.Style
	s = s.Bold().Underline().ForegroundColor(color.Black)
	if s.String() != "\x1b[1;4;38;2;0;0;0m" {
		t.Errorf("Unexpected sequence: %q", s)
	}
}

func BenchmarkStyle(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ansi.Style{}.
			Bold().
			DoubleUnderline().
			ForegroundColor(color.RGBA{255, 255, 255, 255}).
			String()
	}
}
