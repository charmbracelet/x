package pony

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// ScrollView represents a scrollable container.
// Content can be larger than the viewport and will be clipped.
type ScrollView struct {
	BaseElement
	child Element

	// Scroll position
	offsetX int
	offsetY int

	// Viewport size constraints
	width  SizeConstraint
	height SizeConstraint

	// Scrollbar options
	showScrollbar  bool
	scrollbarColor color.Color

	// Scroll direction
	horizontal bool // If true, scrolls horizontally
	vertical   bool // If true, scrolls vertically (default)
}

var _ Element = (*ScrollView)(nil)

// NewScrollView creates a new scrollable view.
func NewScrollView(child Element) *ScrollView {
	return &ScrollView{
		child:         child,
		vertical:      true, // Default to vertical scrolling
		showScrollbar: true,
	}
}

// Offset sets the scroll offset and returns the scroll view for chaining.
func (s *ScrollView) Offset(x, y int) *ScrollView {
	s.offsetX = x
	s.offsetY = y
	return s
}

// Vertical enables/disables vertical scrolling.
func (s *ScrollView) Vertical(enabled bool) *ScrollView {
	s.vertical = enabled
	return s
}

// Horizontal enables/disables horizontal scrolling.
func (s *ScrollView) Horizontal(enabled bool) *ScrollView {
	s.horizontal = enabled
	return s
}

// Scrollbar enables/disables scrollbar.
func (s *ScrollView) Scrollbar(show bool) *ScrollView {
	s.showScrollbar = show
	return s
}

// ScrollbarColor sets the scrollbar color.
func (s *ScrollView) ScrollbarColor(c color.Color) *ScrollView {
	s.scrollbarColor = c
	return s
}

// Width sets the width constraint.
func (s *ScrollView) Width(width SizeConstraint) *ScrollView {
	s.width = width
	return s
}

// Height sets the height constraint.
func (s *ScrollView) Height(height SizeConstraint) *ScrollView {
	s.height = height
	return s
}

// ScrollUp scrolls up by the given amount.
func (s *ScrollView) ScrollUp(amount int) {
	s.offsetY = max(0, s.offsetY-amount)
}

// ScrollDown scrolls down by the given amount.
func (s *ScrollView) ScrollDown(amount int, contentHeight, viewportHeight int) {
	maxOffset := max(0, contentHeight-viewportHeight)
	s.offsetY = min(maxOffset, s.offsetY+amount)
}

// ScrollLeft scrolls left by the given amount.
func (s *ScrollView) ScrollLeft(amount int) {
	s.offsetX = max(0, s.offsetX-amount)
}

// ScrollRight scrolls right by the given amount.
func (s *ScrollView) ScrollRight(amount int, contentWidth, viewportWidth int) {
	maxOffset := max(0, contentWidth-viewportWidth)
	s.offsetX = min(maxOffset, s.offsetX+amount)
}

// Draw renders the scrollable view.
func (s *ScrollView) Draw(scr uv.Screen, area uv.Rectangle) {
	s.SetBounds(area)

	if s.child == nil {
		return
	}

	// Calculate viewport size
	viewportWidth := area.Dx()
	viewportHeight := area.Dy()

	// Reserve space for scrollbar if shown
	scrollbarWidth := 0
	scrollbarHeight := 0
	if s.showScrollbar {
		if s.vertical && !s.horizontal {
			scrollbarWidth = 1
			viewportWidth -= scrollbarWidth
		}
		if s.horizontal && !s.vertical {
			scrollbarHeight = 1
			viewportHeight -= scrollbarHeight
		}
		if s.horizontal && s.vertical {
			scrollbarWidth = 1
			scrollbarHeight = 1
			viewportWidth -= scrollbarWidth
			viewportHeight -= scrollbarHeight
		}
	}

	// Layout child with unbounded constraints to get full content size
	contentConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  1 << 30, // Very large number
		MinHeight: 0,
		MaxHeight: 1 << 30,
	}
	contentSize := s.child.Layout(contentConstraints)

	// Create a buffer for the full content
	contentBuffer := uv.NewScreenBuffer(contentSize.Width, contentSize.Height)
	contentArea := uv.Rect(0, 0, contentSize.Width, contentSize.Height)
	s.child.Draw(contentBuffer, contentArea)

	// Adjust child bounds to screen coordinates (accounting for viewport position and scroll offset)
	s.adjustChildBounds(s.child, area.Min.X-s.offsetX, area.Min.Y-s.offsetY)

	// Copy visible portion to screen (with offset)
	for y := 0; y < viewportHeight; y++ {
		for x := 0; x < viewportWidth; x++ {
			// Source position in content buffer (with offset)
			srcX := x + s.offsetX
			srcY := y + s.offsetY

			// Destination position on screen
			dstX := area.Min.X + x
			dstY := area.Min.Y + y

			// Copy cell if in bounds
			if srcY < contentSize.Height && srcX < contentSize.Width {
				cell := contentBuffer.CellAt(srcX, srcY)
				scr.SetCell(dstX, dstY, cell)
			}
		}
	}

	// Draw scrollbar if enabled
	if s.showScrollbar {
		if s.vertical {
			s.drawVerticalScrollbar(scr, area, contentSize.Height, viewportHeight, scrollbarWidth)
		}
		if s.horizontal {
			s.drawHorizontalScrollbar(scr, area, contentSize.Width, viewportWidth, scrollbarHeight)
		}
	}
}

// drawVerticalScrollbar draws a vertical scrollbar.
func (s *ScrollView) drawVerticalScrollbar(scr uv.Screen, area uv.Rectangle, contentHeight, viewportHeight, scrollbarWidth int) {
	if contentHeight <= viewportHeight {
		return // No need for scrollbar
	}

	scrollbarX := area.Max.X - scrollbarWidth
	scrollbarStart := area.Min.Y
	scrollbarEnd := area.Max.Y
	trackHeight := scrollbarEnd - scrollbarStart

	// Calculate scrollbar thumb size
	thumbHeight := max(1, (viewportHeight*trackHeight)/contentHeight)

	// Calculate scrollbar thumb position
	// scrollableRange is how far we can scroll
	scrollableRange := contentHeight - viewportHeight
	// trackRange is how far the thumb can move
	trackRange := trackHeight - thumbHeight

	// Position the thumb proportionally
	thumbPos := scrollbarStart
	if scrollableRange > 0 {
		thumbPos = scrollbarStart + (s.offsetY*trackRange)/scrollableRange
	}

	// Ensure thumb stays within bounds (handle rounding edge cases)
	if thumbPos+thumbHeight > scrollbarEnd {
		thumbPos = scrollbarEnd - thumbHeight
	}
	if thumbPos < scrollbarStart {
		thumbPos = scrollbarStart
	}

	// Create scrollbar cells
	trackCell := uv.NewCell(scr.WidthMethod(), "░")
	thumbCell := uv.NewCell(scr.WidthMethod(), "█")
	if thumbCell != nil && s.scrollbarColor != nil {
		thumbCell.Style = uv.Style{Fg: s.scrollbarColor}
	}

	// Draw scrollbar
	for y := scrollbarStart; y < scrollbarEnd; y++ {
		if y >= thumbPos && y < thumbPos+thumbHeight {
			scr.SetCell(scrollbarX, y, thumbCell)
		} else {
			scr.SetCell(scrollbarX, y, trackCell)
		}
	}
}

// drawHorizontalScrollbar draws a horizontal scrollbar.
func (s *ScrollView) drawHorizontalScrollbar(scr uv.Screen, area uv.Rectangle, contentWidth, viewportWidth, scrollbarHeight int) {
	if contentWidth <= viewportWidth {
		return
	}

	scrollbarY := area.Max.Y - scrollbarHeight
	scrollbarStart := area.Min.X
	scrollbarEnd := area.Max.X
	trackWidth := scrollbarEnd - scrollbarStart

	// Calculate scrollbar thumb size
	thumbWidth := max(1, (viewportWidth*trackWidth)/contentWidth)

	// Calculate scrollbar thumb position
	scrollableRange := contentWidth - viewportWidth
	trackRange := trackWidth - thumbWidth

	thumbPos := scrollbarStart
	if scrollableRange > 0 {
		thumbPos = scrollbarStart + (s.offsetX*trackRange)/scrollableRange
	}

	// Ensure thumb stays within bounds (handle rounding edge cases)
	if thumbPos+thumbWidth > scrollbarEnd {
		thumbPos = scrollbarEnd - thumbWidth
	}
	if thumbPos < scrollbarStart {
		thumbPos = scrollbarStart
	}

	trackCell := uv.NewCell(scr.WidthMethod(), "░")
	thumbCell := uv.NewCell(scr.WidthMethod(), "█")
	if thumbCell != nil && s.scrollbarColor != nil {
		thumbCell.Style = uv.Style{Fg: s.scrollbarColor}
	}

	for x := scrollbarStart; x < scrollbarEnd; x++ {
		if x >= thumbPos && x < thumbPos+thumbWidth {
			scr.SetCell(x, scrollbarY, thumbCell)
		} else {
			scr.SetCell(x, scrollbarY, trackCell)
		}
	}
}

// Layout calculates the scroll view size.
func (s *ScrollView) Layout(constraints Constraints) Size {
	// Start with max available
	viewportWidth := constraints.MaxWidth
	viewportHeight := constraints.MaxHeight

	// Apply width/height constraints if specified
	if !s.width.IsAuto() {
		viewportWidth = s.width.Apply(constraints.MaxWidth, constraints.MaxWidth)
	}

	if !s.height.IsAuto() {
		viewportHeight = s.height.Apply(constraints.MaxHeight, constraints.MaxHeight)
	}

	// Constrain final size
	return Size{
		Width:  min(viewportWidth, constraints.MaxWidth),
		Height: min(viewportHeight, constraints.MaxHeight),
	}
}

// Children returns the child element.
func (s *ScrollView) Children() []Element {
	if s.child == nil {
		return nil
	}
	return []Element{s.child}
}

// ContentSize returns the full size of the content.
func (s *ScrollView) ContentSize() Size {
	if s.child == nil {
		return Size{Width: 0, Height: 0}
	}

	// Layout with unbounded constraints to get full size
	unbounded := Constraints{
		MinWidth:  0,
		MaxWidth:  1 << 30,
		MinHeight: 0,
		MaxHeight: 1 << 30,
	}

	return s.child.Layout(unbounded)
}

// adjustChildBounds recursively adjusts the bounds of all child elements
// to account for the scroll view's viewport position and scroll offset.
// This ensures hit testing works correctly for elements inside scroll views.
func (s *ScrollView) adjustChildBounds(elem Element, offsetX, offsetY int) {
	if elem == nil {
		return
	}

	// Get current bounds (relative to content buffer at 0,0)
	bounds := elem.Bounds()

	// Translate to screen coordinates
	newBounds := uv.Rect(
		bounds.Min.X+offsetX,
		bounds.Min.Y+offsetY,
		bounds.Dx(),
		bounds.Dy(),
	)

	// Update the element's bounds
	elem.SetBounds(newBounds)

	// Recursively adjust children
	for _, child := range elem.Children() {
		if child != nil {
			s.adjustChildBounds(child, offsetX, offsetY)
		}
	}
}
