package vt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/cellbuf"
)

// XXX: Use golden files for these tests

func TestBuffer_new(t *testing.T) {
	t.Parallel()

	b := cellbuf.NewBuffer(10, 5)
	if b == nil {
		t.Error("expected buffer, got nil")
	}
	if b.Width() != 10 {
		t.Errorf("expected width %d, got %d", 10, b.Width())
	}
	if b.Height() != 5 {
		t.Errorf("expected height %d, got %d", 5, b.Height())
	}
}

func TestBuffer_setCell(t *testing.T) {
	t.Parallel()

	b := cellbuf.NewBuffer(10, 5)
	if !b.SetCell(0, 0, cellbuf.NewCell('a')) {
		t.Error("expected SetCell to return true")
	}
	if cell := b.Cell(0, 0); cell == nil || cell.String() != "a" {
		t.Errorf("expected cell at 0,0 to be 'a', got %v", cell)
	}

	// Single rune emoji
	if !b.SetCell(1, 0, cellbuf.NewCell('👍')) {
		t.Error("expected SetCell to return true")
	}
	if cell := b.Cell(1, 0); cell == nil || cell.String() != "👍" || cell.Width != 2 {
		t.Errorf("expected cell at 1,0 to be '👍', got %v", cell)
	}
	if cell := b.Cell(2, 0); cell == nil || cell.String() != "" || cell.Width != 0 {
		t.Errorf("expected cell at 2,0 to be empty, got %#v", cell)
	}

	// Wide rune character
	if !b.SetCell(3, 0, cellbuf.NewCell('あ')) {
		t.Error("expected SetCell to return true")
	}
	if cell := b.Cell(3, 0); cell == nil || cell.String() != "あ" || cell.Width != 2 {
		t.Errorf("expected cell at 3,0 to be 'あ', got %v", cell)
	}

	// Overwrite a wide cell with a single rune
	if !b.SetCell(3, 0, cellbuf.NewCell('b')) {
		t.Error("expected SetCell to return true")
	}
	if cell := b.Cell(3, 0); cell == nil || cell.String() != "b" || cell.Width != 1 {
		t.Errorf("expected cell at 3,0 to be 'b', got %v", cell)
	}
	if cell := b.Cell(4, 0); cell == nil || cell.String() != " " || cell.Width != 1 {
		t.Errorf("expected cell at 4,0 to be blank, got %v", cell)
	}

	// Overwrite a wide cell placeholder with a single rune
	if !b.SetCell(3, 0, cellbuf.NewCell('あ')) {
		t.Error("expected SetCell to return true")
	}
	if !b.SetCell(4, 0, cellbuf.NewCell('c')) {
		t.Error("expected SetCell to return true")
	}
	if cell := b.Cell(3, 0); cell == nil || cell.String() != " " || cell.Width != 1 {
		t.Errorf("expected cell at 3,0 to be 'あ', got %v", cell)
	}
	if cell := b.Cell(4, 0); cell == nil || cell.String() != "c" || cell.Width != 1 {
		t.Errorf("expected cell at 4,0 to be 'c', got %v", cell)
	}
}

func TestBuffer_resize(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.SetCell(0, 0, cellbuf.NewCell('a'))
	b.SetCell(1, 0, cellbuf.NewCell('b'))
	b.SetCell(2, 0, cellbuf.NewCell('c'))
	if b.Width() != 10 {
		t.Errorf("expected width %d, got %d", 10, b.Width())
	}
	if b.Height() != 5 {
		t.Errorf("expected height %d, got %d", 5, b.Height())
	}

	b.Resize(5, 3)
	if b.Width() != 5 {
		t.Errorf("expected width %d, got %d", 5, b.Width())
	}
	if b.Height() != 3 {
		t.Errorf("expected height %d, got %d", 3, b.Height())
	}
}

func TestBuffer_fill(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	for y := range b.Height() {
		for x := range b.Width() {
			if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
				t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
			}
		}
	}
}

func TestBuffer_clear(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	b.Clear()
	for y := range b.Height() {
		for x := range b.Width() {
			if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
				t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
			}
		}
	}
}

func TestBuffer_fillClearRect(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	r := cellbuf.Rect(1, 1, 3, 3)
	b.ClearRect(r)
	for y := range b.Height() {
		for x := range b.Width() {
			pt := cellbuf.Pos(x, y)
			if pt.In(r) {
				if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}
}

func TestBuffer_insertLine(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	b.InsertLine(1, 1, nil)
	for y := range b.Height() {
		for x := range b.Width() {
			if y == 1 {
				if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_insertLineInRect(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	r := cellbuf.Rect(1, 1, 3, 3)
	n := 2                         // The number of lines to insert
	b.InsertLineRect(1, n, nil, r) // Insert n lines at y=1 within the rectangle r
	for y := range b.Height() {
		for x := range b.Width() {
			pt := cellbuf.Pos(x, y)
			if pt.In(r) && y >= 1 && y < 1+n {
				if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_deleteLine(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	b.FillRect(cellbuf.NewCell('b'), cellbuf.Rect(0, 1, 10, 1))
	t.Log("\n" + renderBuffer(b))

	b.DeleteLine(1, 1, nil)
	if b.Height() != 5 {
		t.Error("expected height to be less than 5")
	}
	for y := range b.Height() {
		for x := range b.Width() {
			if y == b.Height()-1 {
				if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_deleteLineInRect(t *testing.T) {
	b := cellbuf.NewBuffer(10, 5)
	b.Fill(cellbuf.NewCell('a'))
	t.Log("\n" + renderBuffer(b))
	r := cellbuf.Rect(1, 1, 3, 3)
	n := 2                         // The number of lines to delete
	b.DeleteLineRect(1, n, nil, r) // Delete n lines at y=1 within the rectangle r
	t.Log("\n" + renderBuffer(b))
	for y := r.Max.Y - 1; y < r.Dy(); y++ {
		for x := range b.Width() {
			pt := cellbuf.Pos(x, y)
			if pt.In(r) && y >= 1 && y < 1+n {
				if cell := b.Cell(x, y); cell == nil || cell.String() != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell := b.Cell(x, y); cell == nil || cell.String() != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}
}

func renderBuffer(b *Buffer) string {
	var out strings.Builder
	for y := range b.Height() {
		var line string
		for x := range b.Width() {
			cell := b.Cell(x, y)
			if cell == nil {
				cell = cellbuf.NewCell(' ')
			}
			line += cell.String()
		}
		out.WriteString(fmt.Sprintf("%q\n", line))
	}
	return out.String()
}
