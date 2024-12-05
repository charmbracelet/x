package ansi

import (
	"bytes"
	"strconv"
)

type SixelSequence struct {
	PixelWidth    int
	PixelHeight   int
	ImageWidth    int
	ImageHeight   int
	PaletteString string
	PixelString   string
}

var _ Sequence = &SixelSequence{}

// Clone returns a copy of the DCS sequence.
func (s SixelSequence) Clone() Sequence {
	return SixelSequence{
		PixelWidth:    s.PixelWidth,
		PixelHeight:   s.PixelHeight,
		ImageWidth:    s.ImageWidth,
		ImageHeight:   s.ImageHeight,
		PaletteString: s.PaletteString,
		PixelString:   s.PixelString,
	}
}

// String returns a string representation of the sequence.
// The string will always be in the 7-bit format i.e (ESC P p..p i..i f <data> ESC \).
func (s SixelSequence) String() string {
	return s.buffer().String()
}

// Bytes returns the byte representation of the sequence.
// The bytes will always be in the 7-bit format i.e (ESC P p..p i..i F <data> ESC \).
func (s SixelSequence) Bytes() []byte {
	return s.buffer().Bytes()
}

func (s SixelSequence) buffer() *bytes.Buffer {
	var b bytes.Buffer

	b.WriteByte(ESC)
	// P<a>;<b>;<c>q
	// a = pixel aspect ratio (deprecated)
	// b = how to color unfilled pixels, 1 = transparent
	//     (0 means background color but terminals seem to draw it arbitrarily making it useless)
	// c = horizontal grid size, I think everyone ignores this
	b.WriteString("P0;1;0q")
	// "<a>;<b>;<c>;<d>
	// a = pixel width
	// b = pixel height
	// c = image width in pixels
	// d = image height in pixels
	b.WriteByte('"')
	b.WriteString(strconv.Itoa(s.PixelWidth))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(s.PixelHeight))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(s.ImageWidth))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(s.ImageHeight))
	b.WriteString(s.PaletteString)
	b.WriteString(s.PixelString)
	// ST
	b.WriteByte(ESC)
	b.WriteByte('\\')

	return &b
}
