package sixel

import (
	"fmt"
	"image/color"
	"io"

	"github.com/lucasb-eyer/go-colorful"
)

// ErrInvalidColor is returned when a Sixel color is invalid.
var ErrInvalidColor = fmt.Errorf("invalid color")

// WriteColor writes a Sixel color to a writer. If pu is 0, the rest of the
// parameters are ignored.
func WriteColor(w io.Writer, pc, pu, px, py, pz int) (int, error) {
	if pu <= 0 || pu > 2 {
		return fmt.Fprintf(w, "#%d", pc)
	}

	return fmt.Fprintf(w, "#%d;%d;%d;%d;%d", pc, pu, px, py, pz)
}

// ConvertChannel converts a color channel from color.Color 0xffff to 0-100
// Sixel RGB format.
func ConvertChannel(c uint32) uint32 {
	// We add 328 because that is about 0.5 in the sixel 0-100 color range, we're trying to
	// round to the nearest value
	return (c + 328) * 100 / 0xffff
}

// FromColor returns a Sixel color from a color.Color. It converts the color
// channels to the 0-100 range.
func FromColor(c color.Color) Color {
	if c == nil {
		return Color{}
	}

	r, g, b, _ := c.RGBA()
	return Color{
		Pu: 2, // Always use RGB format "2"
		Px: int(ConvertChannel(r)),
		Py: int(ConvertChannel(g)),
		Pz: int(ConvertChannel(b)),
	}
}

// DecodeColor decodes a Sixel color from a byte slice. It returns the Color and
// the number of bytes read.
func DecodeColor(data []byte) (c Color, n int) {
	if len(data) == 0 || data[0] != ColorIntroducer {
		return
	}

	if len(data) < 2 { // The minimum length is 2: the introducer and a digit.
		return
	}

	// Parse the color number and optional color system.
	pc := &c.Pc
	for n = 1; n < len(data); n++ {
		if data[n] == ';' {
			if pc == &c.Pc {
				pc = &c.Pu
			} else {
				n++
				break
			}
		} else if data[n] >= '0' && data[n] <= '9' {
			*pc = (*pc)*10 + int(data[n]-'0')
		} else {
			break
		}
	}

	// Parse the color components.
	ptr := &c.Px
	for ; n < len(data); n++ {
		if data[n] == ';' {
			if ptr == &c.Px {
				ptr = &c.Py
			} else if ptr == &c.Py {
				ptr = &c.Pz
			} else {
				n++
				break
			}
		} else if data[n] >= '0' && data[n] <= '9' {
			*ptr = (*ptr)*10 + int(data[n]-'0')
		} else {
			break
		}
	}

	return
}

// Color represents a Sixel color.
type Color struct {
	// Pc is the color number (0-255).
	Pc int
	// Pu is an optional color system
	//  - 0: default color map
	//  - 1: HLS
	//  - 2: RGB
	Pu int
	// Color components range from 0-100 for RGB values. For HLS format, the Px
	// (Hue) component ranges from 0-360 degrees while L (Lightness) and S
	// (Saturation) are 0-100.
	Px, Py, Pz int
}

// RGBA implements the color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	switch c.Pu {
	case 1:
		return sixelHLS(c.Px, c.Py, c.Pz).RGBA()
	case 2:
		return sixelRGB(c.Px, c.Py, c.Pz).RGBA()
	default:
		return colorPalette[c.Pc].RGBA()
	}
}

// #define PALVAL(n,a,m) (((n) * (a) + ((m) / 2)) / (m))
func palval(n, a, m int) int {
	return (n*a + m/2) / m
}

func sixelRGB(r, g, b int) color.Color {
	return color.NRGBA{uint8(palval(r, 0xff, 100)), uint8(palval(g, 0xff, 100)), uint8(palval(b, 0xff, 100)), 0xFF} //nolint:gosec
}

func sixelHLS(h, l, s int) color.Color {
	return colorful.Hsl(float64(h), float64(s)/100.0, float64(l)/100.0).Clamped()
}
