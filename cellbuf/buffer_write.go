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
func setContent[
	T string | []byte,
	TReplaceAllFunc func(s T, old T, new T) T, //nolint:predeclared
	TDecodeRuneFunc func(p T) (rune, int),
](
	buf Grid,
	data T,
	x, y int,
	w, h int,
	method WidthMethod,
	replaceAll TReplaceAllFunc,
	decodeRune TDecodeRuneFunc,
) []int {
	var cell Cell
	var pen CellStyle
	var link CellLink
	origX := x

	p := ansi.GetParser()
	defer ansi.PutParser(p)
	data = replaceAll(data, T("\r\n"), T("\n"))

	// linew is a slice of line widths. We use this to keep track of the
	// written widths of each line. We use this information later to optimize
	// rendering of the buffer.
	linew := make([]int, h)

	var pendingWidth int

	var state byte
	for len(data) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(data, state, p)

		switch width {
		case 2, 3, 4: // wide cells can go up to 4 cells wide

			switch method {
			case WcWidth:
				if r, rw := decodeRune(data); r != utf8.RuneError {
					n = rw
					width = wcwidth.RuneWidth(r)
					seq = T(string(r))
					newState = 0
				}
			case GraphemeWidth:
				// [ansi.DecodeSequence] already handles grapheme clusters
			}

			// Mark wide cells with emptyCell zero width
			// We set the wide cell down below
			for j := 1; j < width; j++ {
				buf.Set(x+j, y, emptyCell) //nolint:errcheck
			}
			fallthrough
		case 1:
			cell.Content = string(seq)
			cell.Width = width
			cell.Style = pen
			cell.Link = link

			// When a wide cell is partially overwritten, we need
			// to fill the rest of the cell with space cells to
			// avoid rendering issues.
			if prev, err := buf.At(x, y); err == nil {
				if !cell.Equal(prev) && prev.Width > 1 {
					c := prev
					c.Content = " "
					c.Width = 1
					for j := 0; j < prev.Width; j++ {
						buf.Set(x+j, y, c) //nolint:errcheck
					}
				} else if prev.Width == 0 {
					// Find the wide cell and set it to space cell.
					var wide Cell
					var wx, wy int
					for j := 1; j < 4; j++ {
						if c, err := buf.At(x-j, y); err == nil && c.Width > 1 {
							wide = c
							wx, wy = x-j, y
							break
						}
					}
					if !wide.IsEmpty() {
						c := wide
						c.Content = " "
						c.Width = 1
						for j := 0; j < wide.Width; j++ {
							buf.Set(wx+j, wy, c) //nolint:errcheck
						}
					}
				}
			}

			buf.Set(x, y, cell) //nolint:errcheck

			// Advance the cursor and line width
			x += cell.Width
			if cell.Equal(spaceCell) {
				pendingWidth += cell.Width
			} else {
				linew[y] += cell.Width + pendingWidth
				pendingWidth = 0
			}

			cell.Reset()
		default:
			// Valid sequences always have a non-zero Cmd.
			switch {
			case ansi.HasCsiPrefix(seq) && p.Cmd != 0:
				params := p.Params[:p.ParamsLen]
				switch p.Cmd {
				case 'm': // SGR - Select Graphic Rendition
					if p.ParamsLen == 0 {
						pen.Reset()
					}
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
			case ansi.HasOscPrefix(seq) && p.Cmd != 0:
				switch p.Cmd {
				case 8: // Hyperlinks
					params := bytes.Split(p.Data[:p.DataLen], []byte{';'})
					if len(params) != 3 {
						break
					}
					var id string
					for _, param := range bytes.Split(params[1], []byte{':'}) {
						if bytes.HasPrefix(param, []byte("id=")) {
							id = string(param)
						}
					}
					link.URLID = id
					link.URL = string(params[2])
				}
			case ansi.Equal(seq, T("\n")):
				// Reset the rest of the line
				for x < w {
					buf.Set(x, y, spaceCell) //nolint:errcheck
					x++
				}

				y++
				// XXX: We gotta reset the x position here because we're moving
				// to the next line. We shouldn't have any "\r\n" sequences,
				// those are replaced above.
				x = origX
			}
		}

		// Advance the state and data
		state = newState
		data = data[n:]
	}

	for x < w {
		buf.Set(x, y, spaceCell) //nolint:errcheck
		x++
	}

	return linew
}

// Grid represents an interface for a grid of cells that can be written to and
// read from.
type Grid interface {
	// Width returns the width of the grid.
	Width() int

	// Height returns the height of the grid.
	Height() int

	// Set writes a cell to the grid at the given position.
	Set(x, y int, c Cell)

	// At returns the cell at the given position.
	At(x, y int) (Cell, error)

	// SetContent writes the given data to the grid starting from the first cell.
	SetContent(data string) []int

	// Method returns the width method used by the grid.
	Method() WidthMethod

	// SetMethod sets the width method of the grid.
	SetMethod(method WidthMethod)
}

func setStringContent(b Grid, c string, x, y, w, h int, method WidthMethod) []int {
	return setContent(b, c, x, y, w, h, method, strings.ReplaceAll, utf8.DecodeRuneInString)
}

// setBufferContent writes the given data to the buffer starting from the first cell.
func setBufferContent(g Grid, b *Buffer, content string) []int {
	height := Height(content)
	if area := b.width * height; len(b.cells) < area {
		ln := len(b.cells)
		b.cells = append(b.cells, make([]Cell, area-ln)...)
		// Fill the buffer with space cells
		for i := ln; i < area; i++ {
			b.cells[i] = spaceCell
		}
	} else if len(b.cells) > area {
		// Truncate the buffer if necessary
		b.cells = b.cells[:area]
	}

	return setStringContent(g, content, 0, 0, b.width, height, b.method)
}

// SetContent sets the content of the buffer from a string.
func (b *Buffer) SetContent(s string) []int {
	return setBufferContent(b, b, s)
}
