package ansi_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSetForegroundColorNil(t *testing.T) {
	s := ansi.SetForegroundColor(nil)
	if s != "\x1b]10;\x07" {
		t.Errorf("Unexpected string for SetForegroundColor: got %q", s)
	}
}

func TestStringImplementations(t *testing.T) {
	foregroundColor := ansi.SetForegroundColor(ansi.BrightMagenta)
	backgroundColor := ansi.SetBackgroundColor(ansi.ExtendedColor(255))
	cursorColor := ansi.SetCursorColor(ansi.TrueColor(0xffeeaa))

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

func TestColorizer(t *testing.T) {
	hex := ansi.HexColorizer{ansi.BrightBlack}
	xrgb := ansi.XRGBColorizer{ansi.ExtendedColor(235)}
	xrgba := ansi.XRGBAColorizer{ansi.TrueColor(0x00ff00)}

	if seq := ansi.SetForegroundColor(hex); seq != "\x1b]10;#808080\x07" {
		t.Errorf("Unexpected sequence for HexColorizer: got %q", seq)
	}
	if seq := ansi.SetForegroundColor(xrgb); seq != "\x1b]10;rgb:2626/2626/2626\x07" {
		t.Errorf("Unexpected sequence for XRGBColorizer: got %q", seq)
	}
	if seq := ansi.SetForegroundColor(xrgba); seq != "\x1b]10;rgba:0000/ffff/0000/ffff\x07" {
		t.Errorf("Unexpected sequence for XRGBAColorizer: got %q", seq)
	}
}
