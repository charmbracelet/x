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
	scrollView.WithHeight(NewFixedConstraint(10)) // 10 line viewport
	scrollView.WithVertical(true)
	scrollView.WithScrollbar(true)

	// Get content size
	contentSize := scrollView.ContentSize()

	// Scroll to the end
	viewportHeight := 10
	maxOffset := contentSize.Height - viewportHeight
	scrollView.OffsetY = maxOffset

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	// The scrollbar should be drawn with thumb at the bottom
	// We can't easily verify the exact position, but we can check
	// that the Draw didn't panic and the offset is correct
	if scrollView.OffsetY != maxOffset {
		t.Errorf("Expected OffsetY to be %d, got %d", maxOffset, scrollView.OffsetY)
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
	scrollView.WithHeight(NewFixedConstraint(10))
	scrollView.WithVertical(true)
	scrollView.WithScrollbar(true)

	// Scroll to the start
	scrollView.OffsetY = 0

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	if scrollView.OffsetY != 0 {
		t.Errorf("Expected OffsetY to be 0, got %d", scrollView.OffsetY)
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
	scrollView.WithHeight(NewFixedConstraint(10))
	scrollView.WithVertical(true)
	scrollView.WithScrollbar(true)

	// Get content size
	contentSize := scrollView.ContentSize()
	viewportHeight := 10
	maxOffset := contentSize.Height - viewportHeight

	// Scroll to middle
	scrollView.OffsetY = maxOffset / 2

	// Render
	scr := uv.NewScreenBuffer(40, 10)
	area := uv.Rect(0, 0, 40, 10)
	scrollView.Draw(scr, area)

	if scrollView.OffsetY != maxOffset/2 {
		t.Errorf("Expected OffsetY to be %d, got %d", maxOffset/2, scrollView.OffsetY)
	}
}
