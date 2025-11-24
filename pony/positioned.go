package pony

import uv "github.com/charmbracelet/ultraviolet"

// Positioned represents an absolutely positioned element.
// The element is positioned at specific coordinates relative to its parent.
type Positioned struct {
	BaseElement
	Child  Element
	X      int // X position (cells from left)
	Y      int // Y position (cells from top)
	Right  int // Distance from right edge (if >= 0, overrides X)
	Bottom int // Distance from bottom edge (if >= 0, overrides Y)
	Width  SizeConstraint
	Height SizeConstraint
}

var _ Element = (*Positioned)(nil)

// NewPositioned creates a new absolutely positioned element.
func NewPositioned(child Element, x, y int) *Positioned {
	return &Positioned{
		Child:  child,
		X:      x,
		Y:      y,
		Right:  -1,
		Bottom: -1,
	}
}

// WithRight sets the right edge distance and returns the positioned element for chaining.
// When set (>= 0), this overrides the X position.
func (p *Positioned) WithRight(right int) *Positioned {
	p.Right = right
	return p
}

// WithBottom sets the bottom edge distance and returns the positioned element for chaining.
// When set (>= 0), this overrides the Y position.
func (p *Positioned) WithBottom(bottom int) *Positioned {
	p.Bottom = bottom
	return p
}

// WithWidth sets the width constraint and returns the positioned element for chaining.
func (p *Positioned) WithWidth(width SizeConstraint) *Positioned {
	p.Width = width
	return p
}

// WithHeight sets the height constraint and returns the positioned element for chaining.
func (p *Positioned) WithHeight(height SizeConstraint) *Positioned {
	p.Height = height
	return p
}

// Draw renders the positioned element.
func (p *Positioned) Draw(scr uv.Screen, area uv.Rectangle) {
	p.SetBounds(area)

	if p.Child == nil {
		return
	}

	// Calculate child size
	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  area.Dx(),
		MinHeight: 0,
		MaxHeight: area.Dy(),
	}

	// Apply width/height constraints if specified
	if !p.Width.IsAuto() {
		width := p.Width.Apply(area.Dx(), area.Dx())
		constraints.MinWidth = width
		constraints.MaxWidth = width
	}

	if !p.Height.IsAuto() {
		height := p.Height.Apply(area.Dy(), area.Dy())
		constraints.MinHeight = height
		constraints.MaxHeight = height
	}

	childSize := p.Child.Layout(constraints)

	// Calculate position based on positioning constraints
	var childArea uv.Rectangle

	// Handle right/bottom positioning using UV layout helpers
	if p.Right >= 0 && p.Bottom >= 0 {
		// Both right and bottom are set - position from bottom-right corner
		childArea = uv.BottomRightRect(area, childSize.Width+p.Right, childSize.Height+p.Bottom)
		// Adjust for the offset
		childArea.Min.X = childArea.Max.X - childSize.Width - p.Right
		childArea.Max.X = childArea.Max.X - p.Right
		childArea.Min.Y = childArea.Max.Y - childSize.Height - p.Bottom
		childArea.Max.Y = childArea.Max.Y - p.Bottom
	} else if p.Right >= 0 {
		// Right is set - position from right edge
		x := area.Max.X - p.Right - childSize.Width
		y := area.Min.Y + p.Y
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	} else if p.Bottom >= 0 {
		// Bottom is set - position from bottom edge
		x := area.Min.X + p.X
		y := area.Max.Y - p.Bottom - childSize.Height
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	} else {
		// Standard X/Y positioning from top-left
		x := area.Min.X + p.X
		y := area.Min.Y + p.Y
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	}

	// Ensure the area is within parent bounds
	childArea = childArea.Intersect(area)

	p.Child.Draw(scr, childArea)
}

// Layout calculates the positioned element size.
// Positioned elements don't affect parent layout - they return 0 size.
func (p *Positioned) Layout(_ Constraints) Size {
	// Positioned elements are taken out of normal flow
	return Size{Width: 0, Height: 0}
}

// Children returns the child element.
func (p *Positioned) Children() []Element {
	if p.Child == nil {
		return nil
	}
	return []Element{p.Child}
}
