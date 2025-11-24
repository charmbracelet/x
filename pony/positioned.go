package pony

import uv "github.com/charmbracelet/ultraviolet"

// Positioned represents an absolutely positioned element.
// The element is positioned at specific coordinates relative to its parent.
type Positioned struct {
	BaseElement
	child  Element
	x      int // X position (cells from left)
	y      int // Y position (cells from top)
	right  int // Distance from right edge (if >= 0, overrides X)
	bottom int // Distance from bottom edge (if >= 0, overrides Y)
	width  SizeConstraint
	height SizeConstraint
}

var _ Element = (*Positioned)(nil)

// NewPositioned creates a new absolutely positioned element.
func NewPositioned(child Element, x, y int) *Positioned {
	return &Positioned{
		child:  child,
		x:      x,
		y:      y,
		right:  -1,
		bottom: -1,
	}
}

// Right sets the right edge distance and returns the positioned element for chaining.
// When set (>= 0), this overrides the X position.
func (p *Positioned) Right(right int) *Positioned {
	p.right = right
	return p
}

// Bottom sets the bottom edge distance and returns the positioned element for chaining.
// When set (>= 0), this overrides the Y position.
func (p *Positioned) Bottom(bottom int) *Positioned {
	p.bottom = bottom
	return p
}

// Width sets the width constraint and returns the positioned element for chaining.
func (p *Positioned) Width(width SizeConstraint) *Positioned {
	p.width = width
	return p
}

// Height sets the height constraint and returns the positioned element for chaining.
func (p *Positioned) Height(height SizeConstraint) *Positioned {
	p.height = height
	return p
}

// Draw renders the positioned element.
func (p *Positioned) Draw(scr uv.Screen, area uv.Rectangle) {
	p.SetBounds(area)

	if p.child == nil {
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
	if !p.width.IsAuto() {
		width := p.width.Apply(area.Dx(), area.Dx())
		constraints.MinWidth = width
		constraints.MaxWidth = width
	}

	if !p.height.IsAuto() {
		height := p.height.Apply(area.Dy(), area.Dy())
		constraints.MinHeight = height
		constraints.MaxHeight = height
	}

	childSize := p.child.Layout(constraints)

	// Calculate position based on positioning constraints
	var childArea uv.Rectangle

	// Handle right/bottom positioning using UV layout helpers
	if p.right >= 0 && p.bottom >= 0 {
		// Both right and bottom are set - position from bottom-right corner
		childArea = uv.BottomRightRect(area, childSize.Width+p.right, childSize.Height+p.bottom)
		// Adjust for the offset
		childArea.Min.X = childArea.Max.X - childSize.Width - p.right
		childArea.Max.X = childArea.Max.X - p.right
		childArea.Min.Y = childArea.Max.Y - childSize.Height - p.bottom
		childArea.Max.Y = childArea.Max.Y - p.bottom
	} else if p.right >= 0 {
		// Right is set - position from right edge
		x := area.Max.X - p.right - childSize.Width
		y := area.Min.Y + p.y
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	} else if p.bottom >= 0 {
		// Bottom is set - position from bottom edge
		x := area.Min.X + p.x
		y := area.Max.Y - p.bottom - childSize.Height
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	} else {
		// Standard X/Y positioning from top-left
		x := area.Min.X + p.x
		y := area.Min.Y + p.y
		childArea = uv.Rect(x, y, childSize.Width, childSize.Height)
	}

	// Ensure the area is within parent bounds
	childArea = childArea.Intersect(area)

	p.child.Draw(scr, childArea)
}

// Layout calculates the positioned element size.
// Positioned elements don't affect parent layout - they return 0 size.
func (p *Positioned) Layout(_ Constraints) Size {
	// Positioned elements are taken out of normal flow
	return Size{Width: 0, Height: 0}
}

// Children returns the child element.
func (p *Positioned) Children() []Element {
	if p.child == nil {
		return nil
	}
	return []Element{p.child}
}
