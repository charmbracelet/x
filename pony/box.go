package pony

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Box represents a container with an optional border.
type Box struct {
	BaseElement
	Child        Element
	Border       string // normal, rounded, thick, double, hidden, none
	BorderStyle  uv.Style
	Width        SizeConstraint
	Height       SizeConstraint
	Padding      int
	Margin       int // margin on all sides
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int
}

var _ Element = (*Box)(nil)

// NewBox creates a new box element.
func NewBox(child Element) *Box {
	return &Box{
		Child:  child,
		Border: BorderNone,
	}
}

// WithBorder sets the border style and returns the box for chaining.
func (b *Box) WithBorder(border string) *Box {
	b.Border = border
	return b
}

// WithBorderStyle sets the border style and returns the box for chaining.
func (b *Box) WithBorderStyle(style uv.Style) *Box {
	b.BorderStyle = style
	return b
}

// WithPadding sets the padding and returns the box for chaining.
func (b *Box) WithPadding(padding int) *Box {
	b.Padding = padding
	return b
}

// WithMargin sets the margin on all sides and returns the box for chaining.
func (b *Box) WithMargin(margin int) *Box {
	b.Margin = margin
	return b
}

// WithMarginTop sets the top margin and returns the box for chaining.
func (b *Box) WithMarginTop(margin int) *Box {
	b.MarginTop = margin
	return b
}

// WithMarginRight sets the right margin and returns the box for chaining.
func (b *Box) WithMarginRight(margin int) *Box {
	b.MarginRight = margin
	return b
}

// WithMarginBottom sets the bottom margin and returns the box for chaining.
func (b *Box) WithMarginBottom(margin int) *Box {
	b.MarginBottom = margin
	return b
}

// WithMarginLeft sets the left margin and returns the box for chaining.
func (b *Box) WithMarginLeft(margin int) *Box {
	b.MarginLeft = margin
	return b
}

// WithWidth sets the width constraint and returns the box for chaining.
func (b *Box) WithWidth(width SizeConstraint) *Box {
	b.Width = width
	return b
}

// WithHeight sets the height constraint and returns the box for chaining.
func (b *Box) WithHeight(height SizeConstraint) *Box {
	b.Height = height
	return b
}

// Draw renders the box to the screen.
func (b *Box) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	// Apply margin (shrink the area before drawing)
	marginTop := b.MarginTop
	if marginTop == 0 {
		marginTop = b.Margin
	}
	marginRight := b.MarginRight
	if marginRight == 0 {
		marginRight = b.Margin
	}
	marginBottom := b.MarginBottom
	if marginBottom == 0 {
		marginBottom = b.Margin
	}
	marginLeft := b.MarginLeft
	if marginLeft == 0 {
		marginLeft = b.Margin
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
	if b.Border != "" && b.Border != BorderNone {
		var uvBorder uv.Border
		switch b.Border {
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

		// Apply border style if specified
		if !b.BorderStyle.IsZero() {
			uvBorder = uvBorder.Style(b.BorderStyle)
		}

		uvBorder.Draw(scr, area)

		// Shrink area for child content (leave space for border)
		if area.Dx() > 2 && area.Dy() > 2 {
			area = uv.Rect(area.Min.X+1, area.Min.Y+1, area.Dx()-2, area.Dy()-2)
		}
	}

	// Apply padding
	if b.Padding > 0 {
		padH := b.Padding * 2 // left + right
		padV := b.Padding * 2 // top + bottom
		if area.Dx() > padH && area.Dy() > padV {
			area = uv.Rect(
				area.Min.X+b.Padding,
				area.Min.Y+b.Padding,
				area.Dx()-padH,
				area.Dy()-padV,
			)
		}
	}

	// Draw child if present
	if b.Child != nil {
		b.Child.Draw(scr, area)
	}
}

// Layout calculates the box size.
func (b *Box) Layout(constraints Constraints) Size {
	// Account for margin
	marginTop := b.MarginTop
	if marginTop == 0 {
		marginTop = b.Margin
	}
	marginRight := b.MarginRight
	if marginRight == 0 {
		marginRight = b.Margin
	}
	marginBottom := b.MarginBottom
	if marginBottom == 0 {
		marginBottom = b.Margin
	}
	marginLeft := b.MarginLeft
	if marginLeft == 0 {
		marginLeft = b.Margin
	}

	marginWidth := marginLeft + marginRight
	marginHeight := marginTop + marginBottom

	// Account for border
	borderWidth := 0
	borderHeight := 0
	if b.Border != "" && b.Border != BorderNone {
		borderWidth = 2
		borderHeight = 2
	}

	// Account for padding
	paddingWidth := b.Padding * 2
	paddingHeight := b.Padding * 2

	totalReduction := marginWidth + borderWidth + paddingWidth
	totalReductionH := marginHeight + borderHeight + paddingHeight

	childConstraints := Constraints{
		MinWidth:  max(0, constraints.MinWidth-totalReduction),
		MaxWidth:  max(0, constraints.MaxWidth-totalReduction),
		MinHeight: max(0, constraints.MinHeight-totalReductionH),
		MaxHeight: max(0, constraints.MaxHeight-totalReductionH),
	}

	var childSize Size
	if b.Child != nil {
		childSize = b.Child.Layout(childConstraints)
	}

	totalSize := Size{
		Width:  childSize.Width + marginWidth + borderWidth + paddingWidth,
		Height: childSize.Height + marginHeight + borderHeight + paddingHeight,
	}

	// Apply width constraint if specified
	if !b.Width.IsAuto() {
		totalSize.Width = b.Width.Apply(constraints.MaxWidth, totalSize.Width)
	}

	// Apply height constraint if specified
	if !b.Height.IsAuto() {
		totalSize.Height = b.Height.Apply(constraints.MaxHeight, totalSize.Height)
	}

	return constraints.Constrain(totalSize)
}

// Children returns the child element.
func (b *Box) Children() []Element {
	if b.Child == nil {
		return nil
	}
	return []Element{b.Child}
}
