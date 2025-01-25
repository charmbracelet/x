package cellbuf

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// Window represents a [Screen] 2D window.
type Window struct {
	s      *Screen     // the screen this window belongs to
	cur    Cursor      // the current cursor pos, style, and link
	method ansi.Method // the method to use for calculating the width of the cells
	x, y   int         // the starting position of the window
	w, h   int         // the width and height of the window
}

// NewWindow creates a new window. Note that the window is not
// bound to the screen until it is used to draw something.
func (s *Screen) NewWindow(x, y, w, h int) *Window {
	c := new(Window)
	c.s = s
	c.x, c.y = x, y
	c.w, c.h = w, h
	return c
}

// DefaultWindow creates a new window that covers the whole screen.
func (s *Screen) DefaultWindow() *Window {
	return s.NewWindow(0, 0, s.Width(), s.Height())
}

// NewWindow creates a new sub-window. Note that the window is not bound to
// the screen until it is used to draw something.
func (c *Window) NewWindow(x, y, w, h int) *Window {
	return c.s.NewWindow(c.x+x, c.y+y, w, h)
}

// Bounds returns the bounds of the window.
func (c *Window) Bounds() Rectangle {
	return Rect(c.x, c.y, c.w, c.h)
}

// Width returns the width of the window.
func (c *Window) Width() int {
	return c.w
}

// Height returns the height of the window.
func (c *Window) Height() int {
	return c.h
}

// X returns the x position of the window.
func (c *Window) X() int {
	return c.x
}

// Y returns the y position of the window.
func (c *Window) Y() int {
	return c.y
}

// CellAt returns the cell at the given position. If the position is out of
// bounds, it will return nil.
func (c *Window) CellAt(x, y int) *Cell {
	if !Pos(x, y).In(c.Bounds()) {
		return nil
	}
	return c.s.Cell(c.x+x, c.y+y)
}

// SetMethod sets the method to use for calculating the width of the cells.
// The default method is [WcWidth].
func (c *Window) SetMethod(method ansi.Method) {
	c.method = method
}

// SetForegroundColor sets the foreground color of the window. Use `nil` to
// use the default color.
func (c *Window) SetForegroundColor(color color.Color) {
	c.cur.Style.Fg = color
}

// SetBackgroundColor sets the background color of the window. Use `nil` to
// use the default color.
func (c *Window) SetBackgroundColor(color color.Color) {
	c.cur.Style.Bg = color
}

// SetAttributes sets the text attributes of the window.
func (c *Window) SetAttributes(attrs AttrMask) {
	c.cur.Style.Attrs = attrs
}

// EnableAttributes enables the given text attributes of the window. Use zero
// to disable all attributes.
func (c *Window) EnableAttributes(attrs AttrMask) {
	c.cur.Style.Attrs |= attrs
}

// DisableAttributes disables the given text attributes of the window.
func (c *Window) DisableAttributes(attrs AttrMask) {
	c.cur.Style.Attrs &^= attrs
}

// SetUnderlineStyle sets the underline attribute of the window. Use
// [NoUnderline] or zero to remove the underline attribute.
func (c *Window) SetUnderlineStyle(u UnderlineStyle) {
	c.cur.Style.UlStyle = u
}

// SetUnderlineColor sets the underline color of the window. Use `nil` to use
// the default color.
func (c *Window) SetUnderlineColor(color color.Color) {
	c.cur.Style.Ul = color
}

// SetHyperlink sets the hyperlink of the window. Use an empty string to
// remove the hyperlink. Use opts to set the hyperlink options such as `id=123`
// etc.
func (c *Window) SetHyperlink(link string, opts ...string) {
	c.cur.Link = Link{
		URL:   link,
		URLID: strings.Join(opts, ":"),
	}
}

// ResetHyperlink resets the hyperlink of the window.
func (c *Window) ResetHyperlink() {
	c.cur.Link = Link{}
}

// Reset resets the cursor position, styles and attributes.
func (c *Window) Reset() {
	c.cur = Cursor{}
}

// Resize resizes the window to the given width and height. If the new size is
// out of bounds, it will do nothing.
func (c *Window) Resize(w, h int) {
	c.w, c.h = w, h
}

// SetContent clears the window with blank cells, and draws the given string.
func (c *Window) SetContent(s string) {
	// Replace all "\n" with "\r\n" to ensure the cursor is reset to the start
	// of the line. Make sure we don't replace "\r\n" with "\r\r\n".
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\r\n")
	c.Clear()
	c.cur.X, c.cur.Y = 0, 0
	c.PrintTruncate(s, "")
}

// Fill fills the window with the given cell and resets the cursor position,
// styles and attributes.
func (c *Window) Fill(cell *Cell) bool {
	return c.s.FillRect(cell, c.Bounds())
}

// FillString fills the window with the given string and resets the cursor
// position, styles and attributes.
func (c *Window) FillString(s string) (v bool) {
	switch c.method {
	case ansi.WcWidth:
		v = c.Fill(NewCellString(s))
	case ansi.GraphemeWidth:
		v = c.Fill(NewGraphemeCell(s))
	}
	return
}

// Clear clears the window with blank cells and resets the cursor position,
// styles and attributes.
func (c *Window) Clear() bool {
	return c.s.ClearRect(c.Bounds())
}

// MoveTo moves the cursor to the given position. If the position is out of
// bounds, it will do nothing.
func (c *Window) MoveTo(x, y int) (v bool) {
	if !Pos(c.x+x, c.y+y).In(c.Bounds()) {
		return
	}
	c.cur.X, c.cur.Y = x, y
	return c.s.MoveTo(c.x+x, c.y+y)
}

// Print prints the given string at the current cursor position wrapping the
// text if necessary. If the cursor is out of bounds, it will do nothing.
func (c *Window) Print(format string, v ...interface{}) {
	if len(v) > 0 {
		format = fmt.Sprintf(format, v...)
	}
	c.drawString(format, c.cur.X, c.cur.Y, defaultDrawOpts)
}

// PrintAt prints the given string at the given position wrapping the text if
// necessary. If the position is out of bounds, it will do nothing.
func (c *Window) PrintAt(x, y int, format string, v ...interface{}) {
	if !Pos(c.x+x, c.y+y).In(c.Bounds()) {
		return
	}
	if len(v) > 0 {
		format = fmt.Sprintf(format, v...)
	}
	c.drawString(format, x, y, defaultDrawOpts)
}

// PrintTruncate draws a string starting at the given position and
// truncates the string with the given tail if necessary.
func (c *Window) PrintTruncate(s string, tail string) {
	c.drawString(s, c.cur.X, c.cur.Y, &drawOpts{tail: tail, truncate: true})
}

// PrintTruncateAt draws a string starting at the given position and
// truncates the string with the given tail if necessary.
// If the position is out of bounds, it will do nothing.
func (c *Window) PrintTruncateAt(x, y int, s string, tail string) {
	if !Pos(c.x+x, c.y+y).In(c.Bounds()) {
		return
	}
	c.drawString(s, x, y, &drawOpts{tail: tail, truncate: true})
}

// SetCell sets a cell at the given position. If the position is out of bounds,
// it will do nothing.
func (c *Window) SetCell(x, y int, cell *Cell) (v bool) {
	pos := Pos(c.x+x, c.y+y)
	if !pos.In(c.Bounds()) {
		return
	}
	return c.s.SetCell(pos.X, pos.Y, cell)
}

// drawOpts represents the options for drawing a string.
type drawOpts struct {
	tail     string // the tail to append if the string is truncated, empty by default to crop
	truncate bool   // truncate the string if it's too long
}

var defaultDrawOpts = &drawOpts{}

// drawString draws a string starting at the given position.
func (c *Window) drawString(s string, x, y int, opts *drawOpts) {
	if opts == nil {
		opts = defaultDrawOpts
	}

	wrapCursor := func() {
		// Wrap the string to the width of the window
		x = 0
		y++
	}

	p := ansi.GetParser()
	defer ansi.PutParser(p)

	var tail Cell
	if opts.truncate && len(opts.tail) > 0 {
		if c.method == ansi.WcWidth {
			tail = *NewCellString(opts.tail)
		} else {
			tail = *NewGraphemeCell(opts.tail)
		}
	}

	var state byte
	for len(s) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(s, state, p)

		var cell *Cell
		switch width {
		case 1, 2, 3, 4: // wide cells can go up to 4 cells wide
			switch c.method {
			case ansi.WcWidth:
				cell = NewCellString(seq)

				// We're breaking the grapheme to respect wcwidth's behavior
				// while keeping combining characters together.
				n = utf8.RuneLen(cell.Rune)
				for _, c := range cell.Comb {
					n += utf8.RuneLen(c)
				}
				newState = 0

			case ansi.GraphemeWidth:
				// [ansi.DecodeSequence] already handles grapheme clusters
				cell = newGraphemeCell(seq, width)
			}

			if !opts.truncate && x >= c.w {
				// Auto wrap the cursor.
				wrapCursor()
				if y >= c.h {
					break
				}
			} else if opts.truncate && x+width > c.w-tail.Width {
				if !Pos(x, y).In(c.Bounds()) {
					break
				}

				// Truncate the string and append the tail if any.
				cell := tail
				cell.Style = c.cur.Style
				cell.Link = c.cur.Link
				c.SetCell(x, y, &cell)
				break
			}

			cell.Style = c.cur.Style
			cell.Link = c.cur.Link

			// NOTE: [Window.SetCell] will handle out of bounds positions.
			c.SetCell(x, y, cell) //nolint:errcheck

			// Advance the cursor and line width
			x += cell.Width
		default:
			// Valid sequences always have a non-zero Cmd.
			// TODO: Handle cursor movement and other sequences
			switch {
			case ansi.HasCsiPrefix(seq) && p.Command() != 0:
				switch p.Command() {
				case 'm': // SGR - Select Graphic Rendition
					ReadStyle(p.Params(), &c.cur.Style)
				}
			case ansi.HasOscPrefix(seq) && p.Command() != 0:
				switch p.Command() {
				case 8: // Hyperlinks
					ReadLink(p.Data(), &c.cur.Link)
				}
			case ansi.Equal(seq, "\n"):
				if y+1 < c.y+c.h {
					y++
				}
			case ansi.Equal(seq, "\r"):
				x = 0
			}
		}

		// Advance the state and data
		state = newState
		s = s[n:]
	}

	c.cur.X, c.cur.Y = x, y
}
