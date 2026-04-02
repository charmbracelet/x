package vt

import "testing"

func TestReverseIndexClampsOversizedScrollMarginsAfterShrink(t *testing.T) {
	t.Parallel()

	term := NewEmulator(40, 24)
	term.Resize(40, 13)

	if _, err := term.WriteString("\x1b[1;21r\x1b[H\x1bM"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	pos := term.CursorPosition()
	if pos.X != 0 || pos.Y != 0 {
		t.Fatalf("CursorPosition() = (%d, %d), want (0, 0)", pos.X, pos.Y)
	}
	if got := term.CellAt(0, 0).Content; got != " " {
		t.Fatalf("CellAt(0, 0).Content = %q, want blank top cell after reverse index scroll", got)
	}
}

func TestReflowWrappedPositionHandlesEmptyWrappedCounts(t *testing.T) {
	t.Parallel()

	pos := reflowWrappedPosition(nil, reflowPosition{logical: 3, offset: 12}, 20)
	if pos.X != 0 || pos.Y != 0 {
		t.Fatalf("reflowWrappedPosition(nil, ...) = (%d, %d), want (0, 0)", pos.X, pos.Y)
	}
}

func TestAltScreenEntryPreservesHiddenCursorWhenHideArrivesFirst(t *testing.T) {
	t.Parallel()

	term := NewEmulator(40, 24)

	if _, err := term.WriteString("\x1b[?25l\x1b[?1049h"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	if !term.Cursor().Hidden {
		t.Fatal("Cursor().Hidden = false after hide-before-alt-screen, want true")
	}
}
