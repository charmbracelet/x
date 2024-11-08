package cellbuf

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
	Link Link

	// Content is the string representation of the cell as a grapheme cluster.
	Content string

	// Width is the mono-space width of the grapheme cluster.
	Width int
}

// Equal returns whether the cell is equal to the other cell.
func (c Cell) Equal(o Cell) bool {
	return c.Width == o.Width &&
		c.Content == o.Content &&
		c.Style.Equal(o.Style) &&
		c.Link.Equal(o.Link)
}

// Empty returns whether the cell is empty.
func (c Cell) Empty() bool {
	return c.Content == "" &&
		c.Width == 0 &&
		c.Style.Empty() &&
		c.Link.Empty()
}

// Reset resets the cell to the default state zero value.
func (c *Cell) Reset() {
	c.Content = ""
	c.Width = 0
	c.Style.Reset()
	c.Link.Reset()
}
