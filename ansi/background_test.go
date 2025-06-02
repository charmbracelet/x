package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

func TestSetForegroundColorNil(t *testing.T) {
	s := ansi.SetForegroundColor("")
	if s != "\x1b]10;\x07" {
		t.Errorf("Unexpected string for SetForegroundColor: got %q", s)
	}
}

func TestStringImplementations(t *testing.T) {
	brightMagenta, ok := colorful.MakeColor(ansi.BrightMagenta)
	if !ok {
		t.Fatalf("Failed to create color for BrightMagenta: %v", ansi.BrightMagenta)
	}
	color255, ok := colorful.MakeColor(ansi.ExtendedColor(255))
	if !ok {
		t.Fatalf("Failed to create color for ExtendedColor(255): %v", ansi.ExtendedColor(255))
	}
	hexColor := ansi.HexColor("#ffeeaa")
	foregroundColor := ansi.SetForegroundColor(brightMagenta.Hex())
	backgroundColor := ansi.SetBackgroundColor(color255.Hex())
	cursorColor := ansi.SetCursorColor(hexColor.Hex())

	if foregroundColor != "\x1b]10;#ff00ff\x07" {
		t.Errorf("Unexpected string for SetForegroundColor: got %q",
			foregroundColor)
	}
	if backgroundColor != "\x1b]11;#eeeeee\x07" {
		t.Errorf("Unexpected string for SetBackgroundColor: got %q",
			backgroundColor)
	}
	if cursorColor != "\x1b]12;#ffeeaa\x07" {
		t.Errorf("Unexpected string for SetCursorColor: got %q",
			cursorColor)
	}
}
