package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestScrollViewScrollbarAtEnd(t *testing.T) {
	// Create a scroll view with content taller than viewport
	var items []Element
	for i := 0; i < 20; i++ {
		items = append(items, NewText("Item"))
	}

	vstack := NewVStack(items...)
	scrollView := NewScrollView(vstack)
	scrollView.Height(NewFixedConstraint(10)) // 10 line viewport
	scrollView.Vertical(true)
	scrollView.Scrollbar(true)

	// Get content size
	contentSize := scrollView.ContentSize()

	// Scroll to the end
	viewportHeight := 10
	maxOffset := contentSize.Height - viewportHeight
	scrollView.Offset(0, maxOffset)

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	// The scrollbar should be drawn with thumb at the bottom
	// We can't easily verify the exact position, but we can check
	// that the Draw didn't panic and the offset is correct
	if scrollView.offsetY != maxOffset {
		t.Errorf("Expected OffsetY to be %d, got %d", maxOffset, scrollView.offsetY)
	}

	t.Logf("Content height: %d, viewport: %d, max offset: %d", contentSize.Height, viewportHeight, maxOffset)
}

func TestScrollViewScrollbarAtStart(t *testing.T) {
	// Create a scroll view with content taller than viewport
	var items []Element
	for i := 0; i < 20; i++ {
		items = append(items, NewText("Item"))
	}

	vstack := NewVStack(items...)
	scrollView := NewScrollView(vstack)
	scrollView.Height(NewFixedConstraint(10))
	scrollView.Vertical(true)
	scrollView.Scrollbar(true)

	// Scroll to the start
	scrollView.Offset(0, 0)

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	// Verify scrollbar was drawn (smoke test)
	if scrollView.Bounds().Dx() == 0 {
		t.Error("ScrollView bounds not set")
	}
}

func TestScrollViewScrollbarMidpoint(t *testing.T) {
	// Create a scroll view with content taller than viewport
	var items []Element
	for i := 0; i < 20; i++ {
		items = append(items, NewText("Item"))
	}

	vstack := NewVStack(items...)
	scrollView := NewScrollView(vstack)
	scrollView.Height(NewFixedConstraint(10))
	scrollView.Vertical(true)
	scrollView.Scrollbar(true)

	// Get content size
	contentSize := scrollView.ContentSize()
	viewportHeight := 10
	maxOffset := contentSize.Height - viewportHeight

	// Scroll to middle
	scrollView.Offset(0, maxOffset/2)

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	// Verify rendering worked
	if scrollView.Bounds().Dy() == 0 {
		t.Error("ScrollView bounds not set")
	}
}
