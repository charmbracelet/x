package ansi

import (
	"image/color"
	"testing"
)

func TestRGBAToHex(t *testing.T) {
	cases := []struct {
		r, g, b, a uint32
		want       uint32
	}{
		{0, 0, 255, 0xffff, 0x0000ff},
		{255, 255, 255, 0xffff, 0xffffff},
		{255, 0, 0, 0xffff, 0xff0000},
	}

	for _, c := range cases {
		gotR, gotG, gotB, _ := TrueColor(c.want).RGBA()
		gotR /= 256
		gotG /= 256
		gotB /= 256
		if gotR != c.r || gotG != c.g || gotB != c.b {
			t.Errorf("RGBA() of TrueColor(%06x): got (%v, %v, %v), want (%v, %v, %v)",
				c.want, gotR, gotG, gotB, c.r, c.g, c.b)
		}
	}
}

func TestColorToHexString(t *testing.T) {
	cases := []struct {
		color color.Color
		want  string
	}{
		{TrueColor(0x0000ff), "#0000ff"},
		{TrueColor(0xffffff), "#ffffff"},
		{TrueColor(0xff0000), "#ff0000"},
	}

	for _, c := range cases {
		got := colorToHexString(c.color)
		if got != c.want {
			t.Errorf("colorToHexString(%v): got %v, want %v", c.color, got, c.want)
		}
	}
}

func TestAnsiToRGB(t *testing.T) {
	cases := []struct {
		ansi    uint32
		r, g, b uint32
	}{
		{0, 0, 0, 0},         // black
		{1, 128, 0, 0},       // red
		{255, 238, 238, 238}, // highest ANSI color (grayscale)
	}

	for _, c := range cases {
		gotR, gotG, gotB := ansiToRGB(c.ansi)
		if gotR != c.r || gotG != c.g || gotB != c.b {
			t.Errorf("ansiToRGB(%v): got (%v, %v, %v), want (%v, %v, %v)",
				c.ansi, gotR, gotG, gotB, c.r, c.g, c.b)
		}
	}
}

func TestHexToRGB(t *testing.T) {
	cases := []struct {
		hex     uint32
		r, g, b uint32
	}{
		{0x0000FF, 0, 0, 255},     // blue
		{0xFFFFFF, 255, 255, 255}, // white
		{0xFF0000, 255, 0, 0},     // red
	}

	for _, c := range cases {
		gotR, gotG, gotB := hexToRGB(c.hex)
		if gotR != c.r || gotG != c.g || gotB != c.b {
			t.Errorf("hexToRGB(%v): got (%v, %v, %v), want (%v, %v, %v)",
				c.hex, gotR, gotG, gotB, c.r, c.g, c.b)
		}
	}
}
