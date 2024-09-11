package cellbuf

import (
	"fmt"
	"strings"
)

var (
	// spaceCell is 1-cell wide, has no style, and a space rune.
	spaceCell = Cell{
		Content: " ",
		Width:   1,
	}

	// emptyCell is an empty cell.
	emptyCell = Cell{}
)

// Cell represents a single cell in the terminal screen.
type Cell struct {
	// The style of the cell. Nil style means no style. Zero value prints a
	// reset sequence.
	Style Style

	// Link is the hyperlink of the cell.
	Link Hyperlink

	// Content is the string representation of the cell as a grapheme cluster.
	Content string

	// Width is the mono-space width of the grapheme cluster.
	Width int
}

// Equal returns true if the cell is equal to the other cell.
func (c *Cell) Equal(o Cell) bool {
	return c.Content == o.Content &&
		c.Width == o.Width &&
		c.Style.Equal(o.Style) &&
		c.Link.Equal(o.Link)
}

// IsEmpty returns true if the cell is empty.
func (c *Cell) IsEmpty() bool {
	return c.Content == "" &&
		c.Width == 0 &&
		c.Style.IsEmpty() &&
		c.Link.IsEmpty()
}

// Reset resets the cell to the default state zero value.
func (c *Cell) Reset() {
	c.Content = ""
	c.Width = 0
	c.Style.Reset()
	c.Link.Reset()
}

// Info returns a string representation of the cell.
func (c *Cell) Info() string {
	var b strings.Builder

	if c.Width != 0 {
		b.WriteString("Cell{")
		b.WriteString("Grapheme: \"")
		b.WriteString(c.Content)
		b.WriteString("\", Width: ")
		fmt.Fprint(&b, c.Width)
	} else {
		b.WriteString("<empty>")
	}

	if !c.Style.IsEmpty() {
		b.WriteString(", ")
		b.WriteString(c.Style.Info())
	}

	if !c.Link.IsEmpty() {
		b.WriteString(", ")
		b.WriteString(c.Link.Info())
	}

	b.WriteString("}")

	return b.String()
}

// Hyperlink represents a hyperlink in the terminal screen.
type Hyperlink struct {
	URL   string
	URLID string
}

// Reset resets the hyperlink to the default state zero value.
func (h *Hyperlink) Reset() {
	h.URL = ""
	h.URLID = ""
}

// Equal returns true if the hyperlink is equal to the other hyperlink.
func (h Hyperlink) Equal(o Hyperlink) bool {
	return h.URL == o.URL && h.URLID == o.URLID
}

// IsEmpty returns true if the hyperlink is empty.
func (h Hyperlink) IsEmpty() bool {
	return h.URL == "" && h.URLID == ""
}

// Info returns a string representation of the hyperlink.
func (h Hyperlink) Info() string {
	if h.URL == "" && h.URLID == "" {
		return "Hyperlink{}"
	} else if h.URL == "" {
		return "Hyperlink{URLID: \"" + h.URLID + "\"}"
	}
	return "Hyperlink{URL: \"" + h.URL + "\", URLID: \"" + h.URLID + "\"}"
}
