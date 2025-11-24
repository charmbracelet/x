package pony

import uv "github.com/charmbracelet/ultraviolet"

// VStack represents a vertical stack container.
type VStack struct {
	BaseElement
	items     []Element
	spacing   int
	width     SizeConstraint
	height    SizeConstraint
	alignment string // leading, center, trailing (horizontal alignment of children)
}

var _ Element = (*VStack)(nil)

// NewVStack creates a new vertical stack.
func NewVStack(children ...Element) *VStack {
	return &VStack{items: children}
}

// Spacing sets the spacing between children and returns the vstack for chaining.
func (v *VStack) Spacing(spacing int) *VStack {
	v.spacing = spacing
	return v
}

// Alignment sets the horizontal alignment of children and returns the vstack for chaining.
func (v *VStack) Alignment(alignment string) *VStack {
	v.alignment = alignment
	return v
}

// Width sets the width constraint and returns the vstack for chaining.
func (v *VStack) Width(width SizeConstraint) *VStack {
	v.width = width
	return v
}

// Height sets the height constraint and returns the vstack for chaining.
func (v *VStack) Height(height SizeConstraint) *VStack {
	v.height = height
	return v
}

// calculateChildSizes performs two-pass layout for VStack children.
// Pass 1: Layout fixed children, Pass 2: Distribute space to flexible children (flex-grow).
func (v *VStack) calculateChildSizes(constraints Constraints) []Size {
	childSizes := make([]Size, len(v.items))
	if len(v.items) == 0 {
		return childSizes
	}

	// Pass 1: Layout fixed children and count flexible items
	fixedHeight := 0
	totalFlexGrow := 0

	for i, child := range v.items {
		flexGrow := GetFlexGrow(child)

		if flexGrow > 0 {
			// Flexible item - will be sized in pass 2
			totalFlexGrow += flexGrow
			childSizes[i] = Size{Width: 0, Height: 0}
		} else {
			// Fixed item - layout now
			childConstraints := Constraints{
				MinWidth:  0,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: 0,
				MaxHeight: constraints.MaxHeight - fixedHeight,
			}

			size := child.Layout(childConstraints)
			childSizes[i] = size
			fixedHeight += size.Height
		}

		if i < len(v.items)-1 {
			fixedHeight += v.spacing
		}
	}

	// Pass 2: Distribute remaining space among flexible items based on flex-grow
	if totalFlexGrow > 0 {
		remainingHeight := constraints.MaxHeight - fixedHeight
		if remainingHeight > 0 {
			for i, child := range v.items {
				flexGrow := GetFlexGrow(child)
				if flexGrow > 0 {
					// Allocate space proportional to flex-grow
					flexHeight := (remainingHeight * flexGrow) / totalFlexGrow

					// Layout the flexible child with its allocated space
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  constraints.MaxWidth,
						MinHeight: flexHeight,
						MaxHeight: flexHeight,
					}

					childSizes[i] = child.Layout(childConstraints)
				}
			}
		}
	}

	return childSizes
}

// Draw renders the vertical stack to the screen.
func (v *VStack) Draw(scr uv.Screen, area uv.Rectangle) {
	v.SetBounds(area)

	if len(v.items) == 0 {
		return
	}

	// Calculate child sizes using two-pass layout
	childSizes := v.calculateChildSizes(Constraints{
		MinWidth:  0,
		MaxWidth:  area.Dx(),
		MinHeight: 0,
		MaxHeight: area.Dy(),
	})

	// Draw all children with calculated sizes
	y := area.Min.Y

	for i, child := range v.items {
		if y >= area.Max.Y {
			break
		}

		childSize := childSizes[i]

		// Calculate x position based on horizontal alignment
		var x int
		switch v.alignment {
		case AlignmentCenter:
			if childSize.Width < area.Dx() {
				x = area.Min.X + (area.Dx()-childSize.Width)/2
			} else {
				x = area.Min.X
			}
		case AlignmentTrailing:
			if childSize.Width < area.Dx() {
				x = area.Max.X - childSize.Width
			} else {
				x = area.Min.X
			}
		default: // AlignmentLeading
			x = area.Min.X
		}

		childArea := uv.Rect(x, y, childSize.Width, childSize.Height)
		// Clip to parent bounds
		childArea = childArea.Intersect(area)

		child.Draw(scr, childArea)

		y += childSize.Height
		if i < len(v.items)-1 {
			y += v.spacing
		}
	}
}

// Layout calculates the total size of the vertical stack.
func (v *VStack) Layout(constraints Constraints) Size {
	if len(v.items) == 0 {
		return Size{Width: 0, Height: 0}
	}

	// Calculate child sizes using two-pass layout
	childSizes := v.calculateChildSizes(constraints)

	// Sum up total size
	totalHeight := 0
	maxWidth := 0

	for i, size := range childSizes {
		totalHeight += size.Height
		if size.Width > maxWidth {
			maxWidth = size.Width
		}

		if i < len(v.items)-1 {
			totalHeight += v.spacing
		}
	}

	result := Size{Width: maxWidth, Height: totalHeight}

	if !v.width.IsAuto() {
		result.Width = v.width.Apply(constraints.MaxWidth, result.Width)
	}

	if !v.height.IsAuto() {
		result.Height = v.height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (v *VStack) Children() []Element {
	return v.items
}

// HStack represents a horizontal stack container.
type HStack struct {
	BaseElement
	items     []Element
	spacing   int
	width     SizeConstraint
	height    SizeConstraint
	alignment string // top, center, bottom (vertical alignment of children)
}

var _ Element = (*HStack)(nil)

// NewHStack creates a new horizontal stack.
func NewHStack(children ...Element) *HStack {
	return &HStack{items: children}
}

// Spacing sets the spacing between children and returns the hstack for chaining.
func (h *HStack) Spacing(spacing int) *HStack {
	h.spacing = spacing
	return h
}

// Alignment sets the vertical alignment of children and returns the hstack for chaining.
func (h *HStack) Alignment(alignment string) *HStack {
	h.alignment = alignment
	return h
}

// Width sets the width constraint and returns the hstack for chaining.
func (h *HStack) Width(width SizeConstraint) *HStack {
	h.width = width
	return h
}

// Height sets the height constraint and returns the hstack for chaining.
func (h *HStack) Height(height SizeConstraint) *HStack {
	h.height = height
	return h
}

// calculateChildSizes performs two-pass layout for HStack children.
// Pass 1: Layout fixed children, Pass 2: Distribute space to flexible children (flex-grow).
func (h *HStack) calculateChildSizes(constraints Constraints) []Size {
	childSizes := make([]Size, len(h.items))
	if len(h.items) == 0 {
		return childSizes
	}

	// Pass 1: Layout fixed children and count flexible items
	fixedWidth := 0
	totalFlexGrow := 0

	for i, child := range h.items {
		flexGrow := GetFlexGrow(child)

		if flexGrow > 0 {
			// Flexible item - will be sized in pass 2
			totalFlexGrow += flexGrow
			childSizes[i] = Size{Width: 0, Height: 0}
		} else {
			// Fixed item - layout now
			childConstraints := Constraints{
				MinWidth:  0,
				MaxWidth:  constraints.MaxWidth - fixedWidth,
				MinHeight: 0,
				MaxHeight: constraints.MaxHeight,
			}

			size := child.Layout(childConstraints)
			childSizes[i] = size
			fixedWidth += size.Width
		}

		if i < len(h.items)-1 {
			fixedWidth += h.spacing
		}
	}

	// Pass 2: Distribute remaining space among flexible items based on flex-grow
	if totalFlexGrow > 0 {
		remainingWidth := constraints.MaxWidth - fixedWidth
		if remainingWidth > 0 {
			for i, child := range h.items {
				flexGrow := GetFlexGrow(child)
				if flexGrow > 0 {
					// Allocate space proportional to flex-grow
					flexWidth := (remainingWidth * flexGrow) / totalFlexGrow

					// Layout the flexible child with its allocated space
					childConstraints := Constraints{
						MinWidth:  flexWidth,
						MaxWidth:  flexWidth,
						MinHeight: 0,
						MaxHeight: constraints.MaxHeight,
					}

					childSizes[i] = child.Layout(childConstraints)
				}
			}
		}
	}

	return childSizes
}

// Draw renders the horizontal stack to the screen.
func (h *HStack) Draw(scr uv.Screen, area uv.Rectangle) {
	h.SetBounds(area)

	if len(h.items) == 0 {
		return
	}

	// Calculate child sizes using two-pass layout
	childSizes := h.calculateChildSizes(Constraints{
		MinWidth:  0,
		MaxWidth:  area.Dx(),
		MinHeight: 0,
		MaxHeight: area.Dy(),
	})

	// Draw all children
	x := area.Min.X

	for i, child := range h.items {
		if x >= area.Max.X {
			break
		}

		childSize := childSizes[i]

		// Calculate y position based on vertical alignment
		var y int
		switch h.alignment {
		case AlignmentCenter:
			if childSize.Height < area.Dy() {
				y = area.Min.Y + (area.Dy()-childSize.Height)/2
			} else {
				y = area.Min.Y
			}
		case AlignmentBottom:
			if childSize.Height < area.Dy() {
				y = area.Max.Y - childSize.Height
			} else {
				y = area.Min.Y
			}
		default: // AlignmentTop
			y = area.Min.Y
		}

		childArea := uv.Rect(x, y, childSize.Width, childSize.Height)
		// Clip to parent bounds
		childArea = childArea.Intersect(area)

		child.Draw(scr, childArea)

		x += childSize.Width
		if i < len(h.items)-1 {
			x += h.spacing
		}
	}
}

// Layout calculates the total size of the horizontal stack.
func (h *HStack) Layout(constraints Constraints) Size {
	if len(h.items) == 0 {
		return Size{Width: 0, Height: 0}
	}

	// Calculate child sizes using two-pass layout
	childSizes := h.calculateChildSizes(constraints)

	// Sum up total size
	totalWidth := 0
	maxHeight := 0

	for i, size := range childSizes {
		totalWidth += size.Width
		if size.Height > maxHeight {
			maxHeight = size.Height
		}

		if i < len(h.items)-1 {
			totalWidth += h.spacing
		}
	}

	result := Size{Width: totalWidth, Height: maxHeight}

	if !h.width.IsAuto() {
		result.Width = h.width.Apply(constraints.MaxWidth, result.Width)
	}

	if !h.height.IsAuto() {
		result.Height = h.height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (h *HStack) Children() []Element {
	return h.items
}
