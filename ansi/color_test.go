package ansi

import (
	"image/color"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestRGBAToHex(t *testing.T) {
	cases := []struct {
		r, g, b, a uint32
		want       uint32
	}{
		{0, 0, 255, 0xffff, 0x0000ff},
		{255, 255, 255, 0xffff, 0xffffff},
		{255, 0, 0, 0xffff, 0xffff0000},
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
		ansi    byte
		r, g, b uint32
	}{
		{0, 0, 0, 0},         // black
		{1, 128, 0, 0},       // red
		{255, 238, 238, 238}, // highest ANSI color (grayscale)
	}

	for _, c := range cases {
		gotR, gotG, gotB, _ := ansiToRGB(c.ansi).RGBA()
		// We need to shift the values down to 8 bits
		gotR >>= 8
		gotR &= 0xff
		gotG >>= 8
		gotG &= 0xff
		gotB >>= 8
		gotB &= 0xff
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

func TestHexTo256(t *testing.T) {
	testCases := map[string]struct {
		input          colorful.Color
		expectedHex    string
		expectedOutput IndexedColor
	}{
		"white": {
			input:          colorful.Color{R: 1, G: 1, B: 1},
			expectedHex:    "#ffffff",
			expectedOutput: 231,
		},
		"offwhite": {
			input:          colorful.Color{R: 0.9333, G: 0.9333, B: 0.933},
			expectedHex:    "#eeeeee",
			expectedOutput: 255,
		},
		"slightly brighter than offwhite": {
			input:          colorful.Color{R: 0.95, G: 0.95, B: 0.95},
			expectedHex:    "#f2f2f2",
			expectedOutput: 255,
		},
		"red": {
			input:          colorful.Color{R: 1, G: 0, B: 0},
			expectedHex:    "#ff0000",
			expectedOutput: 196,
		},
		"silver foil": {
			input:          colorful.Color{R: 0.6863, G: 0.6863, B: 0.6863},
			expectedHex:    "#afafaf",
			expectedOutput: 145,
		},
		"silver chalice": {
			input:          colorful.Color{R: 0.698, G: 0.698, B: 0.698},
			expectedHex:    "#b2b2b2",
			expectedOutput: 249,
		},
		"slightly closer to silver foil": {
			input:          colorful.Color{R: 0.692, G: 0.692, B: 0.692},
			expectedHex:    "#b0b0b0",
			expectedOutput: 145,
		},
		"slightly closer to silver chalice": {
			input:          colorful.Color{R: 0.694, G: 0.694, B: 0.694},
			expectedHex:    "#b1b1b1",
			expectedOutput: 249,
		},
		"gray": {
			input:          colorful.Color{R: 0.5, G: 0.5, B: 0.5},
			expectedHex:    "#808080",
			expectedOutput: 244,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// hex := fmt.Sprintf("#%02x%02x%02x", uint8(testCase.input.R*255), uint8(testCase.input.G*255), uint8(testCase.input.B*255))
			output := Convert256(testCase.input)
			if testCase.input.Hex() != testCase.expectedHex {
				t.Errorf("Expected %+v to map to %s, but instead received %s", testCase.input, testCase.expectedHex, testCase.input.Hex())
			}
			if output != testCase.expectedOutput {
				t.Errorf("Expected truecolor %+v to map to 256 color %d, but instead received %d", testCase.input, testCase.expectedOutput, output)
			}
		})
	}
}
