package vt

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func testLine(s string, width int) uv.Line {
	line := uv.NewLine(width)
	for i, r := range s {
		line[i] = uv.Cell{Content: string(r), Width: 1}
	}
	return line
}

func lineText(l uv.Line) string {
	var b strings.Builder
	for _, c := range l {
		b.WriteString(c.Content)
	}
	return b.String()
}

func scrollbackText(sb *Scrollback) []string {
	var lines []string
	for i := range sb.Len() {
		lines = append(lines, lineText(sb.Line(i)))
	}
	return lines
}

func TestScrollbackRing(t *testing.T) {
	sb := NewScrollback(3)
	for i := range 8 {
		// Lines of different lengths, so reused memory must be resized.
		sb.Push(testLine(strings.Repeat(fmt.Sprint(i), 1+i%3), 10))
	}
	want := []string{"55555"[:3], "6", "77"}
	if got := scrollbackText(sb); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("expected the newest lines %q oldest first, got %q", want, got)
	}
	if got := sb.Lines(); len(got) != 3 || lineText(got[0]) != "555" || lineText(got[2]) != "77" {
		t.Fatalf("expected Lines in order, got %d lines", len(got))
	}
	if sb.Line(-1) != nil || sb.Line(3) != nil || sb.CellAt(5, 0) != nil {
		t.Fatal("expected nil out of bounds")
	}

	// Shrinking keeps the newest lines in order.
	sb.SetMaxLines(2)
	if got := scrollbackText(sb); fmt.Sprint(got) != fmt.Sprint([]string{"6", "77"}) {
		t.Fatalf("expected the 2 newest lines after shrinking, got %q", got)
	}
	// Growing keeps them, and new lines go after them.
	sb.SetMaxLines(4)
	sb.Push(testLine("8", 10))
	sb.Push(testLine("9", 10))
	sb.Push(testLine("a", 10))
	if got := scrollbackText(sb); fmt.Sprint(got) != fmt.Sprint([]string{"77", "8", "9", "a"}) {
		t.Fatalf("expected lines in order after growing, got %q", got)
	}

	sb.Clear()
	if sb.Len() != 0 || sb.Line(0) != nil {
		t.Fatal("expected an empty scrollback after clear")
	}
	sb.Push(testLine("b", 10))
	if got := scrollbackText(sb); fmt.Sprint(got) != "[b]" {
		t.Fatalf("expected a line after clearing, got %q", got)
	}
}

func TestScrollbackTrimsBlanks(t *testing.T) {
	sb := NewScrollback(10)
	line := testLine("ab", 8)
	line[5] = uv.Cell{} // a zero cell, e.g. the rest of a wide character
	sb.Push(line)
	if got := len(sb.Line(0)); got != 2 {
		t.Fatalf("expected trailing blanks trimmed to 2 cells, got %d", got)
	}

	// Blank cells with a background are kept.
	line = testLine("ab", 8)
	line[4].Style.Bg = uv.EmptyCell.Style.Fg // nil: still blank
	line[6] = uv.Cell{Content: " ", Width: 1, Style: uv.Style{Attrs: uv.AttrReverse}}
	sb.Push(line)
	if got := len(sb.Line(1)); got != 7 {
		t.Fatalf("expected a styled blank to be kept, got %d cells", got)
	}

	// The pushed line is copied.
	line = testLine("cd", 8)
	sb.Push(line)
	line[0].Content = "x"
	if got := lineText(sb.Line(2)); got != "cd" {
		t.Fatalf("expected the scrollback to keep its copy, got %q", got)
	}
}
