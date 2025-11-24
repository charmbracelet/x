package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Element represents a renderable component in the TUI. Elements implement
// the UV Drawable interface and can be composed into trees.
type Element interface {
	uv.Drawable

	// Layout calculates the element's desired size within the given constraints.
	// It returns the actual size the element will occupy.
	Layout(constraints Constraints) Size

	// Children returns the child elements for container types.
	// Returns nil for leaf elements.
	Children() []Element

	// ID returns a unique identifier for this element.
	// Used for hit testing and event handling.
	ID() string

	// SetID sets the element's identifier.
	SetID(id string)

	// Bounds returns the element's last rendered screen coordinates.
	// Updated during Draw() and used for mouse hit testing.
	Bounds() uv.Rectangle

	// SetBounds records the element's rendered bounds.
	// This should be called at the start of Draw().
	SetBounds(bounds uv.Rectangle)
}

// Size represents dimensions in terminal cells.
type Size struct {
	Width  int
	Height int
}

// Constraints define the size constraints for layout calculations.
type Constraints struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// Constrain returns a size that satisfies the constraints.
func (c Constraints) Constrain(size Size) Size {
	w := size.Width
	h := size.Height

	if w < c.MinWidth {
		w = c.MinWidth
	}
	if w > c.MaxWidth {
		w = c.MaxWidth
	}
	if h < c.MinHeight {
		h = c.MinHeight
	}
	if h > c.MaxHeight {
		h = c.MaxHeight
	}

	return Size{Width: w, Height: h}
}

// Unbounded returns constraints with no limits.
func Unbounded() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  1<<31 - 1,
		MinHeight: 0,
		MaxHeight: 1<<31 - 1,
	}
}

// Fixed returns constraints for a fixed size.
func Fixed(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Constraint represents a size constraint that can be applied.
type Constraint interface {
	// Apply applies the constraint to the given available size.
	Apply(available int) int
}

// FixedConstraint represents a fixed size in cells.
type FixedConstraint int

// Apply returns the fixed size, clamped to available space.
func (f FixedConstraint) Apply(available int) int {
	if int(f) > available {
		return available
	}
	if f < 0 {
		return 0
	}
	return int(f)
}

// PercentConstraint represents a percentage of available space (0-100).
type PercentConstraint int

// Apply returns the percentage of available space.
func (p PercentConstraint) Apply(available int) int {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return available
	}
	return available * int(p) / 100
}

// AutoConstraint represents content-based sizing.
type AutoConstraint struct{}

// Apply returns the available size (will be calculated based on content).
func (a AutoConstraint) Apply(available int) int {
	return available
}

// Props is a map of properties passed to elements.
type Props map[string]string

// Get returns a property value or empty string if not found.
func (p Props) Get(key string) string {
	if p == nil {
		return ""
	}
	return p[key]
}

// GetOr returns a property value or the default if not found.
func (p Props) GetOr(key, defaultValue string) string {
	if p == nil {
		return defaultValue
	}
	if v, ok := p[key]; ok {
		return v
	}
	return defaultValue
}

// Has checks if a property exists.
func (p Props) Has(key string) bool {
	if p == nil {
		return false
	}
	_, ok := p[key]
	return ok
}
