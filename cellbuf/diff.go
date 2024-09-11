package cellbuf

// DiffCell represents a cell that has changed between in the screen.
type DiffCell struct {
	Cell
	X, Y int
}
