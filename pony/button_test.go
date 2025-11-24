package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestButtonWithMethods(t *testing.T) {
	style := NewStyle().Fg(Hex("#00FF00")).Build()
	hoverStyle := NewStyle().Fg(Hex("#00FFFF")).Build()
	activeStyle := NewStyle().Fg(Hex("#FF00FF")).Build()

	btn := NewButton("Click Me").
		WithStyle(style).
		WithHoverStyle(hoverStyle).
		WithActiveStyle(activeStyle).
		WithBorder("rounded").
		WithPadding(2).
		WithWidth(NewFixedConstraint(20)).
		WithHeight(NewFixedConstraint(3))

	if btn.Text != "Click Me" {
		t.Errorf("Text = %s, want Click Me", btn.Text)
	}

	if btn.Style != style {
		t.Error("Style not set")
	}

	if btn.HoverStyle != hoverStyle {
		t.Error("HoverStyle not set")
	}

	if btn.ActiveStyle != activeStyle {
		t.Error("ActiveStyle not set")
	}

	if btn.Border != BorderRounded {
		t.Errorf("Border = %s, want %s", btn.Border, BorderRounded)
	}

	if btn.Padding != 2 {
		t.Errorf("Padding = %d, want 2", btn.Padding)
	}

	if btn.Width.IsAuto() {
		t.Error("Width constraint should not be auto")
	}

	if btn.Height.IsAuto() {
		t.Error("Height constraint should not be auto")
	}
}

func TestButtonLayout(t *testing.T) {
	btn := NewButton("Submit")
	size := btn.Layout(Fixed(40, 10))

	if size.Width == 0 || size.Height == 0 {
		t.Errorf("Button size is zero: %v", size)
	}
}

func TestButtonDraw(t *testing.T) {
	btn := NewButton("Test").
		WithStyle(NewStyle().Bold().Build()).
		WithBorder("rounded").
		WithPadding(1)

	scr := uv.NewScreenBuffer(30, 10)
	area := uv.Rect(0, 0, 30, 10)

	btn.Draw(scr, area)

	bounds := btn.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Error("Button bounds not set after draw")
	}
}

func TestButtonChildren(t *testing.T) {
	btn := NewButton("Test")
	children := btn.Children()

	if children != nil {
		t.Error("Button should have nil children")
	}
}

func TestBoundsMapAllElements(t *testing.T) {
	text1 := NewText("Hello")
	text1.SetID("text1")

	text2 := NewText("World")
	text2.SetID("text2")

	vstack := NewVStack(text1, text2)
	vstack.SetID("vstack")

	constraints := Fixed(20, 10)
	vstack.Layout(constraints)

	scr := uv.NewScreenBuffer(20, 10)
	area := uv.Rect(0, 0, 20, 10)
	vstack.Draw(scr, area)

	boundsMap := NewBoundsMap()
	walkAndRegister(vstack, boundsMap)

	allElements := boundsMap.AllElements()
	if len(allElements) == 0 {
		t.Error("AllElements returned empty list")
	}

	if len(allElements) != 3 {
		t.Errorf("AllElements returned %d elements, want 3", len(allElements))
	}

	for _, eb := range allElements {
		if eb.Element == nil {
			t.Error("Element in AllElements is nil")
		}
		if eb.Bounds.Dx() == 0 && eb.Bounds.Dy() == 0 {
			t.Errorf("Element %s has zero bounds", eb.Element.ID())
		}
	}
}
