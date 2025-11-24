package pony

import (
	"testing"
)

func TestNestedInteractiveElementsInSlot(t *testing.T) {
	// This tests the exact scenario from the interactive-form example:
	// Buttons inside a component that's rendered into a slot

	tmpl := `
<vstack gap="1">
	<text id="header">Form</text>
	<slot name="button-bar" />
	<text id="footer">Footer</text>
</vstack>
`

	type Data struct{}
	template := MustParse[Data](tmpl)

	// Create buttons like ButtonBar does
	btn1 := NewButton("Submit")
	btn1.SetID("submit-btn")

	btn2 := NewButton("Clear")
	btn2.SetID("clear-btn")

	btn3 := NewButton("Quit")
	btn3.SetID("quit-btn")

	// Create an HStack containing the buttons (like ButtonBar.Render())
	buttonBar := NewHStack(btn1, btn2, btn3).WithGap(2)
	buttonBar.SetID("button-bar-hstack")

	slots := map[string]Element{
		"button-bar": buttonBar,
	}

	// Render with bounds
	_, boundsMap := template.RenderWithBounds(Data{}, slots, 80, 20)

	// Verify all nested buttons are registered
	tests := []struct {
		id          string
		description string
	}{
		{"header", "header text"},
		{"footer", "footer text"},
		{"button-bar-hstack", "button bar container"},
		{"submit-btn", "submit button inside slot"},
		{"clear-btn", "clear button inside slot"},
		{"quit-btn", "quit button inside slot"},
	}

	for _, tt := range tests {
		elem, ok := boundsMap.GetByID(tt.id)
		if !ok {
			t.Errorf("%s (%s) not found in bounds map", tt.description, tt.id)
			continue
		}

		bounds := elem.Bounds()
		if bounds.Dx() == 0 || bounds.Dy() == 0 {
			t.Errorf("%s (%s) has zero-size bounds", tt.description, tt.id)
		}
	}

	// Test hit testing on a button inside the slot
	submitBtn, ok := boundsMap.GetByID("submit-btn")
	if !ok {
		t.Fatal("submit-btn not found")
	}

	submitBounds := submitBtn.Bounds()
	centerX := submitBounds.Min.X + submitBounds.Dx()/2
	centerY := submitBounds.Min.Y + submitBounds.Dy()/2

	hit := boundsMap.HitTest(centerX, centerY)
	if hit == nil {
		t.Fatal("hit test failed on button inside slot")
	}

	// The hit could be the button itself or a child element
	// We just want to verify hit testing works
	t.Logf("Hit test on submit button returned: %s", hit.ID())
}

func TestMultipleNestedLevelsInSlots(t *testing.T) {
	// Test deeply nested structure: slot -> vstack -> hstack -> buttons

	tmpl := `
<vstack>
	<slot name="content" />
</vstack>
`

	type Data struct{}
	template := MustParse[Data](tmpl)

	// Create deeply nested structure
	btn1 := NewButton("Button 1")
	btn1.SetID("btn1")

	btn2 := NewButton("Button 2")
	btn2.SetID("btn2")

	buttonRow := NewHStack(btn1, btn2).WithGap(1)
	buttonRow.SetID("button-row")

	text := NewText("Label")
	text.SetID("label-text")

	content := NewVStack(
		text,
		buttonRow,
	).WithGap(1)
	content.SetID("content-vstack")

	slots := map[string]Element{
		"content": content,
	}

	// Render with bounds
	_, boundsMap := template.RenderWithBounds(Data{}, slots, 80, 20)

	// Verify all levels are registered
	expectedIDs := []string{
		"content-vstack",
		"label-text",
		"button-row",
		"btn1",
		"btn2",
	}

	for _, id := range expectedIDs {
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

	// Test that we can hit test buttons at any nesting level
	btn1Elem, ok := boundsMap.GetByID("btn1")
	if !ok {
		t.Fatal("btn1 not found")
	}

	btn1Bounds := btn1Elem.Bounds()
	hit := boundsMap.HitTest(btn1Bounds.Min.X+1, btn1Bounds.Min.Y+1)
	if hit == nil {
		t.Fatal("hit test failed on deeply nested button")
	}

	t.Logf("Hit test on deeply nested button returned: %s", hit.ID())
}
