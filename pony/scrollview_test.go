package pony

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
)

func TestScrollViewBasic(t *testing.T) {
	const markup = `<scrollview height="5"><vstack><text>A</text><text>B</text><text>C</text><text>D</text><text>E</text><text>F</text><text>G</text></vstack></scrollview>`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestScrollViewWithOffset(t *testing.T) {
	scroll := NewScrollView(
		NewVStack(
			NewText("Line 1"),
			NewText("Line 2"),
			NewText("Line 3"),
			NewText("Line 4"),
			NewText("Line 5"),
		),
	).Height(NewFixedConstraint(3)).Offset(0, 2)

	// With offset 2, should start from line 3
	// The viewport itself should be 3 lines tall
	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	size := scroll.Layout(constraints)
	if size.Height != 3 {
		t.Errorf("ScrollView viewport height = %d, want 3", size.Height)
	}
}

func TestScrollViewContentSize(t *testing.T) {
	// Create content larger than viewport
	var lines []Element
	for i := 1; i <= 20; i++ {
		lines = append(lines, NewText("Line"))
	}

	scroll := NewScrollView(NewVStack(lines...))

	contentSize := scroll.ContentSize()
	if contentSize.Height != 20 {
		t.Errorf("ContentSize().Height = %d, want 20", contentSize.Height)
	}
}

func TestScrollMethods(t *testing.T) {
	scroll := NewScrollView(NewText("Content")).
		Height(NewFixedConstraint(10))

	// Test scroll down
	scroll.ScrollDown(5, 100, 10)
	if scroll.offsetY != 5 {
		t.Errorf("ScrollDown: OffsetY = %d, want 5", scroll.offsetY)
	}

	// Test scroll up
	scroll.ScrollUp(2)
	if scroll.offsetY != 3 {
		t.Errorf("ScrollUp: OffsetY = %d, want 3", scroll.offsetY)
	}

	// Test scroll bounds
	scroll.ScrollDown(1000, 100, 10)
	if scroll.offsetY > 90 {
		t.Errorf("ScrollDown should limit to maxOffset = %d, got %d", 90, scroll.offsetY)
	}

	// Test scroll up to 0
	scroll.ScrollUp(1000)
	if scroll.offsetY != 0 {
		t.Errorf("ScrollUp should limit to 0, got %d", scroll.offsetY)
	}
}

func TestScrollViewInMarkup(t *testing.T) {
	const markup = `
<scrollview height="5" scrollbar="true">
	<vstack>
		<text>Line 1</text>
		<text>Line 2</text>
		<text>Line 3</text>
		<text>Line 4</text>
		<text>Line 5</text>
		<text>Line 6</text>
		<text>Line 7</text>
	</vstack>
</scrollview>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 80, 24)
	golden.RequireEqual(t, output)
}

func TestScrollViewWithSlot(t *testing.T) {
	const markup = `
<box border="rounded">
	<slot name="scrollable" />
</box>
`

	var lines []Element
	for i := 1; i <= 20; i++ {
		lines = append(lines, NewText(fmt.Sprintf("Line %d", i)))
	}

	scroll := NewScrollView(NewVStack(lines...)).
		Height(NewFixedConstraint(10)).
		Offset(0, 5)

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	slots := map[string]Element{
		"scrollable": scroll,
	}

	output := tmpl.RenderWithSlots(nil, slots, 80, 24)
	golden.RequireEqual(t, output)
}

// Test ScrollView chaining methods.
func TestScrollViewWithMethods(t *testing.T) {
	scroll := NewScrollView(NewText("test"))

	scroll.Vertical(false)
	scroll.Horizontal(true)
	scroll.Scrollbar(false)
	scroll.Width(NewFixedConstraint(10))

	// Test Children
	children := scroll.Children()
	if len(children) != 1 {
		t.Error("ScrollView Children should return child")
	}
}

// Test horizontal scrolling methods.
func TestScrollHorizontalMethods(t *testing.T) {
	scroll := NewScrollView(NewText("test"))

	// Test ScrollLeft
	scroll.offsetX = 10
	scroll.ScrollLeft(5)
	if scroll.offsetX != 5 {
		t.Errorf("ScrollLeft: expected offset 5, got %d", scroll.offsetX)
	}

	scroll.ScrollLeft(10)
	if scroll.offsetX != 0 {
		t.Errorf("ScrollLeft should clamp to 0, got %d", scroll.offsetX)
	}

	// Test ScrollRight
	scroll.offsetX = 0
	scroll.ScrollRight(5, 100, 10)
	if scroll.offsetX != 5 {
		t.Errorf("ScrollRight: expected offset 5, got %d", scroll.offsetX)
	}

	scroll.ScrollRight(100, 100, 10)
	if scroll.offsetX > 90 {
		t.Errorf("ScrollRight should clamp to maxOffset 90, got %d", scroll.offsetX)
	}
}

// Test ContentSize with nil child.
func TestScrollViewContentSizeNil(t *testing.T) {
	scroll := &ScrollView{child: nil}
	size := scroll.ContentSize()
	if size.Width != 0 || size.Height != 0 {
		t.Error("ContentSize with nil child should return zero size")
	}
}

// Test horizontal scrollbar rendering.
func TestHorizontalScrollbar(t *testing.T) {
	// Create wide content
	var items []Element
	for i := 0; i < 20; i++ {
		items = append(items, NewText("Word"))
	}

	scroll := NewScrollView(NewHStack(items...).Spacing(1)).
		Width(NewFixedConstraint(30)).
		Height(NewFixedConstraint(5)).
		Horizontal(true).
		Vertical(false).
		Offset(10, 0)

	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  30,
		MinHeight: 0,
		MaxHeight: 5,
	}

	size := scroll.Layout(constraints)
	if size.Width != 30 {
		t.Errorf("Horizontal scroll layout width = %d, want 30", size.Width)
	}

	// Test rendering to actually trigger drawHorizontalScrollbar
	buf := uv.NewScreenBuffer(size.Width, size.Height)
	area := uv.Rect(0, 0, size.Width, size.Height)
	scroll.Draw(buf, area)

	output := buf.Render()
	if len(output) == 0 {
		t.Error("Should render horizontal scrollbar")
	}
}

// Test children with nil child.
func TestScrollViewChildrenNil(t *testing.T) {
	scroll := &ScrollView{child: nil}
	if scroll.Children() != nil {
		t.Error("ScrollView Children with nil child should return nil")
	}
}
