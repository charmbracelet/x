package cellbuf

// Method is a type that represents the how the renderer should calculate the
// display width of cells.
type Method uint8

// Display width modes.
const (
	WcWidth Method = iota
	GraphemeWidth
)
