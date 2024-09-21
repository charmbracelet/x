package color

import (
	"image/color"
	"testing"
)

func TestColorToHex(t *testing.T) {
	for i, test := range []struct {
		c    color.Color
		want string
	}{
		{color.RGBA{0, 0, 0, 255}, "#000000"},
		{color.RGBA{255, 255, 255, 255}, "#FFFFFF"},
		{color.RGBA{255, 0, 0, 255}, "#FF0000"},
		{color.RGBA{0, 255, 0, 255}, "#00FF00"},
		{color.RGBA{0, 0, 255, 255}, "#0000FF"},
		{color.RGBA{107, 80, 255, 255}, "#6B50FF"},
	} {
		got := ColorToHex(test.c)
		if got != test.want {
			t.Errorf("Test %d: ColorToHex(%v)\nGot:  %v\nWant: %v", i, test.c, got, test.want)
		}
	}
}

func TestHSVToRGBA(t *testing.T) {
	for i, test := range []struct {
		h, s, v float64
		want    color.RGBA
	}{
		{0, 0, 0, color.RGBA{0, 0, 0, 255}},
		{0, 0, 1, color.RGBA{255, 255, 255, 255}},
		{0, 1, 1, color.RGBA{255, 0, 0, 255}},
		{120, 1, 1, color.RGBA{0, 255, 0, 255}},
		{240, 1, 1, color.RGBA{0, 0, 255, 255}},
		{249, 0.69, 1, color.RGBA{105, 79, 255, 255}},
	} {
		got := HSVToRGBA(test.h, test.s, test.v)
		if got != test.want {
			t.Errorf("Test %d: HSVToRGBA(%v, %v, %v)\nGot:  %v\nWant: %v", i, test.h, test.s, test.v, got, test.want)
		}
	}
}

func TestColorToHSV(t *testing.T) {
	for i, test := range []struct {
		c    color.Color
		want struct {
			h, s, v float64
		}
	}{
		{color.RGBA{0, 0, 0, 255}, struct{ h, s, v float64 }{0, 0, 0}},
		{color.RGBA{255, 255, 255, 255}, struct{ h, s, v float64 }{0, 0, 1}},
		{color.RGBA{255, 0, 0, 255}, struct{ h, s, v float64 }{0, 1, 1}},
		{color.RGBA{0, 255, 0, 255}, struct{ h, s, v float64 }{120, 1, 1}},
		{color.RGBA{0, 0, 255, 255}, struct{ h, s, v float64 }{240, 1, 1}},
		{color.RGBA{105, 79, 255, 255}, struct{ h, s, v float64 }{249, 0.69, 1}},
	} {
		h, s, v := ColorToHSV(test.c)
		if h != test.want.h || s != test.want.s || v != test.want.v {
			t.Errorf("Test %d: ColorToHSV(%v)\nGot:  %v\nWant: %v", i, test.c, struct{ h, s, v float64 }{h, s, v}, test.want)
		}
	}
}
