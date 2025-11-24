package pony

import (
	"fmt"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// BoundsMap tracks all rendered elements and their positions for hit testing.
// It is immutable after creation and safe for concurrent reads.
type BoundsMap struct {
	elements   map[string]elementBounds
	byPosition []elementBounds // ordered for z-index (last = on top)
}

type elementBounds struct {
	id     string
	elem   Element
	bounds uv.Rectangle
}

// NewBoundsMap creates a new empty bounds map.
func NewBoundsMap() *BoundsMap {
	return &BoundsMap{
		elements: make(map[string]elementBounds),
	}
}

// Register records an element and its rendered bounds.
// This should be called during the render pass.
func (bm *BoundsMap) Register(elem Element, bounds uv.Rectangle) {
	eb := elementBounds{
		id:     elem.ID(),
		elem:   elem,
		bounds: bounds,
	}
	bm.elements[elem.ID()] = eb
	bm.byPosition = append(bm.byPosition, eb)
}

// HitTest returns the top-most element at the given screen coordinates.
// When multiple elements overlap at a point, it prefers elements with explicitly
// set IDs over auto-generated IDs (elem_*).
//
// This behavior is crucial for interactive components: when you click inside
// a component's rendered area, you want the component's ID, not its children's.
// For example, clicking anywhere in an Input component should return the Input's
// ID, not the Text or Box child inside it.
//
// To achieve this, set your component's ID on the root element it returns:
//
//	func (i *Input) Render() pony.Element {
//	    vstack := pony.NewVStack(...)
//	    vstack.SetID(i.ID())  // Pass through component ID
//	    return vstack
//	}
//
// Returns nil if no element is found at that position.
func (bm *BoundsMap) HitTest(x, y int) Element {
	var bestMatch Element
	var bestMatchHasExplicitID bool

	// Search from end (last drawn = on top)
	for i := len(bm.byPosition) - 1; i >= 0; i-- {
		eb := bm.byPosition[i]
		if pointInRect(x, y, eb.bounds) {
			// Check if this element has an explicit ID (not auto-generated)
			hasExplicitID := !strings.HasPrefix(eb.id, "elem_")

			// First match or better match (explicit ID preferred)
			if bestMatch == nil || (!bestMatchHasExplicitID && hasExplicitID) {
				bestMatch = eb.elem
				bestMatchHasExplicitID = hasExplicitID
			}

			// If we found an element with explicit ID, that's our best match
			if hasExplicitID {
				return bestMatch
			}
		}
	}

	return bestMatch
}

// HitTestAll returns all elements at the given screen coordinates,
// ordered from top to bottom (first element is visually on top).
//
// This is useful for nested interactive components like scroll views
// with clickable children, where you need to know both the child
// that was clicked and the parent containers.
//
// Example usage:
//
//	hits := boundsMap.HitTestAll(x, y)
//	for _, elem := range hits {
//	    switch elem.ID() {
//	    case "list-item-5":
//	        // Handle item click
//	    case "main-scroll-view":
//	        // Also track that we're in the scroll view
//	    }
//	}
//
// Returns empty slice if no elements are found at that position.
func (bm *BoundsMap) HitTestAll(x, y int) []Element {
	var hits []Element

	// Search from end (last drawn = on top)
	for i := len(bm.byPosition) - 1; i >= 0; i-- {
		eb := bm.byPosition[i]
		if pointInRect(x, y, eb.bounds) {
			hits = append(hits, eb.elem)
		}
	}

	return hits
}

// HitTestWithContainer returns the top element and the first parent
// container with an explicit ID. This is useful for scroll views with
// clickable items where you want to know both what was clicked and
// which container it's in.
//
// The "top" element is the visually topmost element at the coordinates.
// The "container" is the first element in the hit stack (after top) that
// has an explicit ID (not auto-generated with "elem_" prefix).
//
// Example usage:
//
//	top, container := boundsMap.HitTestWithContainer(x, y)
//	if top != nil {
//	    handleClick(top.ID())
//	}
//	if container != nil && container.ID() == "scroll-view" {
//	    // We know we clicked inside a scroll view
//	}
//
// Returns (nil, nil) if no element is found at that position.
func (bm *BoundsMap) HitTestWithContainer(x, y int) (top Element, container Element) {
	hits := bm.HitTestAll(x, y)
	if len(hits) == 0 {
		return nil, nil
	}

	top = hits[0]

	// Find first container (element with explicit ID that's not the top)
	for i := 1; i < len(hits); i++ {
		if !strings.HasPrefix(hits[i].ID(), "elem_") {
			container = hits[i]
			break
		}
	}

	return top, container
}

// GetByID retrieves an element by its ID.
func (bm *BoundsMap) GetByID(id string) (Element, bool) {
	eb, ok := bm.elements[id]
	return eb.elem, ok
}

// GetBounds returns the rendered bounds for an element by ID.
func (bm *BoundsMap) GetBounds(id string) (uv.Rectangle, bool) {
	eb, ok := bm.elements[id]
	return eb.bounds, ok
}

// AllElements returns all registered elements with their bounds.
func (bm *BoundsMap) AllElements() []ElementWithBounds {
	result := make([]ElementWithBounds, 0, len(bm.byPosition))
	for _, eb := range bm.byPosition {
		result = append(result, ElementWithBounds{
			Element: eb.elem,
			Bounds:  eb.bounds,
		})
	}
	return result
}

// ElementWithBounds pairs an element with its rendered bounds.
type ElementWithBounds struct {
	Element Element
	Bounds  uv.Rectangle
}

// pointInRect checks if a point is inside a rectangle.
func pointInRect(x, y int, rect uv.Rectangle) bool {
	return x >= rect.Min.X && x < rect.Max.X &&
		y >= rect.Min.Y && y < rect.Max.Y
}

// BaseElement provides common functionality for all elements.
// Elements should embed this to get ID and bounds tracking.
type BaseElement struct {
	id     string
	bounds uv.Rectangle
}

// ID returns the element's identifier.
// If no ID was explicitly set, returns a pointer-based ID.
func (b *BaseElement) ID() string {
	if b.id == "" {
		return fmt.Sprintf("elem_%p", b)
	}
	return b.id
}

// SetID sets the element's identifier.
func (b *BaseElement) SetID(id string) {
	b.id = id
}

// Bounds returns the element's last rendered bounds.
func (b *BaseElement) Bounds() uv.Rectangle {
	return b.bounds
}

// SetBounds records the element's rendered bounds.
// This should be called at the start of Draw().
func (b *BaseElement) SetBounds(bounds uv.Rectangle) {
	b.bounds = bounds
}

// walkAndRegister recursively walks an element tree and registers all elements.
func walkAndRegister(elem Element, bm *BoundsMap) {
	bm.Register(elem, elem.Bounds())

	for _, child := range elem.Children() {
		if child != nil {
			walkAndRegister(child, bm)
		}
	}
}
