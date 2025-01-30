package sixel

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestFullImage(t *testing.T) {
	testCases := map[string]struct {
		imageWidth  int
		imageHeight int
		bandCount   int
		// When filling the image, we'll use a map of indices to colors and change colors every
		// time the current index is in the map- this will prevent dozens of lines with the same color
		// in a row and make this slightly more legible
		colors map[int]color.RGBA
	}{
		"3x12 single color filled": {
			3, 12, 2,
			map[int]color.RGBA{
				0: {255, 0, 0, 255},
			},
		},
		"3x12 two color filled": {
			3, 12, 2,
			map[int]color.RGBA{
				// 3-pixel high alternating bands
				0:  {0, 0, 255, 255},
				9:  {0, 255, 0, 255},
				18: {0, 0, 255, 255},
				27: {0, 255, 0, 255},
			},
		},
		"3x12 8 color with right gutter": {
			3, 12, 2,
			map[int]color.RGBA{
				0:  {255, 0, 0, 255},
				2:  {0, 255, 0, 255},
				3:  {255, 0, 0, 255},
				5:  {0, 255, 0, 255},
				6:  {255, 0, 0, 255},
				8:  {0, 255, 0, 255},
				9:  {0, 0, 255, 255},
				11: {128, 128, 0, 255},
				12: {0, 0, 255, 255},
				14: {128, 128, 0, 255},
				15: {0, 0, 255, 255},
				17: {128, 128, 0, 255},
				18: {0, 128, 128, 255},
				20: {128, 0, 128, 255},
				21: {0, 128, 128, 255},
				23: {128, 0, 128, 255},
				24: {0, 128, 128, 255},
				26: {128, 0, 128, 255},
				27: {64, 0, 0, 255},
				29: {0, 64, 0, 255},
				30: {64, 0, 0, 255},
				32: {0, 64, 0, 255},
				33: {64, 0, 0, 255},
				35: {0, 64, 0, 255},
			},
		},
		"3x12 single color with transparent band in the middle": {
			3, 12, 2,
			map[int]color.RGBA{
				0:  {255, 0, 0, 255},
				15: {0, 0, 0, 0},
				21: {255, 0, 0, 255},
			},
		},
		"3x5 single color": {
			3, 5, 1,
			map[int]color.RGBA{
				0: {255, 0, 0, 255},
			},
		},
		"12x4 single color use RLE": {
			12, 4, 1,
			map[int]color.RGBA{
				0: {255, 0, 0, 255},
			},
		},
		"12x1 two color use RLE": {
			12, 1, 1,
			map[int]color.RGBA{
				0: {255, 0, 0, 255},
				6: {0, 255, 0, 255},
			},
		},
		"12x12 single color use RLE": {
			12, 12, 2,
			map[int]color.RGBA{
				0: {255, 0, 0, 255},
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, testCase.imageWidth, testCase.imageHeight))

			currentColor := color.RGBA{0, 0, 0, 0}
			for y := 0; y < testCase.imageHeight; y++ {
				for x := 0; x < testCase.imageWidth; x++ {
					index := y*testCase.imageWidth + x
					newColor, changingColor := testCase.colors[index]
					if changingColor {
						currentColor = newColor
					}

					img.Set(x, y, currentColor)
				}
			}

			buffer := bytes.NewBuffer(nil)
			encoder := Encoder{}
			decoder := Decoder{}

			err := encoder.Encode(buffer, img)
			if err != nil {
				t.Errorf("Unexpected error: %+v", err)
				return
			}

			compareImg, err := decoder.Decode(buffer.Bytes())
			if err != nil {
				t.Errorf("Unexpected error: %+v", err)
				return
			}

			expectedWidth := img.Bounds().Dx()
			expectedHeight := img.Bounds().Dy()
			actualWidth := compareImg.Bounds().Dx()
			actualHeight := compareImg.Bounds().Dy()

			if actualHeight != expectedHeight {
				t.Errorf("SixelImage had a height of %d, but a height of %d was expected", actualHeight, expectedHeight)
				return
			}
			if actualWidth != expectedWidth {
				t.Errorf("SixelImage had a width of %d, but a width of %d was expected", actualWidth, expectedWidth)
				return
			}

			for y := 0; y < expectedHeight; y++ {
				for x := 0; x < expectedWidth; x++ {
					r, g, b, a := compareImg.At(x, y).RGBA()
					expectedR, expectedG, expectedB, expectedA := img.At(x, y).RGBA()

					if r != expectedR || g != expectedG || b != expectedB || a != expectedA {
						t.Errorf("SixelImage had color (%d,%d,%d,%d) at coordinates (%d,%d), but color (%d,%d,%d,%d) was expected",
							r, g, b, a, x, y, expectedR, expectedG, expectedB, expectedA)
						return
					}
				}
			}
		})
	}
}
