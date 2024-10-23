package cellbuf

// WidthMethod is a type that represents the how the renderer should calculate
// the display width of cells.
type WidthMethod uint8

// Display width modes.
const (
	WcWidth WidthMethod = iota
	GraphemeWidth
)
