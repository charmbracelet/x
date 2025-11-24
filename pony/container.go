package pony

import uv "github.com/charmbracelet/ultraviolet"

// VStack represents a vertical stack container.
type VStack struct {
	BaseElement
	Items  []Element
	Gap    int
	Width  SizeConstraint
	Height SizeConstraint
	Align  string // left, center, right (horizontal alignment of children)
}

var _ Element = (*VStack)(nil)

// NewVStack creates a new vertical stack.
func NewVStack(children ...Element) *VStack {
	return &VStack{Items: children}
}

// WithGap sets the gap and returns the vstack for chaining.
func (v *VStack) WithGap(gap int) *VStack {
	v.Gap = gap
	return v
}

// WithAlign sets the alignment and returns the vstack for chaining.
func (v *VStack) WithAlign(align string) *VStack {
	v.Align = align
	return v
}

// WithWidth sets the width constraint and returns the vstack for chaining.
func (v *VStack) WithWidth(width SizeConstraint) *VStack {
	v.Width = width
	return v
}

// WithHeight sets the height constraint and returns the vstack for chaining.
func (v *VStack) WithHeight(height SizeConstraint) *VStack {
	v.Height = height
	return v
}

// calculateChildSizes performs two-pass layout for VStack children.
// Pass 1: Layout fixed children, Pass 2: Distribute space to flexible children (flex-grow).
func (v *VStack) calculateChildSizes(constraints Constraints) []Size {
	childSizes := make([]Size, len(v.Items))
	if len(v.Items) == 0 {
		return childSizes
	}

	// Pass 1: Layout fixed children and count flexible items
	fixedHeight := 0
	totalFlexGrow := 0

	for i, child := range v.Items {
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

		if i < len(v.Items)-1 {
			fixedHeight += v.Gap
		}
	}

	// Pass 2: Distribute remaining space among flexible items based on flex-grow
	if totalFlexGrow > 0 {
		remainingHeight := constraints.MaxHeight - fixedHeight
		if remainingHeight > 0 {
			for i, child := range v.Items {
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

	if len(v.Items) == 0 {
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

	for i, child := range v.Items {
		if y >= area.Max.Y {
			break
		}

		childSize := childSizes[i]

		// Calculate x position based on horizontal alignment
		var x int
		switch v.Align {
		case AlignCenter:
			if childSize.Width < area.Dx() {
				x = area.Min.X + (area.Dx()-childSize.Width)/2
			} else {
				x = area.Min.X
			}
		case AlignRight:
			if childSize.Width < area.Dx() {
				x = area.Max.X - childSize.Width
			} else {
				x = area.Min.X
			}
		default: // AlignLeft
			x = area.Min.X
		}

		childArea := uv.Rect(x, y, childSize.Width, childSize.Height)
		// Clip to parent bounds
		childArea = childArea.Intersect(area)

		child.Draw(scr, childArea)

		y += childSize.Height
		if i < len(v.Items)-1 {
			y += v.Gap
		}
	}
}

// Layout calculates the total size of the vertical stack.
func (v *VStack) Layout(constraints Constraints) Size {
	if len(v.Items) == 0 {
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

		if i < len(v.Items)-1 {
			totalHeight += v.Gap
		}
	}

	result := Size{Width: maxWidth, Height: totalHeight}

	if !v.Width.IsAuto() {
		result.Width = v.Width.Apply(constraints.MaxWidth, result.Width)
	}

	if !v.Height.IsAuto() {
		result.Height = v.Height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (v *VStack) Children() []Element {
	return v.Items
}

// HStack represents a horizontal stack container.
type HStack struct {
	BaseElement
	Items  []Element
	Gap    int
	Width  SizeConstraint
	Height SizeConstraint
	Valign string // top, middle, bottom (vertical alignment of children)
}

var _ Element = (*HStack)(nil)

// NewHStack creates a new horizontal stack.
func NewHStack(children ...Element) *HStack {
	return &HStack{Items: children}
}

// WithGap sets the gap and returns the hstack for chaining.
func (h *HStack) WithGap(gap int) *HStack {
	h.Gap = gap
	return h
}

// WithValign sets the vertical alignment and returns the hstack for chaining.
func (h *HStack) WithValign(valign string) *HStack {
	h.Valign = valign
	return h
}

// WithWidth sets the width constraint and returns the hstack for chaining.
func (h *HStack) WithWidth(width SizeConstraint) *HStack {
	h.Width = width
	return h
}

// WithHeight sets the height constraint and returns the hstack for chaining.
func (h *HStack) WithHeight(height SizeConstraint) *HStack {
	h.Height = height
	return h
}

// calculateChildSizes performs two-pass layout for HStack children.
// Pass 1: Layout fixed children, Pass 2: Distribute space to flexible children (flex-grow).
func (h *HStack) calculateChildSizes(constraints Constraints) []Size {
	childSizes := make([]Size, len(h.Items))
	if len(h.Items) == 0 {
		return childSizes
	}

	// Pass 1: Layout fixed children and count flexible items
	fixedWidth := 0
	totalFlexGrow := 0

	for i, child := range h.Items {
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

		if i < len(h.Items)-1 {
			fixedWidth += h.Gap
		}
	}

	// Pass 2: Distribute remaining space among flexible items based on flex-grow
	if totalFlexGrow > 0 {
		remainingWidth := constraints.MaxWidth - fixedWidth
		if remainingWidth > 0 {
			for i, child := range h.Items {
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

	if len(h.Items) == 0 {
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

	for i, child := range h.Items {
		if x >= area.Max.X {
			break
		}

		childSize := childSizes[i]

		// Calculate y position based on vertical alignment
		var y int
		switch h.Valign {
		case AlignMiddle:
			if childSize.Height < area.Dy() {
				y = area.Min.Y + (area.Dy()-childSize.Height)/2
			} else {
				y = area.Min.Y
			}
		case AlignBottom:
			if childSize.Height < area.Dy() {
				y = area.Max.Y - childSize.Height
			} else {
				y = area.Min.Y
			}
		default: // AlignTop
			y = area.Min.Y
		}

		childArea := uv.Rect(x, y, childSize.Width, childSize.Height)
		// Clip to parent bounds
		childArea = childArea.Intersect(area)

		child.Draw(scr, childArea)

		x += childSize.Width
		if i < len(h.Items)-1 {
			x += h.Gap
		}
	}
}

// Layout calculates the total size of the horizontal stack.
func (h *HStack) Layout(constraints Constraints) Size {
	if len(h.Items) == 0 {
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

		if i < len(h.Items)-1 {
			totalWidth += h.Gap
		}
	}

	result := Size{Width: totalWidth, Height: maxHeight}

	if !h.Width.IsAuto() {
		result.Width = h.Width.Apply(constraints.MaxWidth, result.Width)
	}

	if !h.Height.IsAuto() {
		result.Height = h.Height.Apply(constraints.MaxHeight, result.Height)
	}

	return constraints.Constrain(result)
}

// Children returns the child elements.
func (h *HStack) Children() []Element {
	return h.Items
}
