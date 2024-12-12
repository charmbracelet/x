package cellbuf

// Segment represents a continuous segment of cells with the same style
// attributes and hyperlink.
type Segment struct {
	Style   Style
	Link    Link
	Content string
	Width   int
}

// Paint writes the given data to the canvas. If rect is not nil, it only
// writes to the rectangle.
func Paint(d Window, content string) []int {
	return PaintRect(d, content, d.Bounds())
}

// PaintRect writes the given data to the canvas starting from the given
// rectangle.
func PaintRect(d Window, content string, rect Rectangle) []int {
	return setContent(d, content, WcWidth, rect)
}
