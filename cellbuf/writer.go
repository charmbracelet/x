package cellbuf

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// ScreenWriter represents a writer that writes to a [Screen] parsing ANSI
// escape sequences and Unicode characters and converting them into cells that
// can be written to a cell [Buffer].
type ScreenWriter struct {
	*Screen
}

// NewScreenWriter creates a new ScreenWriter that writes to the given Screen.
// This is a convenience function for creating a ScreenWriter.
func NewScreenWriter(s *Screen) *ScreenWriter {
	return &ScreenWriter{s}
}

// Write writes the given bytes to the screen.
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) Write(p []byte) (n int, err error) {
	printString(s.Screen, s.cur.X, s.cur.Y, s.Bounds(), p, false, "")
	return len(p), nil
}

// SetContent clears the screen with blank cells, and sets the given string as
// its content. If the height or width of the string exceeds the height or
// width of the screen, it will be truncated.
//
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape sequences.
func (s *ScreenWriter) SetContent(str string) {
	s.SetContentRect(str, s.Bounds())
}

// SetContentRect clears the rectangle within the screen with blank cells, and
// sets the given string as its content. If the height or width of the string
// exceeds the height or width of the screen, it will be truncated.
//
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) SetContentRect(str string, rect Rectangle) {
	// Replace all "\n" with "\r\n" to ensure the cursor is reset to the start
	// of the line. Make sure we don't replace "\r\n" with "\r\r\n".
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\n", "\r\n")
	s.ClearRect(rect)
	printString(s.Screen, rect.Min.X, rect.Min.Y, rect, str, true, "")
}

// Print prints the string at the current cursor position. It will wrap the
// string to the width of the screen if it exceeds the width of the screen.
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) Print(str string, v ...interface{}) {
	if len(v) > 0 {
		str = fmt.Sprintf(str, v...)
	}
	printString(s.Screen, s.cur.X, s.cur.Y, s.Bounds(), str, false, "")
}

// PrintAt prints the string at the given position. It will wrap the string to
// the width of the screen if it exceeds the width of the screen.
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) PrintAt(x, y int, str string, v ...interface{}) {
	if len(v) > 0 {
		str = fmt.Sprintf(str, v...)
	}
	printString(s.Screen, x, y, s.Bounds(), str, false, "")
}

// PrintCrop prints the string at the current cursor position and truncates the
// text if it exceeds the width of the screen. Use tail to specify a string to
// append if the string is truncated.
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) PrintCrop(str string, tail string) {
	printString(s.Screen, s.cur.X, s.cur.Y, s.Bounds(), str, true, tail)
}

// PrintCropAt prints the string at the given position and truncates the text
// if it exceeds the width of the screen. Use tail to specify a string to append
// if the string is truncated.
// This will recognize ANSI [ansi.SGR] style and [ansi.SetHyperlink] escape
// sequences.
func (s *ScreenWriter) PrintCropAt(x, y int, str string, tail string) {
	printString(s.Screen, x, y, s.Bounds(), str, true, tail)
}

// printString draws a string starting at the given position.
func printString[T []byte | string](s *Screen, x, y int, bounds Rectangle, str T, truncate bool, tail string) {
	p := ansi.GetParser()
	defer ansi.PutParser(p)

	var tailc Cell
	if truncate && len(tail) > 0 {
		if s.method == ansi.WcWidth {
			tailc = *NewCellString(tail)
		} else {
			tailc = *NewGraphemeCell(tail)
		}
	}

	decoder := ansi.DecodeSequenceWc[T]
	if s.method == ansi.GraphemeWidth {
		decoder = ansi.DecodeSequence[T]
	}

	var cell Cell
	var state byte
	for len(str) > 0 {
		seq, width, n, newState := decoder(str, state, p)

		switch width {
		case 1, 2, 3, 4: // wide cells can go up to 4 cells wide
			cell.Width += width
			cell.Append([]rune(string(seq))...)

			if !truncate && x+cell.Width > bounds.Max.X && y+1 < bounds.Max.Y {
				// Wrap the string to the width of the window
				x = bounds.Min.X
				y++
			}
			if Pos(x, y).In(bounds) {
				if truncate && tailc.Width > 0 && x+cell.Width > bounds.Max.X-tailc.Width {
					// Truncate the string and append the tail if any.
					cell := tailc
					cell.Style = s.cur.Style
					cell.Link = s.cur.Link
					s.SetCell(x, y, &cell)
					x += tailc.Width
				} else {
					// Print the cell to the screen
					cell.Style = s.cur.Style
					cell.Link = s.cur.Link
					s.SetCell(x, y, &cell) //nolint:errcheck
					x += width
				}
			}

			// String is too long for the line, truncate it.
			// Make sure we reset the cell for the next iteration.
			cell.Reset()
		default:
			// Valid sequences always have a non-zero Cmd.
			// TODO: Handle cursor movement and other sequences
			switch {
			case ansi.HasCsiPrefix(seq) && p.Command() == 'm':
				// SGR - Select Graphic Rendition
				ReadStyle(p.Params(), &s.cur.Style)
			case ansi.HasOscPrefix(seq) && p.Command() == 8:
				// Hyperlinks
				ReadLink(p.Data(), &s.cur.Link)
			case ansi.Equal(seq, T("\n")):
				y++
			case ansi.Equal(seq, T("\r")):
				x = bounds.Min.X
			default:
				cell.Append([]rune(string(seq))...)
			}
		}

		// Advance the state and data
		state = newState
		str = str[n:]
	}

	// Make sure to set the last cell if it's not empty.
	if !cell.Empty() {
		s.SetCell(x, y, &cell) //nolint:errcheck
		cell.Reset()
	}
}
