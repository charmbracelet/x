package ansi

import (
	"image/color"
	"testing"
)

func TestSetPalette(t *testing.T) {
	cases := []struct {
		index int
		color color.Color
		want  string
	}{
		{-1, color.RGBA{255, 0, 0, 255}, ""},
		{0, nil, ""},
		{0, color.RGBA{255, 0, 0, 255}, "\x1b]P0ff0000\x07"},
		{1, color.RGBA{0, 255, 0, 255}, "\x1b]P100ff00\x07"},
		{2, color.RGBA{0, 0, 255, 255}, "\x1b]P20000ff\x07"},
		{3, color.RGBA{255, 255, 0, 255}, "\x1b]P3ffff00\x07"},
		{4, color.RGBA{255, 0, 255, 255}, "\x1b]P4ff00ff\x07"},
		{5, color.RGBA{0, 255, 255, 255}, "\x1b]P500ffff\x07"},
		{6, color.RGBA{192, 192, 192, 255}, "\x1b]P6c0c0c0\x07"},
		{7, color.RGBA{128, 128, 128, 255}, "\x1b]P7808080\x07"},
		{8, color.RGBA{255, 128, 128, 255}, "\x1b]P8ff8080\x07"},
		{9, color.RGBA{128, 255, 128, 255}, "\x1b]P980ff80\x07"},
		{10, color.RGBA{128, 128, 255, 255}, "\x1b]Pa8080ff\x07"},
		{11, color.RGBA{255, 255, 128, 255}, "\x1b]Pbffff80\x07"},
		{12, color.RGBA{255, 128, 255, 255}, "\x1b]Pcff80ff\x07"},
		{13, color.RGBA{128, 255, 255, 255}, "\x1b]Pd80ffff\x07"},
		{14, color.RGBA{192, 192, 192, 255}, "\x1b]Pec0c0c0\x07"},
		{15, color.RGBA{0, 0, 0, 255}, "\x1b]Pf000000\x07"},
		{16, color.RGBA{255, 0, 0, 255}, ""},
		{256, nil, ""},
	}
	for _, c := range cases {
		got := SetPalette(c.index, c.color)
		if got != c.want {
			t.Errorf("SetPalette(%d, %v) = %q; want %q", c.index, c.color, got, c.want)
		}
	}
}
