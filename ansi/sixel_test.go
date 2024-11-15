package ansi

import "testing"

func TestSixelSequence(t *testing.T) {
	expectedResult := "\x1bP0;1;0q\"3;4;5;6PALETTEPIXELS\x1b\\"
	result := (SixelSequence{
		PixelWidth:    3,
		PixelHeight:   4,
		ImageWidth:    5,
		ImageHeight:   6,
		PaletteString: "PALETTE",
		PixelString:   "PIXELS",
	}).String()
	if result != expectedResult {
		t.Errorf("Expected sixel sequence to output %q but it output %q.", expectedResult, result)
		return
	}
}
