package cellbuf

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/wcwidth"
)

// setContent writes the given data to the buffer starting from the first cell.
// It accepts both string and []byte data types.
func setContent(
	dis Screen,
	data string,
	method Method,
) []int {
	var cell Cell
	var pen Style
	var link Link
	var x, y int

	p := ansi.GetParser()
	defer ansi.PutParser(p)
	data = strings.ReplaceAll(data, "\r\n", "\n")

	// linew is a slice of line widths. We use this to keep track of the
	// written widths of each line. We use this information later to optimize
	// rendering of the buffer.
	linew := make([]int, dis.Height())

	var pendingWidth int

	var state byte
	for len(data) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(data, state, p)

		switch width {
		case 2, 3, 4: // wide cells can go up to 4 cells wide

			switch method {
			case WcWidth:
				if r, rw := utf8.DecodeRuneInString(data); r != utf8.RuneError {
					n = rw
					width = wcwidth.RuneWidth(r)
					seq = string(r)
					newState = 0
				}
			case GraphemeWidth:
				// [ansi.DecodeSequence] already handles grapheme clusters
			}
			fallthrough
		case 1:
			cell.Content = seq
			cell.Width = width
			cell.Style = pen
			cell.Link = link

			dis.SetCell(x, y, cell) //nolint:errcheck

			// Advance the cursor and line width
			x += cell.Width
			if cell.Equal(spaceCell) {
				pendingWidth += cell.Width
			} else if y < len(linew) {
				linew[y] += cell.Width + pendingWidth
				pendingWidth = 0
			}

			cell.Reset()
		default:
			// Valid sequences always have a non-zero Cmd.
			switch {
			case ansi.HasCsiPrefix(seq) && p.Cmd != 0:
				switch p.Cmd {
				case 'm': // SGR - Select Graphic Rendition
					handleSgr(p, &pen)
				}
			case ansi.HasOscPrefix(seq) && p.Cmd != 0:
				switch p.Cmd {
				case 8: // Hyperlinks
					handleHyperlinks(p, &link)
				}
			case ansi.Equal(seq, "\n"):
				// Reset the rest of the line
				for x < dis.Width() {
					dis.SetCell(x, y, spaceCell) //nolint:errcheck
					x++
				}

				y++
				// XXX: We gotta reset the x position here because we're moving
				// to the next line. We shouldn't have any "\r\n" sequences,
				// those are replaced above.
				x = 0
			}
		}

		// Advance the state and data
		state = newState
		data = data[n:]
	}

	for x < dis.Width() {
		dis.SetCell(x, y, spaceCell) //nolint:errcheck
		x++
	}

	return linew
}

// handleSgr handles Select Graphic Rendition (SGR) escape sequences.
func handleSgr(p *ansi.Parser, pen *Style) {
	if p.ParamsLen == 0 {
		pen.Reset()
		return
	}

	params := p.Params[:p.ParamsLen]
	for i := 0; i < len(params); i++ {
		r := ansi.Param(params[i])
		param, hasMore := r.Param(), r.HasMore() // Are there more subparameters i.e. separated by ":"?
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
				nextParam := ansi.Param(params[i+1]).Param()
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
			pen.Foreground(ansi.Black + ansi.BasicColor(param-30)) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			if c := readColor(&i, params); c != nil {
				pen.Foreground(c)
			}
		case 39: // Default foreground
			pen.Foreground(nil)
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			pen.Background(ansi.Black + ansi.BasicColor(param-40)) //nolint:gosec
		case 48: // Set background 256 or truecolor
			if c := readColor(&i, params); c != nil {
				pen.Background(c)
			}
		case 49: // Default Background
			pen.Background(nil)
		case 58: // Set underline color
			if c := readColor(&i, params); c != nil {
				pen.UnderlineColor(c)
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

// handleHyperlinks handles hyperlink escape sequences.
func handleHyperlinks(p *ansi.Parser, link *Link) {
	params := bytes.Split(p.Data[:p.DataLen], []byte{';'})
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
