package ansi

import (
	"image/color"
)

// Technically speaking, the 16 basic ANSI colors are arbitrary and can be
// customized at the terminal level. Given that, we're returning what we feel
// are good defaults.
//
// This could also be a slice, but we use a map to make the mappings very
// explicit.
//
// See: https://www.ditig.com/publications/256-colors-cheat-sheet
var lowANSI = map[uint32]uint32{
	0:  0x000000, // black
	1:  0x800000, // red
	2:  0x008000, // green
	3:  0x808000, // yellow
	4:  0x000080, // blue
	5:  0x800080, // magenta
	6:  0x008080, // cyan
	7:  0xc0c0c0, // white
	8:  0x808080, // bright black
	9:  0xff0000, // bright red
	10: 0x00ff00, // bright green
	11: 0xffff00, // bright yellow
	12: 0x0000ff, // bright blue
	13: 0xff00ff, // bright magenta
	14: 0x00ffff, // bright cyan
	15: 0xffffff, // bright white
}

// Color is a color that can be used in a terminal. ANSI (including
// ANSI256) and 24-bit "true colors" fall under this category.
type Color interface {
	color.Color
}

// BasicColor is an ANSI 3-bit or 4-bit color with a value from 0 to 15.
type BasicColor uint8

var _ Color = BasicColor(0)

const (
	// Black is the ANSI black color.
	Black BasicColor = iota

	// Red is the ANSI red color.
	Red

	// Green is the ANSI green color.
	Green

	// Yellow is the ANSI yellow color.
	Yellow

	// Blue is the ANSI blue color.
	Blue

	// Magenta is the ANSI magenta color.
	Magenta

	// Cyan is the ANSI cyan color.
	Cyan

	// White is the ANSI white color.
	White

	// BrightBlack is the ANSI bright black color.
	BrightBlack

	// BrightRed is the ANSI bright red color.
	BrightRed

	// BrightGreen is the ANSI bright green color.
	BrightGreen

	// BrightYellow is the ANSI bright yellow color.
	BrightYellow

	// BrightBlue is the ANSI bright blue color.
	BrightBlue

	// BrightMagenta is the ANSI bright magenta color.
	BrightMagenta

	// BrightCyan is the ANSI bright cyan color.
	BrightCyan

	// BrightWhite is the ANSI bright white color.
	BrightWhite
)

// RGBA returns the red, green, blue and alpha components of the color. It
// satisfies the color.Color interface.
func (c BasicColor) RGBA() (uint32, uint32, uint32, uint32) {
	ansi := uint32(c)
	if ansi > 15 {
		return 0, 0, 0, 0xffff
	}

	r, g, b := ansiToRGB(ansi)
	return toRGBA(r, g, b)
}

// ExtendedColor is an ANSI 256 (8-bit) color with a value from 0 to 255.
type ExtendedColor uint8

var _ Color = ExtendedColor(0)

// RGBA returns the red, green, blue and alpha components of the color. It
// satisfies the color.Color interface.
func (c ExtendedColor) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b := ansiToRGB(uint32(c))
	return toRGBA(r, g, b)
}

// TrueColor is a 24-bit color that can be used in the terminal.
// This can be used to represent RGB colors.
//
// For example, the color red can be represented as:
//
//	TrueColor(0xff0000)
type TrueColor uint32

var _ Color = TrueColor(0)

// RGBA returns the red, green, blue and alpha components of the color. It
// satisfies the color.Color interface.
func (c TrueColor) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b := hexToRGB(uint32(c))
	return toRGBA(r, g, b)
}

// ansiToRGB converts an ANSI color to a 24-bit RGB color.
//
//	r, g, b := ansiToRGB(57)
func ansiToRGB(ansi uint32) (uint32, uint32, uint32) {
	// For out-of-range values return black.
	if ansi > 255 {
		return 0, 0, 0
	}

	// Low ANSI.
	if ansi < 16 {
		h, ok := lowANSI[ansi]
		if !ok {
			return 0, 0, 0
		}
		r, g, b := hexToRGB(h)
		return r, g, b
	}

	// Grays.
	if ansi > 231 {
		s := (ansi-232)*10 + 8
		return s, s, s
	}

	// ANSI256.
	n := ansi - 16
	b := n % 6
	g := (n - b) / 6 % 6
	r := (n - b - g*6) / 36 % 6
	for _, v := range []*uint32{&r, &g, &b} {
		if *v > 0 {
			c := *v*40 + 55
			*v = c
		}
	}

	return r, g, b
}

// hexToRGB converts a number in hexadecimal format to red, green, and blue
// values.
//
//	r, g, b := hexToRGB(0x0000FF)
func hexToRGB(hex uint32) (uint32, uint32, uint32) {
	return hex >> 16 & 0xff, hex >> 8 & 0xff, hex & 0xff
}

// toRGBA converts an RGB 8-bit color values to 32-bit color values suitable
// for color.Color.
//
// color.Color requires 16-bit color values, so we duplicate the 8-bit values
// to fill the 16-bit values.
//
// This always returns 0xffff (opaque) for the alpha channel.
func toRGBA(r, g, b uint32) (uint32, uint32, uint32, uint32) {
	r |= r << 8
	g |= g << 8
	b |= b << 8
	return r, g, b, 0xffff
}

// ReadColor reads a color from a slice of parameters. It returns the number of
// parameters read and the color. This function is used to read SGR color
// parameters following the ITU T.416 standard.
//
// It supports reading the following color types:
//   - 0: implementation defined
//   - 1: transparent
//   - 2: RGB direct color
//   - 3: CMY direct color
//   - 4: CMYK direct color
//   - 5: indexed color
//   - 6: RGBA direct color (WezTerm extension)
//
// The parameters can be separated by semicolons (;) or colons (:). Mixing
// separators is not allowed.
//
// The specs supports defining a color space id, a color tolerance value, and a
// tolerance color space id. However, these values have no effect on the
// returned color and will be ignored.
//
// This implementation includes a few modifications to the specs:
//  1. Support for legacy color values separated by semicolons (;) with respect to RGB, CMY, CMYK, and indexed colors
//  2. Support ignoring and omitting the color space id (second parameter)
//  3. Support ignoring and omitting the 6th parameter with respect to RGB and CMY colors
//  4. Support reading RGBA colors
func ReadColor(params []Parameter) (n int, co Color) {
	if len(params) < 2 { // Need at least SGR type and color type
		return 0, nil
	}

	// First parameter indicates one of 38, 48, or 58 (foreground, background, or underline)
	s := params[0]
	p := params[1]
	colorType := p.Param(0)
	n = 2

	paramsfn := func() (p1, p2, p3, p4 int) {
		// Where should we start reading the color?
		switch {
		case s.HasMore() && p.HasMore() && len(params) > 8 && params[2].HasMore() && params[3].HasMore() && params[4].HasMore() && params[5].HasMore() && params[6].HasMore() && params[7].HasMore():
			// We have color space id, a 6th parameter, a tolerance value, and a tolerance color space
			n += 7
			return params[3].Param(0), params[4].Param(0), params[5].Param(0), params[6].Param(0)
		case s.HasMore() && p.HasMore() && len(params) > 7 && params[2].HasMore() && params[3].HasMore() && params[4].HasMore() && params[5].HasMore() && params[6].HasMore():
			// We have color space id, a 6th parameter, and a tolerance value
			n += 6
			return params[3].Param(0), params[4].Param(0), params[5].Param(0), params[6].Param(0)
		case s.HasMore() && p.HasMore() && len(params) > 6 && params[2].HasMore() && params[3].HasMore() && params[4].HasMore() && params[5].HasMore():
			// We have color space id and a 6th parameter
			// 48 : 4 : : 1 : 2 : 3 :4
			n += 5
			return params[3].Param(0), params[4].Param(0), params[5].Param(0), params[6].Param(0)
		case s.HasMore() && p.HasMore() && len(params) > 5 && params[2].HasMore() && params[3].HasMore() && params[4].HasMore() && !params[5].HasMore():
			// We have color space
			// 48 : 3 : : 1 : 2 : 3
			n += 4
			return params[3].Param(0), params[4].Param(0), params[5].Param(0), -1
		case s.HasMore() && p.HasMore() && p.Param(0) == 2 && params[2].HasMore() && params[3].HasMore() && !params[4].HasMore():
			// We have color values separated by colons (:)
			// 48 : 2 : 1 : 2 : 3
			fallthrough
		case !s.HasMore() && !p.HasMore() && p.Param(0) == 2 && !params[2].HasMore() && !params[3].HasMore() && !params[4].HasMore():
			// Support legacy color values separated by semicolons (;)
			// 48 ; 2 ; 1 ; 2 ; 3
			n += 3
			return params[2].Param(0), params[3].Param(0), params[4].Param(0), -1
		}
		// Ambiguous SGR color
		return -1, -1, -1, -1
	}

	switch colorType {
	case 0: // implementation defined
		return 2, nil
	case 1: // transparent
		return 2, color.Transparent
	case 2: // RGB direct color
		if len(params) < 5 {
			return 0, nil
		}

		r, g, b, _ := paramsfn()
		if r == -1 || g == -1 || b == -1 {
			return 0, nil
		}

		co = color.RGBA{
			R: uint8(r), //nolint:gosec
			G: uint8(g), //nolint:gosec
			B: uint8(b), //nolint:gosec
			A: 0xff,
		}
		return

	case 3: // CMY direct color
		if len(params) < 5 {
			return 0, nil
		}

		c, m, y, _ := paramsfn()
		if c == -1 || m == -1 || y == -1 {
			return 0, nil
		}

		co = color.CMYK{
			C: uint8(c), //nolint:gosec
			M: uint8(m), //nolint:gosec
			Y: uint8(y), //nolint:gosec
			K: 0,
		}
		return

	case 4: // CMYK direct color
		if len(params) < 6 {
			return 0, nil
		}

		c, m, y, k := paramsfn()
		if c == -1 || m == -1 || y == -1 || k == -1 {
			return 0, nil
		}

		co = color.CMYK{
			C: uint8(c), //nolint:gosec
			M: uint8(m), //nolint:gosec
			Y: uint8(y), //nolint:gosec
			K: uint8(k), //nolint:gosec
		}
		return

	case 5: // indexed color
		if len(params) < 3 {
			return 0, nil
		}
		switch {
		case s.HasMore() && p.HasMore() && !params[2].HasMore():
			// Colon separated indexed color
			// 38 : 5 : 234
		case !s.HasMore() && !p.HasMore() && !params[2].HasMore():
			// Legacy semicolon indexed color
			// 38 ; 5 ; 234
		default:
			return 0, nil
		}
		co = ExtendedColor(params[2].Param(0)) //nolint:gosec
		return 3, co

	case 6: // RGBA direct color
		if len(params) < 6 {
			return 0, nil
		}

		r, g, b, a := paramsfn()
		if r == -1 || g == -1 || b == -1 || a == -1 {
			return 0, nil
		}

		co = color.RGBA{
			R: uint8(r), //nolint:gosec
			G: uint8(g), //nolint:gosec
			B: uint8(b), //nolint:gosec
			A: uint8(a), //nolint:gosec
		}
		return

	default:
		return 0, nil
	}
}
