package vt

import "testing"

func TestBuffer_new(t *testing.T) {
	t.Parallel()

	b := NewBuffer(10, 5)
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

	b := NewBuffer(10, 5)
	if !b.SetCell(0, 0, NewCell('a')) {
		t.Error("expected SetCell to return true")
	}
	if cell, ok := b.Cell(0, 0); !ok || cell.Content != "a" {
		t.Errorf("expected cell at 0,0 to be 'a', got %v", cell)
	}

	// Single rune emoji
	if !b.SetCell(1, 0, NewCell('👍')) {
		t.Error("expected SetCell to return true")
	}
	if cell, ok := b.Cell(1, 0); !ok || cell.Content != "👍" || cell.Width != 2 {
		t.Errorf("expected cell at 1,0 to be '👍', got %v", cell)
	}
	if cell, ok := b.Cell(2, 0); !ok || cell.Content != "" || cell.Width != 0 {
		t.Errorf("expected cell at 2,0 to be empty, got %v", cell)
	}

	// Wide rune character
	if !b.SetCell(3, 0, NewCell('あ')) {
		t.Error("expected SetCell to return true")
	}
	if cell, ok := b.Cell(3, 0); !ok || cell.Content != "あ" || cell.Width != 2 {
		t.Errorf("expected cell at 3,0 to be 'あ', got %v", cell)
	}

	// Overwrite a wide cell with a single rune
	if !b.SetCell(3, 0, NewCell('b')) {
		t.Error("expected SetCell to return true")
	}
	if cell, ok := b.Cell(3, 0); !ok || cell.Content != "b" || cell.Width != 1 {
		t.Errorf("expected cell at 3,0 to be 'b', got %v", cell)
	}
	if cell, ok := b.Cell(4, 0); !ok || cell.Content != " " || cell.Width != 1 {
		t.Errorf("expected cell at 4,0 to be blank, got %v", cell)
	}

	// Overwrite a wide cell placeholder with a single rune
	if !b.SetCell(3, 0, NewCell('あ')) {
		t.Error("expected SetCell to return true")
	}
	if !b.SetCell(4, 0, NewCell('c')) {
		t.Error("expected SetCell to return true")
	}
	if cell, ok := b.Cell(3, 0); !ok || cell.Content != " " || cell.Width != 1 {
		t.Errorf("expected cell at 3,0 to be 'あ', got %v", cell)
	}
	if cell, ok := b.Cell(4, 0); !ok || cell.Content != "c" || cell.Width != 1 {
		t.Errorf("expected cell at 4,0 to be 'c', got %v", cell)
	}
}

func TestBuffer_resize(t *testing.T) {
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, NewCell('a'))
	b.SetCell(1, 0, NewCell('b'))
	b.SetCell(2, 0, NewCell('c'))
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
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
				t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
			}
		}
	}
}

func TestBuffer_clear(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	b.Clear()
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if cell, ok := b.Cell(x, y); !ok || cell.Content != " " || cell.Width != 1 {
				t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
			}
		}
	}
}

func TestBuffer_fillClearRect(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	r := Rect(1, 1, 3, 3)
	b.Clear(r)
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			pt := Pos(x, y)
			if r.Contains(pt) {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}
}

func TestBuffer_insertLine(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	b.InsertLine(1, 1)
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if y == 1 {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_insertLineInRect(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	r := Rect(1, 1, 3, 3)
	n := 2                // The number of lines to insert
	b.InsertLine(1, n, r) // Insert n lines at y=1 within the rectangle r
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			pt := Pos(x, y)
			if r.Contains(pt) && y >= 1 && y < 1+n {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_deleteLine(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	b.DeleteLine(1, 1)
	if b.Height() == 5 {
		t.Error("expected height to be less than 5")
	}
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
				t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func TestBuffer_deleteLineInRect(t *testing.T) {
	b := NewBuffer(10, 5)
	b.Fill(NewCell('a'))
	r := Rect(1, 1, 3, 3)
	n := 2                // The number of lines to delete
	b.DeleteLine(1, n, r) // Delete n lines at y=1 within the rectangle r
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			pt := Pos(x, y)
			if r.Contains(pt) && y >= 1 && y < 1+n {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != " " || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be blank, got %v", x, y, cell)
				}
			} else {
				if cell, ok := b.Cell(x, y); !ok || cell.Content != "a" || cell.Width != 1 {
					t.Errorf("expected cell at %d,%d to be 'a', got %v", x, y, cell)
				}
			}
		}
	}

	t.Log("\n" + renderBuffer(b))
}

func renderBuffer(b *Buffer) string {
	var out string
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			cell, _ := b.Cell(x, y)
			out += cell.Content
		}
		out += "\n"
	}
	return out
}