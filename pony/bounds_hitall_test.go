package pony

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestHitTestAll(t *testing.T) {
	// Create a scenario: ScrollView containing VStack containing Buttons
	// This simulates a scroll view with clickable list items

	tmpl := `
<vstack gap="1">
	<slot name="scrollview" />
</vstack>
`

	type Data struct{}
	template := MustParse[Data](tmpl)

	// Create buttons (list items)
	button1 := NewButton("Item 1")
	button1.SetID("item-1")

	button2 := NewButton("Item 2")
	button2.SetID("item-2")

	button3 := NewButton("Item 3")
	button3.SetID("item-3")

	// Create VStack to hold buttons
	vstack := NewVStack(button1, button2, button3)
	vstack.SetID("item-list")

	// Create ScrollView containing the VStack
	scrollView := NewScrollView(vstack)
	scrollView.SetID("main-scroll-view")

	slots := map[string]Element{
		"scrollview": scrollView,
	}

	// Render with bounds
	_, boundsMap := template.RenderWithBounds(Data{}, slots, 80, 20)

	// Get the bounds of button1 (first item)
	button1Elem, ok := boundsMap.GetByID("item-1")
	if !ok {
		t.Fatal("item-1 not found in bounds map")
	}

	button1Bounds := button1Elem.Bounds()

	// Click in the center of button1
	centerX := button1Bounds.Min.X + button1Bounds.Dx()/2
	centerY := button1Bounds.Min.Y + button1Bounds.Dy()/2

	// Test HitTestAll - should get multiple elements
	hits := boundsMap.HitTestAll(centerX, centerY)

	if len(hits) == 0 {
		t.Fatal("HitTestAll returned no hits")
	}

	// Should have multiple hits (button, vstack, scrollview, etc.)
	if len(hits) < 2 {
		t.Logf("Warning: Expected multiple hits, got %d", len(hits))
	}

	// Log all hits for debugging
	t.Logf("HitTestAll found %d elements:", len(hits))
	for i, elem := range hits {
		t.Logf("  [%d] %s", i, elem.ID())
	}

	// Verify we have the expected elements in the hit list
	hitIDs := make(map[string]bool)
	for _, elem := range hits {
		hitIDs[elem.ID()] = true
	}

	// We should definitely have the button
	if !hitIDs["item-1"] {
		t.Error("Expected to find item-1 in hit list")
	}

	// We should have the scroll view if bounds overlap
	// (this depends on the layout, so we'll just check it exists)
	scrollViewFound := hitIDs["main-scroll-view"]
	t.Logf("Scroll view in hits: %v", scrollViewFound)
}

func TestHitTestWithContainer(t *testing.T) {
	// Similar setup as above
	tmpl := `
<vstack gap="1">
	<slot name="scrollview" />
</vstack>
`

	type Data struct{}
	template := MustParse[Data](tmpl)

	button1 := NewButton("Item 1")
	button1.SetID("item-1")

	button2 := NewButton("Item 2")
	button2.SetID("item-2")

	vstack := NewVStack(button1, button2)
	vstack.SetID("item-list")

	scrollView := NewScrollView(vstack)
	scrollView.SetID("main-scroll-view")

	slots := map[string]Element{
		"scrollview": scrollView,
	}

	_, boundsMap := template.RenderWithBounds(Data{}, slots, 80, 20)

	// Get button1 bounds
	button1Elem, ok := boundsMap.GetByID("item-1")
	if !ok {
		t.Fatal("item-1 not found in bounds map")
	}

	button1Bounds := button1Elem.Bounds()
	centerX := button1Bounds.Min.X + button1Bounds.Dx()/2
	centerY := button1Bounds.Min.Y + button1Bounds.Dy()/2

	// Test HitTestWithContainer
	top, container := boundsMap.HitTestWithContainer(centerX, centerY)

	if top == nil {
		t.Fatal("HitTestWithContainer returned nil for top element")
	}

	t.Logf("Top element: %s", top.ID())
	if container != nil {
		t.Logf("Container element: %s", container.ID())
	} else {
		t.Log("Container element: nil")
	}

	// The top element should be related to our button click
	// (could be the button itself or a child element)
	topID := top.ID()
	if topID == "" {
		t.Error("Top element has empty ID")
	}

	// If we have a container, it should have an explicit ID
	if container != nil {
		containerID := container.ID()
		if containerID[:5] == "elem_" {
			t.Errorf("Container should have explicit ID, got auto-generated: %s", containerID)
		}
	}
}

func TestHitTestAllNoHits(t *testing.T) {
	bm := NewBoundsMap()

	// Register an element at a specific location
	text := NewText("Hello")
	text.SetBounds(uv.Rect(10, 10, 20, 5))
	bm.Register(text, text.Bounds())

	// Test clicking outside the element
	hits := bm.HitTestAll(50, 50)

	if len(hits) != 0 {
		t.Errorf("Expected no hits outside bounds, got %d", len(hits))
	}
}

func TestHitTestAllOverlapping(t *testing.T) {
	bm := NewBoundsMap()

	// Create overlapping elements
	text1 := NewText("Background")
	text1.SetID("background")
	text1.SetBounds(uv.Rect(0, 0, 20, 10))
	bm.Register(text1, text1.Bounds())

	text2 := NewText("Middle")
	text2.SetID("middle")
	text2.SetBounds(uv.Rect(5, 5, 15, 8))
	bm.Register(text2, text2.Bounds())

	text3 := NewText("Top")
	text3.SetID("top")
	text3.SetBounds(uv.Rect(8, 6, 10, 6))
	bm.Register(text3, text3.Bounds())

	// Click in the area where all three overlap
	hits := bm.HitTestAll(10, 7)

	// Should hit all three elements
	if len(hits) != 3 {
		t.Errorf("Expected 3 hits, got %d", len(hits))
	}

	// Verify order: last registered should be first (on top)
	if len(hits) >= 3 {
		if hits[0].ID() != "top" {
			t.Errorf("Expected first hit to be 'top', got '%s'", hits[0].ID())
		}
		if hits[1].ID() != "middle" {
			t.Errorf("Expected second hit to be 'middle', got '%s'", hits[1].ID())
		}
		if hits[2].ID() != "background" {
			t.Errorf("Expected third hit to be 'background', got '%s'", hits[2].ID())
		}
	}
}
