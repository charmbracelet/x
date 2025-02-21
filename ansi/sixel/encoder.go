package sixel

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/bits-and-blooms/bitset"
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
)

// Encoder is a Sixel encoder. It encodes an image to Sixel data format.
type Encoder struct {
	// NumColors is the number of colors to use in the palette. It ranges from
	// 1 to 256. Zero or less means to use the default value of 256.
	NumColors int

	// AddTransparent is a flag that indicates whether to add a transparent
	// color to the palette. The default is false.
	AddTransparent bool

	// TransparentColor is the color to use for the transparent color in the
	// palette. If nil, [color.Transparent] will be used.
	// This field is ignored if [Encoder.AddTransparent] is false.
	TransparentColor color.Color
}

// Encode will accept an Image and write sixel data to a Writer. The sixel data
// will be everything after the 'q' that ends the DCS parameters and before the ST
// that ends the sequence.  That means it includes the pixel metrics and color
// palette.
func (e *Encoder) Encode(w io.Writer, img image.Image) error {
	if img == nil {
		return nil
	}

	nc := e.NumColors
	if nc <= 0 || nc > MaxColors {
		nc = MaxColors
	}

	imageBounds := img.Bounds()

	// Set the default raster 1:1 aspect ratio if it's not set
	if _, err := WriteRaster(w, 1, 1, imageBounds.Dx(), imageBounds.Dy()); err != nil {
		return fmt.Errorf("error encoding raster: %w", err)
	}

	palette := newSixelPalette(img, MaxColors)

	for paletteIndex, color := range palette.PaletteColors {
		e.encodePaletteColor(w, paletteIndex, color)
	}

	scratch := newSixelBuilder(imageBounds.Dx(), imageBounds.Dy(), palette)

	for y := 0; y < imageBounds.Dy(); y++ {
		for x := 0; x < imageBounds.Dx(); x++ {
			scratch.SetColor(x, y, img.At(x, y))
		}
	}

	pixels := scratch.GeneratePixels()
	io.WriteString(w, pixels) //nolint:errcheck

	return nil
}

func (e *Encoder) encodePaletteColor(w io.Writer, paletteIndex int, c sixelColor) {
	// Initializing palette entries
	// #<a>;<b>;<c>;<d>;<e>
	// a = palette index
	// b = color type, 2 is RGB
	// c = R
	// d = G
	// e = B

	w.Write([]byte{ColorIntroducer})              //nolint:errcheck
	io.WriteString(w, strconv.Itoa(paletteIndex)) //nolint:errcheck
	io.WriteString(w, ";2;")
	io.WriteString(w, strconv.Itoa(int(c.Red)))   //nolint:errcheck
	w.Write([]byte{';'})                          //nolint:errcheck
	io.WriteString(w, strconv.Itoa(int(c.Green))) //nolint:errcheck
	w.Write([]byte{';'})
	io.WriteString(w, strconv.Itoa(int(c.Blue))) //nolint:errcheck
}

// sixelBuilder is a temporary structure used to create a SixelImage. It handles
// breaking pixels out into bits, and then encoding them into a sixel data string. RLE
// handling is included.
//
// Making use of a sixelBuilder is done in two phases.  First, SetColor is used to write all
// pixels to the internal BitSet data.  Then, GeneratePixels is called to retrieve a string
// representing the pixel data encoded in the sixel format.
type sixelBuilder struct {
	SixelPalette sixelPalette

	imageHeight int
	imageWidth  int

	pixelBands bitset.BitSet

	imageData   strings.Builder
	repeatByte  byte
	repeatCount int
}

// newSixelBuilder creates a sixelBuilder and prepares it for writing
func newSixelBuilder(width, height int, palette sixelPalette) sixelBuilder {
	scratch := sixelBuilder{
		imageWidth:   width,
		imageHeight:  height,
		SixelPalette: palette,
	}

	return scratch
}

// BandHeight returns the number of six-pixel bands this image consists of
func (s *sixelBuilder) BandHeight() int {
	bandHeight := s.imageHeight / 6
	if s.imageHeight%6 != 0 {
		bandHeight++
	}

	return bandHeight
}

// SetColor will write a single pixel to sixelBuilder's internal bitset data to be used by
// GeneratePixels
func (s *sixelBuilder) SetColor(x int, y int, color color.Color) {
	bandY := y / 6
	paletteIndex := s.SixelPalette.ColorIndex(sixelConvertColor(color))

	bit := s.BandHeight()*s.imageWidth*6*paletteIndex + bandY*s.imageWidth*6 + (x * 6) + (y % 6)
	s.pixelBands.Set(uint(bit)) //nolint:gosec
}

// GeneratePixels is used to write the pixel data to the internal imageData string builder.
// All pixels in the image must be written to the sixelBuilder using SetColor before this method is
// called. This method returns a string that represents the pixel data.  Sixel strings consist of five parts:
// ISC <header> <palette> <pixels> ST
// The header contains some arbitrary options indicating how the sixel image is to be drawn.
// The palette maps palette indices to RGB colors
// The pixels indicates which pixels are to be drawn with which palette colors.
//
// GeneratePixels only produces the <pixels> part of the string.  The rest is written by
// Style.RenderSixelImage.
func (s *sixelBuilder) GeneratePixels() string {
	s.imageData = strings.Builder{}
	bandHeight := s.BandHeight()

	for bandY := 0; bandY < bandHeight; bandY++ {
		if bandY > 0 {
			s.writeControlRune(LineBreak)
		}

		hasWrittenAColor := false

		for paletteIndex := 0; paletteIndex < len(s.SixelPalette.PaletteColors); paletteIndex++ {
			if s.SixelPalette.PaletteColors[paletteIndex].Alpha < 1 {
				// Don't draw anything for purely transparent pixels
				continue
			}

			firstColorBit := uint(s.BandHeight()*s.imageWidth*6*paletteIndex + bandY*s.imageWidth*6)
			nextColorBit := firstColorBit + uint(s.imageWidth*6)

			firstSetBitInBand, anySet := s.pixelBands.NextSet(firstColorBit)
			if !anySet || firstSetBitInBand >= nextColorBit {
				// Color not appearing in this row
				continue
			}

			if hasWrittenAColor {
				s.writeControlRune(CarriageReturn)
			}
			hasWrittenAColor = true

			// s.writeControlRune(ColorIntroducer)
			// s.imageData.WriteString(strconv.Itoa(paletteIndex))
			for x := 0; x < s.imageWidth; x += 4 {
				bit := firstColorBit + uint(x*6)
				word := s.pixelBands.GetWord64AtBit(bit)

				pixel1 := byte((word & 63) + '?')
				pixel2 := byte(((word >> 6) & 63) + '?')
				pixel3 := byte(((word >> 12) & 63) + '?')
				pixel4 := byte(((word >> 18) & 63) + '?')

				s.writeImageRune(pixel1)

				if x+1 >= s.imageWidth {
					continue
				}
				s.writeImageRune(pixel2)

				if x+2 >= s.imageWidth {
					continue
				}
				s.writeImageRune(pixel3)

				if x+3 >= s.imageWidth {
					continue
				}
				s.writeImageRune(pixel4)
			}
		}
	}

	s.writeControlRune('-')
	return s.imageData.String()
}

// writeImageRune will write a single line of six pixels to pixel data.  The data
// doesn't get written to the imageData, it gets buffered for the purposes of RLE
func (s *sixelBuilder) writeImageRune(r byte) {
	if r == s.repeatByte {
		s.repeatCount++
		return
	}

	s.flushRepeats()
	s.repeatByte = r
	s.repeatCount = 1
}

// writeControlRune will write a special rune such as a new line or carriage return
// rune. It will call flushRepeats first, if necessary.
func (s *sixelBuilder) writeControlRune(r byte) {
	if s.repeatCount > 0 {
		s.flushRepeats()
		s.repeatCount = 0
		s.repeatByte = 0
	}

	s.imageData.WriteByte(r)
}

// flushRepeats is used to actually write the current repeatByte to the imageData when
// it is about to change. This buffering is used to manage RLE in the sixelBuilder
func (s *sixelBuilder) flushRepeats() {
	if s.repeatCount == 0 {
		return
	}

	// Only write using the RLE form if it's actually providing space savings
	if s.repeatCount > 3 {
		WriteRepeat(&s.imageData, s.repeatCount, s.repeatByte) //nolint:errcheck
		return
	}

	for i := 0; i < s.repeatCount; i++ {
		s.imageData.WriteByte(s.repeatByte)
	}
}
