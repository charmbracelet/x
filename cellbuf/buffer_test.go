package cellbuf

import (
	"testing"
)

func TestNewCell(t *testing.T) {
	tests := []struct {
		name     string
		mainRune rune
		combRune []rune
		want     *Cell
	}{
		{
			name:     "simple ascii",
			mainRune: 'a',
			want:     &Cell{Rune: 'a', Width: 1},
		},
		{
			name:     "wide character",
			mainRune: '世',
			want:     &Cell{Rune: '世', Width: 2},
		},
		{
			name:     "combining character",
			mainRune: 'e',
			combRune: []rune{'́'}, // accent
			want:     &Cell{Rune: 'e', Comb: []rune{'́'}, Width: 1},
		},
		{
			name:     "zero width",
			mainRune: '\u200B', // zero-width space
			want:     &Cell{Rune: '\u200B', Width: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCell(tt.mainRune, tt.combRune...)
			if got.Rune != tt.want.Rune {
				t.Errorf("NewCell().Rune = %v, want %v", got.Rune, tt.want.Rune)
			}
			if got.Width != tt.want.Width {
				t.Errorf("NewCell().Width = %v, want %v", got.Width, tt.want.Width)
			}
			if len(got.Comb) != len(tt.want.Comb) {
				t.Errorf("NewCell().Comb length = %v, want %v", len(got.Comb), len(tt.want.Comb))
			}
		})
	}
}

func TestNewCellString(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want *Cell
	}{
		{
			name: "empty string",
			str:  "",
			want: &Cell{Width: 0},
		},
		{
			name: "simple ascii",
			str:  "a",
			want: &Cell{Rune: 'a', Width: 1},
		},
		{
			name: "combining character",
			str:  "é", // e with acute accent
			want: &Cell{Rune: 'é', Width: 1},
		},
		{
			name: "wide character",
			str:  "世",
			want: &Cell{Rune: '世', Width: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCellString(tt.str)
			if got.Rune != tt.want.Rune {
				t.Errorf("NewCellString().Rune = %v, want %v", got.Rune, tt.want.Rune)
			}
			if got.Width != tt.want.Width {
				t.Errorf("NewCellString().Width = %v, want %v", got.Width, tt.want.Width)
			}
		})
	}
}

func TestLine(t *testing.T) {
	tests := []struct {
		name      string
		line      Line
		wantStr   string
		wantLen   int
		wantWidth int
	}{
		{
			name:      "empty line",
			line:      Line{},
			wantStr:   "",
			wantLen:   0,
			wantWidth: 0,
		},
		{
			name:      "simple line",
			line:      Line{NewCell('a'), NewCell('b'), NewCell('c')},
			wantStr:   "abc",
			wantLen:   3,
			wantWidth: 3,
		},
		{
			name:      "line with nil cells",
			line:      Line{nil, NewCell('a'), nil},
			wantStr:   " a",
			wantLen:   3,
			wantWidth: 3,
		},
		{
			name:      "line with wide chars",
			line:      Line{NewCell('世'), NewCell('界')},
			wantStr:   "世界",
			wantLen:   2,
			wantWidth: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.String(); got != tt.wantStr {
				t.Errorf("Line.String() = %q, want %q", got, tt.wantStr)
			}
			if got := tt.line.Len(); got != tt.wantLen {
				t.Errorf("Line.Len() = %v, want %v", got, tt.wantLen)
			}
			if got := tt.line.Width(); got != tt.wantWidth {
				t.Errorf("Line.Width() = %v, want %v", got, tt.wantWidth)
			}
		})
	}
}

func TestBuffer(t *testing.T) {
	t.Run("creation and resizing", func(t *testing.T) {
		b := NewBuffer(3, 2)
		if b.Width() != 3 {
			t.Errorf("Buffer width = %d, want 3", b.Width())
		}
		if b.Height() != 2 {
			t.Errorf("Buffer height = %d, want 2", b.Height())
		}

		b.Resize(4, 3)
		if b.Width() != 4 {
			t.Errorf("After resize, buffer width = %d, want 4", b.Width())
		}
		if b.Height() != 3 {
			t.Errorf("After resize, buffer height = %d, want 3", b.Height())
		}
	})

	t.Run("cell operations", func(t *testing.T) {
		b := NewBuffer(3, 3)
		cell := NewCell('A')

		b.SetCell(1, 1, cell)
		got := b.Cell(1, 1)
		if got.Rune != 'A' {
			t.Errorf("After SetCell, got rune %c, want A", got.Rune)
		}
	})

	t.Run("clear operations", func(t *testing.T) {
		b := NewBuffer(2, 2)
		b.SetCell(0, 0, NewCell('A'))
		b.SetCell(1, 0, NewCell('B'))
		b.Clear()

		// if b.Cell(0, 0) != nil {
		// TODO: Should we return nil instead of BlankCell? Nil indicates the
		// default cell.

		if !b.Cell(0, 0).Equal(&BlankCell) {
			t.Error("After Clear, cell should be nil")
		}
	})

	t.Run("insert line", func(t *testing.T) {
		b := NewBuffer(3, 3)
		b.SetCell(0, 0, NewCell('A'))
		b.SetCell(0, 1, NewCell('B'))

		b.InsertLine(1, 1, nil)
		got := b.Cell(0, 1)
		if !got.Equal(&BlankCell) {
			t.Error("After InsertLine, inserted line should be empty")
		}
	})

	t.Run("delete line", func(t *testing.T) {
		b := NewBuffer(3, 3)
		b.SetCell(0, 0, NewCell('A'))
		b.SetCell(0, 1, NewCell('B'))

		b.DeleteLine(0, 1, nil)
		got := b.Cell(0, 0)
		if !got.Equal(NewCell('B')) {
			t.Error("After DeleteLine, first line should be empty")
		}
	})
}

func TestBufferBounds(t *testing.T) {
	b := NewBuffer(4, 3)
	bounds := b.Bounds()

	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		t.Errorf("Buffer bounds min = (%d,%d), want (0,0)", bounds.Min.X, bounds.Min.Y)
	}
	if bounds.Max.X != 4 || bounds.Max.Y != 3 {
		t.Errorf("Buffer bounds max = (%d,%d), want (4,3)", bounds.Max.X, bounds.Max.Y)
	}
}
