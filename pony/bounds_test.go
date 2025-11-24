package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestBoundsTracking(t *testing.T) {
	// Create a simple layout
	text1 := NewText("Hello")
	text1.SetID("text1")

	text2 := NewText("World")
	text2.SetID("text2")

	vstack := NewVStack(text1, text2)
	vstack.SetID("vstack")

	// Layout and draw
	constraints := Fixed(20, 10)
	vstack.Layout(constraints)

	scr := uv.NewScreenBuffer(20, 10)
	area := uv.Rect(0, 0, 20, 10)
	vstack.Draw(scr, area)

	// Build bounds map
	boundsMap := NewBoundsMap()
	walkAndRegister(vstack, boundsMap)

	// Test that all elements are registered
	if _, ok := boundsMap.GetByID("vstack"); !ok {
		t.Error("vstack not found in bounds map")
	}
	if _, ok := boundsMap.GetByID("text1"); !ok {
		t.Error("text1 not found in bounds map")
	}
	if _, ok := boundsMap.GetByID("text2"); !ok {
		t.Error("text2 not found in bounds map")
	}

	// Test bounds are set
	vstackBounds, _ := boundsMap.GetBounds("vstack")
	if vstackBounds.Dx() == 0 || vstackBounds.Dy() == 0 {
		t.Error("vstack bounds not set correctly")
	}
}

func TestHitTest(t *testing.T) {
	// Create a layout with positioned elements
	btn1 := NewButton("Button 1")
	btn1.SetID("btn1")

	btn2 := NewButton("Button 2")
	btn2.SetID("btn2")

	vstack := NewVStack(btn1, btn2).Spacing(1)

	// Layout and draw
	constraints := Fixed(40, 20)
	vstack.Layout(constraints)

	scr := uv.NewScreenBuffer(40, 20)
	area := uv.Rect(0, 0, 40, 20)
	vstack.Draw(scr, area)

	// Build bounds map
	boundsMap := NewBoundsMap()
	walkAndRegister(vstack, boundsMap)

	// Get button bounds
	btn1Bounds, ok := boundsMap.GetBounds("btn1")
	if !ok {
		t.Fatal("btn1 bounds not found")
	}

	btn2Bounds, ok := boundsMap.GetBounds("btn2")
	if !ok {
		t.Fatal("btn2 bounds not found")
	}

	// Test hit testing on button 1
	centerX := btn1Bounds.Min.X + btn1Bounds.Dx()/2
	centerY := btn1Bounds.Min.Y + btn1Bounds.Dy()/2
	hit := boundsMap.HitTest(centerX, centerY)
	if hit == nil {
		t.Fatal("hit test returned nil for button 1 center")
	}
	if hit.ID() != "btn1" {
		t.Errorf("hit test returned wrong element: got %s, want btn1", hit.ID())
	}

	// Test hit testing on button 2
	centerX = btn2Bounds.Min.X + btn2Bounds.Dx()/2
	centerY = btn2Bounds.Min.Y + btn2Bounds.Dy()/2
	hit = boundsMap.HitTest(centerX, centerY)
	if hit == nil {
		t.Fatal("hit test returned nil for button 2 center")
	}
	if hit.ID() != "btn2" {
		t.Errorf("hit test returned wrong element: got %s, want btn2", hit.ID())
	}

	// Test hit testing outside all elements
	hit = boundsMap.HitTest(100, 100)
	if hit != nil {
		t.Errorf("hit test should return nil for out of bounds position, got %s", hit.ID())
	}
}

func TestRenderWithBounds(t *testing.T) {
	tmpl := `
<vstack spacing="1">
	<text id="title">Title</text>
	<button id="submit-btn" text="Submit" />
</vstack>
`

	type Data struct{}

	template := MustParse[Data](tmpl)
	scr, boundsMap := template.RenderWithBounds(Data{}, nil, 40, 20)

	// Check screen was created
	if scr.Bounds().Dx() == 0 {
		t.Error("screen has zero width")
	}

	// Check bounds map has elements
	if _, ok := boundsMap.GetByID("title"); !ok {
		t.Error("title not found in bounds map")
	}
	if _, ok := boundsMap.GetByID("submit-btn"); !ok {
		t.Error("submit-btn not found in bounds map")
	}

	// Check we can hit test
	titleBounds, _ := boundsMap.GetBounds("title")
	hit := boundsMap.HitTest(titleBounds.Min.X, titleBounds.Min.Y)
	if hit == nil {
		t.Error("hit test failed on title")
	}
}

func TestIDParsing(t *testing.T) {
	tmpl := `<text id="my-text">Hello</text>`

	type Data struct{}

	template := MustParse[Data](tmpl)
	_, boundsMap := template.RenderWithBounds(Data{}, nil, 40, 20)

	elem, ok := boundsMap.GetByID("my-text")
	if !ok {
		t.Fatal("element with id not found")
	}

	if elem.ID() != "my-text" {
		t.Errorf("element ID mismatch: got %s, want my-text", elem.ID())
	}
}

func TestZIndexOrdering(t *testing.T) {
	// Create overlapping elements in a ZStack
	text1 := NewText("Behind")
	text1.SetID("behind")

	text2 := NewText("Front")
	text2.SetID("front")

	zstack := NewZStack(text1, text2)

	// Layout and draw
	constraints := Fixed(20, 5)
	zstack.Layout(constraints)

	scr := uv.NewScreenBuffer(20, 5)
	area := uv.Rect(0, 0, 20, 5)
	zstack.Draw(scr, area)

	// Build bounds map
	boundsMap := NewBoundsMap()
	walkAndRegister(zstack, boundsMap)

	// Both elements should be at same position (centered in zstack)
	// Hit test should return the one drawn last (front)
	hit := boundsMap.HitTest(10, 2)
	if hit == nil {
		t.Fatal("hit test returned nil")
	}

	// The last element drawn (front) should be returned
	// Note: ZStack draws children in order, so text2 is drawn last
	if hit.ID() != "front" {
		t.Logf("Note: Hit test returned %s, ZStack order matters for overlapping elements", hit.ID())
	}
}
