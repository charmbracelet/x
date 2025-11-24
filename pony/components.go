package pony

import uv "github.com/charmbracelet/ultraviolet"

// Badge represents a badge element (like "NEW", "BETA", status indicators).
type Badge struct {
	BaseElement
	Text  string
	Style uv.Style
}

var _ Element = (*Badge)(nil)

// NewBadge creates a badge component.
func NewBadge(props Props, children []Element) Element {
	text := props.Get("text")
	if text == "" && len(children) > 0 {
		// Use first child as text
		if t, ok := children[0].(*Text); ok {
			text = t.Content
		}
	}

	style := parseStyleAttr(props)

	return &Badge{
		Text:  text,
		Style: style,
	}
}

// Draw renders the badge.
func (b *Badge) Draw(scr uv.Screen, area uv.Rectangle) {
	b.SetBounds(area)

	if b.Text == "" {
		return
	}

	// Badges are rendered as [TEXT] with styling
	content := "[" + b.Text + "]"
	if !b.Style.IsZero() {
		content = b.Style.Styled(content)
	}

	styled := uv.NewStyledString(content)
	styled.Draw(scr, area)
}

// Layout calculates badge size.
func (b *Badge) Layout(constraints Constraints) Size {
	width := len(b.Text) + 2 // text + brackets
	return constraints.Constrain(Size{Width: width, Height: 1})
}

// Children returns nil.
func (b *Badge) Children() []Element {
	return nil
}

// Progress represents a progress bar element.
type Progress struct {
	BaseElement
	Value int
	Max   int
	Width SizeConstraint
	Style uv.Style
	Char  string
}

var _ Element = (*Progress)(nil)

// NewProgress creates a progress bar component.
func NewProgress(props Props, _ []Element) Element {
	value := parseIntAttr(props, "value", 0)
	maxValue := parseIntAttr(props, "max", 100)
	width := parseSizeConstraint(props.Get("width"))
	style := parseStyleAttr(props)
	char := props.GetOr("char", "█")

	return &Progress{
		Value: value,
		Max:   maxValue,
		Width: width,
		Style: style,
		Char:  char,
	}
}

// Draw renders the progress bar.
func (p *Progress) Draw(scr uv.Screen, area uv.Rectangle) {
	p.SetBounds(area)

	if area.Dx() == 0 {
		return
	}

	// Calculate filled portion
	filled := 0
	if p.Max > 0 {
		filled = min((area.Dx()*p.Value)/p.Max, area.Dx())
	}

	// Create cell for filled portion
	filledCell := uv.NewCell(scr.WidthMethod(), p.Char)
	if filledCell != nil && !p.Style.IsZero() {
		filledCell.Style = p.Style
	}

	// Create cell for empty portion
	emptyCell := uv.NewCell(scr.WidthMethod(), "░")

	// Draw progress bar
	for x := 0; x < area.Dx(); x++ {
		if x < filled {
			scr.SetCell(area.Min.X+x, area.Min.Y, filledCell)
		} else {
			scr.SetCell(area.Min.X+x, area.Min.Y, emptyCell)
		}
	}
}

// Layout calculates progress bar size.
func (p *Progress) Layout(constraints Constraints) Size {
	// Default width if not specified
	width := 20

	// Apply width constraint if specified
	if !p.Width.IsAuto() {
		// For fixed width, use the constraint value directly
		width = p.Width.Apply(constraints.MaxWidth, width)
	} else {
		// For auto, take available width
		width = constraints.MaxWidth
	}

	return Size{Width: width, Height: 1}
}

// Children returns nil.
func (p *Progress) Children() []Element {
	return nil
}

// Header is a styled header component.
type Header struct {
	BaseElement
	Text   string
	Level  int // 1-6, like HTML
	Style  uv.Style
	Border bool
}

var _ Element = (*Header)(nil)

// NewHeader creates a header component.
func NewHeader(props Props, children []Element) Element {
	text := props.Get("text")
	if text == "" && len(children) > 0 {
		if t, ok := children[0].(*Text); ok {
			text = t.Content
		}
	}

	level := parseIntAttr(props, "level", 1)
	level = max(1, min(6, level))

	style := parseStyleAttr(props)
	border := parseBoolAttr(props, "border", true)

	return &Header{
		Text:   text,
		Level:  level,
		Style:  style,
		Border: border,
	}
}

// Draw renders the header.
func (h *Header) Draw(scr uv.Screen, area uv.Rectangle) {
	h.SetBounds(area)

	if h.Text == "" {
		return
	}

	// Style based on level
	content := h.Text
	if h.Style.IsZero() {
		// Default styles based on level
		defaultStyle := uv.Style{Attrs: uv.AttrBold}
		content = defaultStyle.Styled(content)
	} else {
		content = h.Style.Styled(content)
	}

	// Draw text
	styled := uv.NewStyledString(content)
	styled.Draw(scr, uv.Rect(area.Min.X, area.Min.Y, area.Dx(), 1))

	// Draw underline if border enabled
	if h.Border && area.Dy() > 1 {
		char := "─"
		underlineCell := uv.NewCell(scr.WidthMethod(), char)
		for x := 0; x < area.Dx(); x++ {
			scr.SetCell(area.Min.X+x, area.Min.Y+1, underlineCell)
		}
	}
}

// Layout calculates header size.
func (h *Header) Layout(constraints Constraints) Size {
	width := len(h.Text)
	height := 1
	if h.Border {
		height = 2 // text + underline
	}

	return constraints.Constrain(Size{Width: width, Height: height})
}

// Children returns nil.
func (h *Header) Children() []Element {
	return nil
}

// init registers built-in custom components.
func init() {
	Register("badge", NewBadge)
	Register("progress", NewProgress)
	Register("header", NewHeader)
}
