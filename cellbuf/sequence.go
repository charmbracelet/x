package cellbuf

import (
	"bytes"
	"image/color"

	"github.com/charmbracelet/x/ansi"
)

// ReadStyle reads a Select Graphic Rendition (SGR) escape sequences from a
// list of parameters.
func ReadStyle(params ansi.Params, pen *Style) {
	if len(params) == 0 {
		pen.Reset()
		return
	}

	for i := 0; i < len(params); i++ {
		param, hasMore, _ := params.Param(i, 0)
		switch param {
		case 0: // Reset
			pen.Reset()
		case 1: // Bold
			pen.Bold(true)
		case 2: // Dim/Faint
			pen.Faint(true)
		case 3: // Italic
			pen.Italic(true)
		case 4: // Underline
			nextParam, _, ok := params.Param(i+1, 0)
			if hasMore && ok { // Only accept subparameters i.e. separated by ":"
				switch nextParam {
				case 0, 1, 2, 3, 4, 5:
					i++
					switch nextParam {
					case 0: // No Underline
						pen.UnderlineStyle(NoUnderline)
					case 1: // Single Underline
						pen.UnderlineStyle(SingleUnderline)
					case 2: // Double Underline
						pen.UnderlineStyle(DoubleUnderline)
					case 3: // Curly Underline
						pen.UnderlineStyle(CurlyUnderline)
					case 4: // Dotted Underline
						pen.UnderlineStyle(DottedUnderline)
					case 5: // Dashed Underline
						pen.UnderlineStyle(DashedUnderline)
					}
				}
			} else {
				// Single Underline
				pen.Underline(true)
			}
		case 5: // Slow Blink
			pen.SlowBlink(true)
		case 6: // Rapid Blink
			pen.RapidBlink(true)
		case 7: // Reverse
			pen.Reverse(true)
		case 8: // Conceal
			pen.Conceal(true)
		case 9: // Crossed-out/Strikethrough
			pen.Strikethrough(true)
		case 22: // Normal Intensity (not bold or faint)
			pen.Bold(false).Faint(false)
		case 23: // Not italic, not Fraktur
			pen.Italic(false)
		case 24: // Not underlined
			pen.Underline(false)
		case 25: // Blink off
			pen.SlowBlink(false).RapidBlink(false)
		case 27: // Positive (not reverse)
			pen.Reverse(false)
		case 28: // Reveal
			pen.Conceal(false)
		case 29: // Not crossed out
			pen.Strikethrough(false)
		case 30, 31, 32, 33, 34, 35, 36, 37: // Set foreground
			pen.Foreground(ansi.Black + ansi.BasicColor(param-30)) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			var c color.Color
			n := ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Foreground(c)
				i += n - 1
			}
		case 39: // Default foreground
			pen.Foreground(nil)
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			pen.Background(ansi.Black + ansi.BasicColor(param-40)) //nolint:gosec
		case 48: // Set background 256 or truecolor
			var c color.Color
			n := ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Background(c)
				i += n - 1
			}
		case 49: // Default Background
			pen.Background(nil)
		case 58: // Set underline color
			var c color.Color
			n := ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.UnderlineColor(c)
				i += n - 1
			}
		case 59: // Default underline color
			pen.UnderlineColor(nil)
		case 90, 91, 92, 93, 94, 95, 96, 97: // Set bright foreground
			pen.Foreground(ansi.BrightBlack + ansi.BasicColor(param-90)) //nolint:gosec
		case 100, 101, 102, 103, 104, 105, 106, 107: // Set bright background
			pen.Background(ansi.BrightBlack + ansi.BasicColor(param-100)) //nolint:gosec
		}
	}
}

// ReadLink reads a hyperlink escape sequence from a data buffer.
func ReadLink(p []byte, link *Link) {
	params := bytes.Split(p, []byte{';'})
	if len(params) != 3 {
		return
	}
	for _, param := range bytes.Split(params[1], []byte{':'}) {
		if bytes.HasPrefix(param, []byte("id=")) {
			link.URLID = string(param)
		}
	}
	link.URL = string(params[2])
}

// ReadStyleColor decodes a color from a slice of parameters. It returns the
// number of parameters read and the color. This function is used to read SGR
// color parameters following the ITU T.416 standard.
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
//  1. Support for legacy color values separated by semicolons (;) with respect to RGB, and indexed colors
//  2. Support ignoring and omitting the color space id (second parameter) with respect to RGB colors
//  3. Support ignoring and omitting the 6th parameter with respect to RGB and CMY colors
//  4. Support reading RGBA colors
func ReadStyleColor(params ansi.Params, co *color.Color) (n int) {
	if len(params) < 2 { // Need at least SGR type and color type
		return 0
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
		return 2
	case 1: // transparent
		*co = color.Transparent
		return 2
	case 2: // RGB direct color
		if len(params) < 5 {
			return 0
		}

		r, g, b, _ := paramsfn()
		if r == -1 || g == -1 || b == -1 {
			return 0
		}

		*co = color.RGBA{
			R: uint8(r), //nolint:gosec
			G: uint8(g), //nolint:gosec
			B: uint8(b), //nolint:gosec
			A: 0xff,
		}
		return

	case 3: // CMY direct color
		if len(params) < 5 {
			return 0
		}

		c, m, y, _ := paramsfn()
		if c == -1 || m == -1 || y == -1 {
			return 0
		}

		*co = color.CMYK{
			C: uint8(c), //nolint:gosec
			M: uint8(m), //nolint:gosec
			Y: uint8(y), //nolint:gosec
			K: 0,
		}
		return

	case 4: // CMYK direct color
		if len(params) < 6 {
			return 0
		}

		c, m, y, k := paramsfn()
		if c == -1 || m == -1 || y == -1 || k == -1 {
			return 0
		}

		*co = color.CMYK{
			C: uint8(c), //nolint:gosec
			M: uint8(m), //nolint:gosec
			Y: uint8(y), //nolint:gosec
			K: uint8(k), //nolint:gosec
		}
		return

	case 5: // indexed color
		if len(params) < 3 {
			return 0
		}
		switch {
		case s.HasMore() && p.HasMore() && !params[2].HasMore():
			// Colon separated indexed color
			// 38 : 5 : 234
		case !s.HasMore() && !p.HasMore() && !params[2].HasMore():
			// Legacy semicolon indexed color
			// 38 ; 5 ; 234
		default:
			return 0
		}
		*co = ansi.ExtendedColor(params[2].Param(0)) //nolint:gosec
		return 3

	case 6: // RGBA direct color
		if len(params) < 6 {
			return 0
		}

		r, g, b, a := paramsfn()
		if r == -1 || g == -1 || b == -1 || a == -1 {
			return 0
		}

		*co = color.RGBA{
			R: uint8(r), //nolint:gosec
			G: uint8(g), //nolint:gosec
			B: uint8(b), //nolint:gosec
			A: uint8(a), //nolint:gosec
		}
		return

	default:
		return 0
	}
}
