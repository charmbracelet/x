package vt

import (
	"github.com/charmbracelet/x/cellbuf"
)

// Cell represents a single cell in the terminal screen.
type Cell = cellbuf.Cell

// Link represents a hyperlink in the terminal screen.
type Link = cellbuf.Link

// Style represents the Style of a cell.
type Style = cellbuf.Style

// Rectangle represents a rectangle in the terminal screen.
type Rectangle = cellbuf.Rectangle

// Position represents a position in the terminal screen.
type Position = cellbuf.Position
