package vt

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
)

// handleSgr handles SGR escape sequences.
// handleSgr handles Select Graphic Rendition (SGR) escape sequences.
func (t *Terminal) handleSgr() {
	pen := &t.scr.cur.Pen
	params := t.parser.Params()
	if len(params) == 0 {
		pen.Reset()
		return
	}

	for i := 0; i < len(params); i++ {
		r := params[i]
		param, hasMore := r.Param(0), r.HasMore() // Are there more subparameters i.e. separated by ":"?
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
			if hasMore { // Only accept subparameters i.e. separated by ":"
				nextParam := params[i+1].Param(0)
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
			col := t.IndexedColor(param - 30)
			pen.Foreground(col) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			if c := t.readColor(&i, params); c != nil {
				pen.Foreground(c)
			}
		case 39: // Default foreground
			pen.Foreground(nil)
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			col := t.IndexedColor(param - 40)
			pen.Background(col) //nolint:gosec
		case 48: // Set background 256 or truecolor
			if c := t.readColor(&i, params); c != nil {
				pen.Background(c)
			}
		case 49: // Default Background
			pen.Background(nil)
		case 58: // Set underline color
			if c := t.readColor(&i, params); c != nil {
				pen.UnderlineColor(c)
			}
		case 59: // Default underline color
			pen.UnderlineColor(nil)
		case 90, 91, 92, 93, 94, 95, 96, 97: // Set bright foreground
			col := t.IndexedColor(param - 90 + 8) // Bright colors start at 8
			pen.Foreground(col)                   //nolint:gosec
		case 100, 101, 102, 103, 104, 105, 106, 107: // Set bright background
			col := t.IndexedColor(param - 100 + 8) // Bright colors start at 8
			pen.Background(col)                    //nolint:gosec
		}
	}
}

func (t *Terminal) readColor(idxp *int, params []ansi.Parameter) (c ansi.Color) {
	i := *idxp
	paramsLen := len(params)
	if i > paramsLen-1 {
		return
	}
	// Note: we accept both main and subparams here
	switch param := params[i+1].Param(0); param {
	case 2: // RGB
		if i > paramsLen-4 {
			return
		}
		c = color.RGBA{
			R: uint8(params[i+2].Param(0)), //nolint:gosec
			G: uint8(params[i+3].Param(0)), //nolint:gosec
			B: uint8(params[i+4].Param(0)), //nolint:gosec
			A: 0xff,
		}
		*idxp += 4
	case 5: // 256 colors
		if i > paramsLen-2 {
			return
		}
		c = t.IndexedColor(params[i+2].Param(0))
		*idxp += 2
	}
	return
}
