package pony

import uv "github.com/charmbracelet/ultraviolet"

// ZStack represents a layered stack container where children are drawn on top of each other.
// Later children in the stack are drawn on top of earlier children.
type ZStack struct {
	BaseElement
	Items  []Element
	Width  SizeConstraint
	Height SizeConstraint
	Align  string // Horizontal alignment: left, center, right
	Valign string // Vertical alignment: top, middle, bottom
}

var _ Element = (*ZStack)(nil)

// NewZStack creates a new layered stack.
func NewZStack(children ...Element) *ZStack {
	return &ZStack{
		Items:  children,
		Align:  AlignCenter,
		Valign: AlignMiddle,
	}
}

// WithAlign sets the horizontal alignment and returns the zstack for chaining.
func (z *ZStack) WithAlign(align string) *ZStack {
	z.Align = align
	return z
}

// WithValign sets the vertical alignment and returns the zstack for chaining.
func (z *ZStack) WithValign(valign string) *ZStack {
	z.Valign = valign
	return z
}

// WithWidth sets the width constraint and returns the zstack for chaining.
func (z *ZStack) WithWidth(width SizeConstraint) *ZStack {
	z.Width = width
	return z
}

// WithHeight sets the height constraint and returns the zstack for chaining.
func (z *ZStack) WithHeight(height SizeConstraint) *ZStack {
	z.Height = height
	return z
}

// Draw renders the layered stack to the screen.
func (z *ZStack) Draw(scr uv.Screen, area uv.Rectangle) {
	z.SetBounds(area)

	if len(z.Items) == 0 {
		return
	}

	// Layout all children first to get their sizes
	childConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  area.Dx(),
		MinHeight: 0,
		MaxHeight: area.Dy(),
	}

	childSizes := make([]Size, len(z.Items))
	for i, child := range z.Items {
		childSizes[i] = child.Layout(childConstraints)
	}

	// Draw each child in order (later children draw on top)
	for i, child := range z.Items {
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
		case z.Align == AlignLeft && z.Valign == AlignTop:
			childArea = uv.TopLeftRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignCenter && z.Valign == AlignTop:
			childArea = uv.TopCenterRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignRight && z.Valign == AlignTop:
			childArea = uv.TopRightRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignLeft && z.Valign == AlignMiddle:
			childArea = uv.LeftCenterRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignCenter && z.Valign == AlignMiddle:
			childArea = uv.CenterRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignRight && z.Valign == AlignMiddle:
			childArea = uv.RightCenterRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignLeft && z.Valign == AlignBottom:
			childArea = uv.BottomLeftRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignCenter && z.Valign == AlignBottom:
			childArea = uv.BottomCenterRect(area, childSize.Width, childSize.Height)
		case z.Align == AlignRight && z.Valign == AlignBottom:
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
	if len(z.Items) == 0 {
		return Size{Width: 0, Height: 0}
	}

	maxWidth := 0
	maxHeight := 0

	// Find maximum dimensions
	for _, child := range z.Items {
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
	if !z.Width.IsAuto() {
		result.Width = z.Width.Apply(constraints.MaxWidth, result.Width)
	}

	// Apply height constraint if specified
	if !z.Height.IsAuto() {
		result.Height = z.Height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (z *ZStack) Children() []Element {
	return z.Items
}
