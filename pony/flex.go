package pony

import uv "github.com/charmbracelet/ultraviolet"

// Flex represents an element wrapper that supports flex-grow and flex-shrink.
// Use this to make elements flexible within VStack or HStack.
type Flex struct {
	BaseElement
	Child  Element
	Grow   int // flex-grow: how much to grow relative to siblings (default 0 = no grow)
	Shrink int // flex-shrink: how much to shrink relative to siblings (default 1)
	Basis  int // flex-basis: initial size before flex calculation (default 0 = auto)
}

var _ Element = (*Flex)(nil)

// NewFlex creates a new flex wrapper.
func NewFlex(child Element) *Flex {
	return &Flex{
		Child:  child,
		Grow:   0,
		Shrink: 1,
		Basis:  0,
	}
}

// WithGrow sets the flex-grow and returns the flex for chaining.
func (f *Flex) WithGrow(grow int) *Flex {
	f.Grow = grow
	return f
}

// WithShrink sets the flex-shrink and returns the flex for chaining.
func (f *Flex) WithShrink(shrink int) *Flex {
	f.Shrink = shrink
	return f
}

// WithBasis sets the flex-basis and returns the flex for chaining.
func (f *Flex) WithBasis(basis int) *Flex {
	f.Basis = basis
	return f
}

// Draw renders the flex child.
func (f *Flex) Draw(scr uv.Screen, area uv.Rectangle) {
	f.SetBounds(area)

	if f.Child != nil {
		f.Child.Draw(scr, area)
	}
}

// Layout calculates the flex child size.
func (f *Flex) Layout(constraints Constraints) Size {
	if f.Child == nil {
		return Size{Width: 0, Height: 0}
	}

	// If basis is set, use it as the initial size
	if f.Basis > 0 {
		// Create constraints with basis as preferred size
		flexConstraints := constraints
		flexConstraints.MinWidth = min(f.Basis, constraints.MaxWidth)
		flexConstraints.MinHeight = min(f.Basis, constraints.MaxHeight)
		return f.Child.Layout(flexConstraints)
	}

	return f.Child.Layout(constraints)
}

// Children returns the child element.
func (f *Flex) Children() []Element {
	if f.Child == nil {
		return nil
	}
	return []Element{f.Child}
}

// GetFlexGrow returns the flex-grow value for an element.
// Returns 0 if the element is not a Flex wrapper.
func GetFlexGrow(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.Grow
	}
	// Check if it's a flexible spacer
	if spacer, ok := elem.(*Spacer); ok && spacer.Size == 0 {
		return 1
	}
	return 0
}

// GetFlexShrink returns the flex-shrink value for an element.
// Returns 1 if the element is not a Flex wrapper.
func GetFlexShrink(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.Shrink
	}
	return 1
}

// GetFlexBasis returns the flex-basis value for an element.
// Returns 0 (auto) if the element is not a Flex wrapper.
func GetFlexBasis(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.Basis
	}
	return 0
}

// IsFlexible returns true if the element can grow.
func IsFlexible(elem Element) bool {
	return GetFlexGrow(elem) > 0
}
