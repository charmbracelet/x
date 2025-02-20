package sixel

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"strconv"

	"github.com/bits-and-blooms/bitset"
	"github.com/soniakeys/quant"
	"github.com/soniakeys/quant/median"
)

// Sixels are a protocol for writing images to the terminal by writing a large blob of ANSI-escaped data.
// They function by encoding columns of 6 pixels into a single character (in much the same way base64
// encodes data 6 bits at a time). Sixel images are paletted, with a palette established at the beginning
// of the image blob and pixels identifying palette entires by index while writing the pixel data.
//
// Sixels are written one 6-pixel-tall band at a time, one color at a time. For each band, a single
// color's pixels are written, then a carriage return is written to bring the "cursor" back to the
// beginning of a band where a new color is selected and pixels written. This continues until the entire
// band has been drawn, at which time a line break is written to begin the next band.

// Sixel control functions.
const (
	LineBreak        byte = '-'
	CarriageReturn   byte = '$'
	RepeatIntroducer byte = '!'
	ColorIntroducer  byte = '#'
	RasterAttribute  byte = '"'

	// MaxColors is the maximum number of colors that can be used in a Sixel
	// image.
	MaxColors = 256
)

// Encoder is a Sixel encoder. It encodes an image to Sixel data format.
type Encoder struct {
	// Colors is the number of colors to use in the palette. The default is
	// 256.
	Colors int

	// Quantizer is the color quantizer to use. The default is median cut.
	Quantizer quant.Quantizer
}

// Encode will accept an Image and write sixel data to a Writer. The sixel data
// will be everything after the 'q' that ends the DCS parameters and before the ST
// that ends the sequence.  That means it includes the pixel metrics and color
// palette.
func (e *Encoder) Encode(w io.Writer, img image.Image) error {
	if img == nil {
		return nil
	}

	nc := MaxColors
	if e.Colors >= 2 {
		nc = e.Colors
	}

	var paletted *image.Paletted
	if p, ok := img.(*image.Paletted); ok && len(p.Palette) < nc {
		paletted = p
	} else {
		// make adaptive palette using median cut alogrithm
		q := e.Quantizer
		if q == nil {
			q = median.Quantizer(nc)
		}
		paletted = q.Paletted(img)
	}

	imageBounds := img.Bounds()
	imageWidth := imageBounds.Dx()
	imageHeight := imageBounds.Dy()

	// Set the default raster 1:1 aspect ratio if it's not set
	if _, err := WriteRaster(w, 1, 1, imageWidth, imageHeight); err != nil {
		return fmt.Errorf("error encoding raster: %w", err)
	}

	// Write palette colors
	for i, c := range paletted.Palette {
		c := FromColor(c)
		// Always use RGB format "2"
		if _, err := WriteColor(w, i, c.Pu, c.Px, c.Py, c.Pz); err != nil {
			return fmt.Errorf("error encoding color: %w", err)
		}
		paletted.Palette[i] = c
	}

	var pixelBands bitset.BitSet
	bandHeight := bandHeight(img)

	// Write pixel data to bitset.
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			setColor(&pixelBands, x, y, imageWidth, bandHeight, int(paletted.ColorIndexAt(x, y)))
		}
	}

	return newEncoder(w, &pixelBands).writePixelData(img, paletted)
}

// setColor will write a single pixel to the bitset data to be used by
// [encoder.writePixelData].
func setColor(bands *bitset.BitSet, x int, y int, imageWidth int, bandHeight int, paletteIndex int) {
	bandY := y / 6
	bit := bandHeight*imageWidth*6*paletteIndex + bandY*imageWidth*6 + (x * 6) + (y % 6)
	bands.Set(uint(bit)) //nolint:gosec
}

func bandHeight(img image.Image) int {
	imageHeight := img.Bounds().Dy()
	bandHeight := imageHeight / 6
	if imageHeight%6 != 0 {
		bandHeight++
	}
	return bandHeight
}

// encoder is the internal encoder used to write sixel pixel data to a writer.
type encoder struct {
	w io.Writer

	bands *bitset.BitSet

	repeatCount int
	repeatChar  byte
}

func newEncoder(w io.Writer, bands *bitset.BitSet) *encoder {
	return &encoder{
		w:     w,
		bands: bands,
	}
}

// writePixelData will write the image pixel data to the writer.
func (s *encoder) writePixelData(img image.Image, paletted *image.Paletted) error {
	imageWidth := img.Bounds().Dx()
	bandHeight := bandHeight(img)
	for bandY := 0; bandY < bandHeight; bandY++ {
		if bandY > 0 {
			s.writeControlRune(LineBreak)
		}

		hasWrittenAColor := false

		for paletteIndex := 0; paletteIndex < len(paletted.Palette); paletteIndex++ {
			c := paletted.Palette[paletteIndex]
			_, _, _, a := c.RGBA()
			if a == 0 {
				// Don't draw anything for purely transparent pixels
				continue
			}

			firstColorBit := uint(bandHeight*imageWidth*6*paletteIndex + bandY*imageWidth*6) //nolint:gosec
			nextColorBit := firstColorBit + uint(imageWidth*6)                               //nolint:gosec

			firstSetBitInBand, anySet := s.bands.NextSet(firstColorBit)
			if !anySet || firstSetBitInBand >= nextColorBit {
				// Color not appearing in this row
				continue
			}

			if hasWrittenAColor {
				s.writeControlRune(CarriageReturn)
			}
			hasWrittenAColor = true

			s.writeControlRune(ColorIntroducer)
			io.WriteString(s.w, strconv.Itoa(paletteIndex)) //nolint:errcheck

			for x := 0; x < imageWidth; x += 4 {
				bit := firstColorBit + uint(x*6) //nolint:gosec
				word := s.bands.GetWord64AtBit(bit)

				pixel1 := byte((word & 63) + '?')
				pixel2 := byte(((word >> 6) & 63) + '?')
				pixel3 := byte(((word >> 12) & 63) + '?')
				pixel4 := byte(((word >> 18) & 63) + '?')

				s.writeImageRune(pixel1)

				if x+1 >= imageWidth {
					continue
				}
				s.writeImageRune(pixel2)

				if x+2 >= imageWidth {
					continue
				}
				s.writeImageRune(pixel3)

				if x+3 >= imageWidth {
					continue
				}
				s.writeImageRune(pixel4)
			}
		}
	}

	s.writeControlRune('-')
	return nil
}

// writeImageRune will write a single line of six pixels to pixel data.  The data
// doesn't get written to the imageData, it gets buffered for the purposes of RLE
func (e *encoder) writeImageRune(r byte) { //nolint:revive
	if r == e.repeatChar {
		e.repeatCount++
		return
	}

	e.flushRepeats()
	e.repeatChar = r
	e.repeatCount = 1
}

// writeControlRune will write a special rune such as a new line or carriage return
// rune. It will call flushRepeats first, if necessary.
func (e *encoder) writeControlRune(r byte) {
	if e.repeatCount > 0 {
		e.flushRepeats()
		e.repeatCount = 0
		e.repeatChar = 0
	}

	e.w.Write([]byte{r}) //nolint:errcheck
}

// flushRepeats is used to actually write the current repeatByte to the imageData when
// it is about to change. This buffering is used to manage RLE in the sixelBuilder
func (e *encoder) flushRepeats() {
	if e.repeatCount == 0 {
		return
	}

	// Only write using the RLE form if it's actually providing space savings
	if e.repeatCount > 3 {
		WriteRepeat(e.w, e.repeatCount, e.repeatChar) //nolint:errcheck
		return
	}

	e.w.Write(bytes.Repeat([]byte{e.repeatChar}, e.repeatCount)) //nolint:errcheck
}
