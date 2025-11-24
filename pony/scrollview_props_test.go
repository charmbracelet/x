package pony

import (
	"testing"
)

func TestScrollViewPropsWithOffset(t *testing.T) {
	tmpl := `<scrollview height="10" offset-y="5">
		<vstack>
			<text>Item 1</text>
			<text>Item 2</text>
			<text>Item 3</text>
		</vstack>
	</scrollview>`

	type Data struct{}
	template := MustParse[Data](tmpl)

	rendered := template.Render(Data{}, 40, 20)
	if rendered == "" {
		t.Error("expected rendered output")
	}

	// The scroll view should have been created with offset-y=5
	// We can verify this by rendering with bounds
	_, boundsMap := template.RenderWithBounds(Data{}, nil, 40, 20)

	// Check if any scroll view elements exist
	allElems := boundsMap.AllElements()

	foundScrollView := false
	for _, eb := range allElems {
		if sv, ok := eb.Element.(*ScrollView); ok {
			foundScrollView = true
			if sv.OffsetY != 5 {
				t.Errorf("expected OffsetY=5, got %d", sv.OffsetY)
			}
		}
	}

	if !foundScrollView {
		t.Error("expected to find a ScrollView element")
	}
}

func TestScrollViewPropsWithTemplateOffset(t *testing.T) {
	tmpl := `<scrollview height="10" offset-y="{{ .Offset }}">
		<vstack>
			<text>Item 1</text>
			<text>Item 2</text>
			<text>Item 3</text>
		</vstack>
	</scrollview>`

	type Data struct {
		Offset int
	}
	template := MustParse[Data](tmpl)

	// Test with offset from template data
	_, boundsMap := template.RenderWithBounds(Data{Offset: 8}, nil, 40, 20)

	// Find the scroll view and verify offset
	allElems := boundsMap.AllElements()
	foundScrollView := false
	for _, eb := range allElems {
		if sv, ok := eb.Element.(*ScrollView); ok {
			foundScrollView = true
			if sv.OffsetY != 8 {
				t.Errorf("expected OffsetY=8, got %d", sv.OffsetY)
			}
		}
	}

	if !foundScrollView {
		t.Error("expected to find a ScrollView element")
	}
}

func TestScrollViewPropsAllAttributes(t *testing.T) {
	tmpl := `<scrollview 
		id="my-scroll" 
		width="30" 
		height="10" 
		offset-x="2"
		offset-y="3"
		vertical="true"
		horizontal="false"
		scrollbar="true">
		<text>Content</text>
	</scrollview>`

	type Data struct{}
	template := MustParse[Data](tmpl)

	_, boundsMap := template.RenderWithBounds(Data{}, nil, 40, 20)

	// Verify the scroll view was registered with the correct ID
	elem, ok := boundsMap.GetByID("my-scroll")
	if !ok {
		t.Fatal("expected to find scroll view with id 'my-scroll'")
	}

	// Type assertion to verify it's a ScrollView
	sv, ok := elem.(*ScrollView)
	if !ok {
		t.Fatalf("expected *ScrollView, got %T", elem)
	}

	// Verify properties were applied
	if sv.OffsetX != 2 {
		t.Errorf("expected OffsetX=2, got %d", sv.OffsetX)
	}

	if sv.OffsetY != 3 {
		t.Errorf("expected OffsetY=3, got %d", sv.OffsetY)
	}

	if !sv.Vertical {
		t.Error("expected Vertical=true")
	}

	if sv.Horizontal {
		t.Error("expected Horizontal=false")
	}

	if !sv.ShowScrollbar {
		t.Error("expected ShowScrollbar=true")
	}
}
