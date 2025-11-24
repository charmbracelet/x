package pony

import uv "github.com/charmbracelet/ultraviolet"

// Flex represents an element wrapper that supports flex-grow and flex-shrink.
// Use this to make elements flexible within VStack or HStack.
type Flex struct {
	BaseElement
	child  Element
	grow   int // flex-grow: how much to grow relative to siblings (default 0 = no grow)
	shrink int // flex-shrink: how much to shrink relative to siblings (default 1)
	basis  int // flex-basis: initial size before flex calculation (default 0 = auto)
}

var _ Element = (*Flex)(nil)

// NewFlex creates a new flex wrapper.
func NewFlex(child Element) *Flex {
	return &Flex{
		child:  child,
		grow:   0,
		shrink: 1,
		basis:  0,
	}
}

// Grow sets the flex-grow and returns the flex for chaining.
func (f *Flex) Grow(grow int) *Flex {
	f.grow = grow
	return f
}

// Shrink sets the flex-shrink and returns the flex for chaining.
func (f *Flex) Shrink(shrink int) *Flex {
	f.shrink = shrink
	return f
}

// Basis sets the flex-basis and returns the flex for chaining.
func (f *Flex) Basis(basis int) *Flex {
	f.basis = basis
	return f
}

// Draw renders the flex child.
func (f *Flex) Draw(scr uv.Screen, area uv.Rectangle) {
	f.SetBounds(area)

	if f.child != nil {
		f.child.Draw(scr, area)
	}
}

// Layout calculates the flex child size.
func (f *Flex) Layout(constraints Constraints) Size {
	if f.child == nil {
		return Size{Width: 0, Height: 0}
	}

	// If basis is set, use it as the initial size
	if f.basis > 0 {
		// Create constraints with basis as preferred size
		flexConstraints := constraints
		flexConstraints.MinWidth = min(f.basis, constraints.MaxWidth)
		flexConstraints.MinHeight = min(f.basis, constraints.MaxHeight)
		return f.child.Layout(flexConstraints)
	}

	return f.child.Layout(constraints)
}

// Children returns the child element.
func (f *Flex) Children() []Element {
	if f.child == nil {
		return nil
	}
	return []Element{f.child}
}

// GetFlexGrow returns the flex-grow value for an element.
// Returns 0 if the element is not a Flex wrapper.
func GetFlexGrow(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.grow
	}
	// Check if it's a flexible spacer
	if spacer, ok := elem.(*Spacer); ok && spacer.fixedSize == 0 {
		return 1
	}
	return 0
}

// GetFlexShrink returns the flex-shrink value for an element.
// Returns 1 if the element is not a Flex wrapper.
func GetFlexShrink(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.shrink
	}
	return 1
}

// GetFlexBasis returns the flex-basis value for an element.
// Returns 0 (auto) if the element is not a Flex wrapper.
func GetFlexBasis(elem Element) int {
	if flex, ok := elem.(*Flex); ok {
		return flex.basis
	}
	return 0
}

// IsFlexible returns true if the element can grow.
func IsFlexible(elem Element) bool {
	return GetFlexGrow(elem) > 0
}
