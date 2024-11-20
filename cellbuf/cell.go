package cellbuf

import "github.com/charmbracelet/x/vt"

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
type Cell = vt.Cell
