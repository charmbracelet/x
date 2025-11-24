package pony

import uv "github.com/charmbracelet/ultraviolet"

// ZStack represents a layered stack container where children are drawn on top of each other.
// Later children in the stack are drawn on top of earlier children.
type ZStack struct {
	BaseElement
	items             []Element
	width             SizeConstraint
	height            SizeConstraint
	alignment         string // Horizontal alignment: leading, center, trailing
	verticalAlignment string // Vertical alignment: top, center, bottom
}

var _ Element = (*ZStack)(nil)

// NewZStack creates a new layered stack.
func NewZStack(children ...Element) *ZStack {
	return &ZStack{
		items:             children,
		alignment:         AlignmentCenter,
		verticalAlignment: AlignmentCenter,
	}
}

// Alignment sets the horizontal alignment and returns the zstack for chaining.
func (z *ZStack) Alignment(alignment string) *ZStack {
	z.alignment = alignment
	return z
}

// VerticalAlignment sets the vertical alignment and returns the zstack for chaining.
func (z *ZStack) VerticalAlignment(alignment string) *ZStack {
	z.verticalAlignment = alignment
	return z
}

// Width sets the width constraint and returns the zstack for chaining.
func (z *ZStack) Width(width SizeConstraint) *ZStack {
	z.width = width
	return z
}

// Height sets the height constraint and returns the zstack for chaining.
func (z *ZStack) Height(height SizeConstraint) *ZStack {
	z.height = height
	return z
}

// Draw renders the layered stack to the screen.
func (z *ZStack) Draw(scr uv.Screen, area uv.Rectangle) {
	z.SetBounds(area)

	if len(z.items) == 0 {
		return
	}

	// Layout all children first to get their sizes
	childConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  area.Dx(),
		MinHeight: 0,
		MaxHeight: area.Dy(),
	}

	childSizes := make([]Size, len(z.items))
	for i, child := range z.items {
		childSizes[i] = child.Layout(childConstraints)
	}

	// Draw each child in order (later children draw on top)
	for i, child := range z.items {
		// Positioned elements handle their own layout and positioning
		if _, isPositioned := child.(*Positioned); isPositioned {
			child.Draw(scr, area)
			continue
		}

		childSize := childSizes[i]

		// Calculate child area based on alignment using UV layout helpers
		var childArea uv.Rectangle

		// Determine positioning based on horizontal and vertical alignment
		switch {
		case z.alignment == AlignmentLeading && z.verticalAlignment == AlignmentTop:
			childArea = uv.TopLeftRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentCenter && z.verticalAlignment == AlignmentTop:
			childArea = uv.TopCenterRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentTrailing && z.verticalAlignment == AlignmentTop:
			childArea = uv.TopRightRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentLeading && z.verticalAlignment == AlignmentCenter:
			childArea = uv.LeftCenterRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentCenter && z.verticalAlignment == AlignmentCenter:
			childArea = uv.CenterRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentTrailing && z.verticalAlignment == AlignmentCenter:
			childArea = uv.RightCenterRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentLeading && z.verticalAlignment == AlignmentBottom:
			childArea = uv.BottomLeftRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentCenter && z.verticalAlignment == AlignmentBottom:
			childArea = uv.BottomCenterRect(area, childSize.Width, childSize.Height)
		case z.alignment == AlignmentTrailing && z.verticalAlignment == AlignmentBottom:
			childArea = uv.BottomRightRect(area, childSize.Width, childSize.Height)
		default:
			// Fallback to top-left
			childArea = uv.TopLeftRect(area, childSize.Width, childSize.Height)
		}

		child.Draw(scr, childArea)
	}
}

// Layout calculates the total size of the layered stack.
// ZStack takes the maximum width and height of all children.
func (z *ZStack) Layout(constraints Constraints) Size {
	if len(z.items) == 0 {
		return Size{Width: 0, Height: 0}
	}

	maxWidth := 0
	maxHeight := 0

	// Find maximum dimensions
	for _, child := range z.items {
		size := child.Layout(constraints)
		if size.Width > maxWidth {
			maxWidth = size.Width
		}
		if size.Height > maxHeight {
			maxHeight = size.Height
		}
	}

	result := Size{Width: maxWidth, Height: maxHeight}

	// Apply width constraint if specified
	if !z.width.IsAuto() {
		result.Width = z.width.Apply(constraints.MaxWidth, result.Width)
	}

	// Apply height constraint if specified
	if !z.height.IsAuto() {
		result.Height = z.height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (z *ZStack) Children() []Element {
	return z.items
}
