package pony

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// Box represents a container with an optional border.
type Box struct {
	BaseElement
	child        Element
	border       string // normal, rounded, thick, double, hidden, none
	borderColor  color.Color
	width        SizeConstraint
	height       SizeConstraint
	padding      int
	margin       int // margin on all sides
	marginTop    int
	marginRight  int
	marginBottom int
	marginLeft   int
}

var _ Element = (*Box)(nil)

// NewBox creates a new box element.
func NewBox(child Element) *Box {
	return &Box{
		child:  child,
		border: BorderNone,
	}
}

// Border sets the border style and returns the box for chaining.
func (b *Box) Border(border string) *Box {
	b.border = border
	return b
}

// BorderColor sets the border color and returns the box for chaining.
func (b *Box) BorderColor(c color.Color) *Box {
	b.borderColor = c
	return b
}

// Padding sets the padding and returns the box for chaining.
func (b *Box) Padding(padding int) *Box {
	b.padding = padding
	return b
}

// Margin sets the margin on all sides and returns the box for chaining.
func (b *Box) Margin(margin int) *Box {
	b.margin = margin
	return b
}

// MarginTop sets the top margin and returns the box for chaining.
func (b *Box) MarginTop(margin int) *Box {
	b.marginTop = margin
	return b
}

// MarginRight sets the right margin and returns the box for chaining.
func (b *Box) MarginRight(margin int) *Box {
	b.marginRight = margin
	return b
}

// MarginBottom sets the bottom margin and returns the box for chaining.
func (b *Box) MarginBottom(margin int) *Box {
	b.marginBottom = margin
	return b
}

// MarginLeft sets the left margin and returns the box for chaining.
func (b *Box) MarginLeft(margin int) *Box {
	b.marginLeft = margin
	return b
}

// Width sets the width constraint and returns the box for chaining.
func (b *Box) Width(width SizeConstraint) *Box {
	b.width = width
	return b
}

// Height sets the height constraint and returns the box for chaining.
func (b *Box) Height(height SizeConstraint) *Box {
	b.height = height
	return b
}

// Draw renders the box to the screen.
func (b *Box) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	// Apply margin (shrink the area before drawing)
	marginTop := b.marginTop
	if marginTop == 0 {
		marginTop = b.margin
	}
	marginRight := b.marginRight
	if marginRight == 0 {
		marginRight = b.margin
	}
	marginBottom := b.marginBottom
	if marginBottom == 0 {
		marginBottom = b.margin
	}
	marginLeft := b.marginLeft
	if marginLeft == 0 {
		marginLeft = b.margin
	}

	marginH := marginLeft + marginRight
	marginV := marginTop + marginBottom

	if area.Dx() > marginH && area.Dy() > marginV {
		area = uv.Rect(
			area.Min.X+marginLeft,
			area.Min.Y+marginTop,
			area.Dx()-marginH,
			area.Dy()-marginV,
		)
	}

	// Draw border if specified
	if b.border != "" && b.border != BorderNone {
		var uvBorder uv.Border
		switch b.border {
		case BorderNormal:
			uvBorder = uv.NormalBorder()
		case BorderRounded:
			uvBorder = uv.RoundedBorder()
		case BorderThick:
			uvBorder = uv.ThickBorder()
		case BorderDouble:
			uvBorder = uv.DoubleBorder()
		case BorderHidden:
			uvBorder = uv.HiddenBorder()
		default:
			uvBorder = uv.NormalBorder()
		}

		// Apply border color if specified
		if b.borderColor != nil {
			uvBorder = uvBorder.Style(uv.Style{Fg: b.borderColor})
		}

		uvBorder.Draw(scr, area)

		// Shrink area for child content (leave space for border)
		if area.Dx() > 2 && area.Dy() > 2 {
			area = uv.Rect(area.Min.X+1, area.Min.Y+1, area.Dx()-2, area.Dy()-2)
		}
	}

	// Apply padding
	if b.padding > 0 {
		padH := b.padding * 2 // left + right
		padV := b.padding * 2 // top + bottom
		if area.Dx() > padH && area.Dy() > padV {
			area = uv.Rect(
				area.Min.X+b.padding,
				area.Min.Y+b.padding,
				area.Dx()-padH,
				area.Dy()-padV,
			)
		}
	}

	// Draw child if present
	if b.child != nil {
		b.child.Draw(scr, area)
	}
}

// Layout calculates the box size.
func (b *Box) Layout(constraints Constraints) Size {
	// Account for margin
	marginTop := b.marginTop
	if marginTop == 0 {
		marginTop = b.margin
	}
	marginRight := b.marginRight
	if marginRight == 0 {
		marginRight = b.margin
	}
	marginBottom := b.marginBottom
	if marginBottom == 0 {
		marginBottom = b.margin
	}
	marginLeft := b.marginLeft
	if marginLeft == 0 {
		marginLeft = b.margin
	}

	marginWidth := marginLeft + marginRight
	marginHeight := marginTop + marginBottom

	// Account for border
	borderWidth := 0
	borderHeight := 0
	if b.border != "" && b.border != BorderNone {
		borderWidth = 2
		borderHeight = 2
	}

	// Account for padding
	paddingWidth := b.padding * 2
	paddingHeight := b.padding * 2

	totalReduction := marginWidth + borderWidth + paddingWidth
	totalReductionH := marginHeight + borderHeight + paddingHeight

	childConstraints := Constraints{
		MinWidth:  max(0, constraints.MinWidth-totalReduction),
		MaxWidth:  max(0, constraints.MaxWidth-totalReduction),
		MinHeight: max(0, constraints.MinHeight-totalReductionH),
		MaxHeight: max(0, constraints.MaxHeight-totalReductionH),
	}

	var childSize Size
	if b.child != nil {
		childSize = b.child.Layout(childConstraints)
	}

	totalSize := Size{
		Width:  childSize.Width + marginWidth + borderWidth + paddingWidth,
		Height: childSize.Height + marginHeight + borderHeight + paddingHeight,
	}

	// Apply width constraint if specified
	if !b.width.IsAuto() {
		totalSize.Width = b.width.Apply(constraints.MaxWidth, totalSize.Width)
	}

	// Apply height constraint if specified
	if !b.height.IsAuto() {
		totalSize.Height = b.height.Apply(constraints.MaxHeight, totalSize.Height)
	}

	return constraints.Constrain(totalSize)
}

// Children returns the child element.
func (b *Box) Children() []Element {
	if b.child == nil {
		return nil
	}
	return []Element{b.child}
}
