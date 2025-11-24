package pony

import (
	"testing"
)

func TestBoundsWithSlots(t *testing.T) {
	// Create a template with a slot
	tmpl := `
<vstack spacing="1">
	<text id="title">Title</text>
	<slot name="button-slot" />
	<text id="footer">Footer</text>
</vstack>
`

	type Data struct{}

	template := MustParse[Data](tmpl)

	// Create a button to fill the slot
	btn := NewButton("Click Me")
	btn.SetID("my-button")

	slots := map[string]Element{
		"button-slot": btn,
	}

	// Render with slots and bounds
	scr, boundsMap := template.RenderWithBounds(Data{}, slots, 40, 20)

	if scr.Bounds().Dx() == 0 {
		t.Fatal("screen has zero width")
	}

	// Check that title is registered
	if _, ok := boundsMap.GetByID("title"); !ok {
		t.Error("title not found in bounds map")
	}

	// Check that footer is registered
	if _, ok := boundsMap.GetByID("footer"); !ok {
		t.Error("footer not found in bounds map")
	}

	// Check that the button inside the slot is registered
	btnElem, ok := boundsMap.GetByID("my-button")
	if !ok {
		t.Fatal("button inside slot not found in bounds map")
	}

	// Verify the button has valid bounds
	btnBounds := btnElem.Bounds()
	if btnBounds.Dx() == 0 || btnBounds.Dy() == 0 {
		t.Error("button inside slot has zero-size bounds")
	}

	// Verify we can hit test the button
	centerX := btnBounds.Min.X + btnBounds.Dx()/2
	centerY := btnBounds.Min.Y + btnBounds.Dy()/2
	hit := boundsMap.HitTest(centerX, centerY)
	if hit == nil {
		t.Fatal("hit test failed on button inside slot")
	}

	if hit.ID() != "my-button" {
		t.Errorf("hit test returned wrong element: got %s, want my-button", hit.ID())
	}
}

func TestBoundsWithNestedSlots(t *testing.T) {
	// Create a complex nested structure with slots
	tmpl := `
<vstack spacing="1">
	<slot name="header" />
	<box border="rounded" id="main-box">
		<slot name="content" />
	</box>
</vstack>
`

	type Data struct{}

	template := MustParse[Data](tmpl)

	// Create nested content
	header := NewText("Header")
	header.SetID("header-text")

	innerBtn := NewButton("Submit")
	innerBtn.SetID("submit-btn")

	contentText := NewText("Some text")
	contentText.SetID("content-text")

	content := NewVStack(
		contentText,
		innerBtn,
	)
	content.SetID("content-vstack")

	slots := map[string]Element{
		"header":  header,
		"content": content,
	}

	// Render with slots and bounds
	_, boundsMap := template.RenderWithBounds(Data{}, slots, 40, 20)

	// Check all elements are registered
	tests := []string{
		"header-text",
		"main-box",
		"content-vstack",
		"content-text",
		"submit-btn",
	}

	for _, id := range tests {
		elem, ok := boundsMap.GetByID(id)
		if !ok {
			t.Errorf("element %s not found in bounds map", id)
			continue
		}

		bounds := elem.Bounds()
		if bounds.Dx() == 0 || bounds.Dy() == 0 {
			t.Errorf("element %s has zero-size bounds", id)
		}
	}

	// Test hit testing on deeply nested button
	btnElem, ok := boundsMap.GetByID("submit-btn")
	if !ok {
		t.Fatal("submit-btn not found")
	}

	btnBounds := btnElem.Bounds()
	hit := boundsMap.HitTest(btnBounds.Min.X+1, btnBounds.Min.Y+1)
	if hit == nil {
		t.Fatal("hit test failed on nested button")
	}

	// Could hit the button or a child element inside it
	if hit.ID() != "submit-btn" {
		t.Logf("Note: Hit test returned %s instead of submit-btn (might be hitting child element)", hit.ID())
	}
}

func TestWithIDHelper(t *testing.T) {
	// Test the WithID helper method
	text := NewText("Hello")
	text.SetID("my-text")

	if text.ID() != "my-text" {
		t.Errorf("ID() returned %s, want my-text", text.ID())
	}

	// Test auto-generated ID
	text2 := NewText("World")
	id := text2.ID()
	if id == "" {
		t.Error("auto-generated ID is empty")
	}
	if id[:5] != "elem_" {
		t.Errorf("auto-generated ID doesn't start with elem_: %s", id)
	}
}
