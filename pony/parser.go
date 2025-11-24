package pony

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"strings"
)

// node represents a parsed XML element.
type node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []*node    `xml:",any"`
}

// Props converts XML attributes to Props map.
func (n *node) Props() Props {
	props := make(Props)
	for _, attr := range n.Attrs {
		props[attr.Name.Local] = attr.Value
	}
	return props
}

// parse parses XML markup into a node tree.
func parse(markup string) (*node, error) {
	// Wrap in root element if not already wrapped
	wrapped := markup
	if !strings.HasPrefix(strings.TrimSpace(markup), "<") {
		wrapped = "<root>" + markup + "</root>"
	}

	decoder := xml.NewDecoder(strings.NewReader(wrapped))
	decoder.Strict = false // Be lenient with XML parsing

	var root node
	if err := decoder.Decode(&root); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty markup")
		}
		return nil, fmt.Errorf("xml decode: %w", err)
	}

	// If we added a wrapper, unwrap it
	if strings.TrimSpace(markup) != wrapped {
		if len(root.Children) == 1 {
			return root.Children[0], nil
		}
		return &root, nil
	}

	return &root, nil
}

// toElement converts an XML node to an Element.
func (n *node) toElement() Element {
	if n == nil {
		return nil
	}

	// Get the tag name
	tagName := n.XMLName.Local
	props := n.Props()

	var elem Element

	// Check custom component registry first
	if factory, ok := GetComponent(tagName); ok {
		elem = factory(props, n.childElements())
	} else {
		// Then check built-in elements
		switch tagName {
		case "vstack":
			elem = n.toVStack(props)
		case "hstack":
			elem = n.toHStack(props)
		case "zstack":
			elem = n.toZStack(props)
		case "text":
			elem = n.toText(props)
		case "box":
			elem = n.toBox(props)
		case "spacer":
			elem = n.toSpacer(props)
		case "flex":
			elem = n.toFlex(props)
		case "positioned":
			elem = n.toPositioned(props)
		case "divider":
			elem = n.toDivider(props)
		case "slot":
			elem = n.toSlot(props)
		case "scrollview":
			elem = n.toScrollView(props)
		case "":
			// Anonymous text node (no tag, just content)
			content := strings.TrimSpace(n.Content)
			if content != "" {
				elem = NewText(content)
			} else {
				return nil
			}
		default:
			// Unknown element, treat as a container
			elem = NewVStack(n.childElements()...)
		}
	}

	// Set ID if provided
	if elem != nil && props.Has("id") {
		if setter, ok := elem.(interface{ SetID(string) }); ok {
			setter.SetID(props.Get("id"))
		}
	}

	return elem
}

// childElements converts child nodes to Elements.
func (n *node) childElements() []Element {
	var elements []Element

	for _, child := range n.Children {
		// Check if this is a text node
		if child.XMLName.Local == "" && strings.TrimSpace(child.Content) != "" {
			elements = append(elements, NewText(strings.TrimSpace(child.Content)))
			continue
		}

		if elem := child.toElement(); elem != nil {
			elements = append(elements, elem)
		}
	}

	return elements
}

// toVStack converts node to VStack element.
func (n *node) toVStack(props Props) Element {
	spacing := parseIntAttr(props, "spacing", 0)
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	alignment := props.GetOr("alignment", AlignmentLeading)

	vstack := NewVStack(n.childElements()...)
	if spacing > 0 {
		vstack = vstack.Spacing(spacing)
	}
	if !width.IsAuto() {
		vstack = vstack.Width(width)
	}
	if !height.IsAuto() {
		vstack = vstack.Height(height)
	}
	if alignment != AlignmentLeading {
		vstack = vstack.Alignment(alignment)
	}

	return vstack
}

// toHStack converts node to HStack element.
func (n *node) toHStack(props Props) Element {
	spacing := parseIntAttr(props, "spacing", 0)
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	alignment := props.GetOr("alignment", AlignmentTop)

	hstack := NewHStack(n.childElements()...)
	if spacing > 0 {
		hstack = hstack.Spacing(spacing)
	}
	if !width.IsAuto() {
		hstack = hstack.Width(width)
	}
	if !height.IsAuto() {
		hstack = hstack.Height(height)
	}
	if alignment != AlignmentTop {
		hstack = hstack.Alignment(alignment)
	}

	return hstack
}

// toZStack converts node to ZStack element.
func (n *node) toZStack(props Props) Element {
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	alignment := props.GetOr("alignment", AlignmentCenter)
	verticalAlignment := props.GetOr("vertical-alignment", AlignmentCenter)

	zstack := NewZStack(n.childElements()...)
	if !width.IsAuto() {
		zstack = zstack.Width(width)
	}
	if !height.IsAuto() {
		zstack = zstack.Height(height)
	}
	if alignment != AlignmentCenter {
		zstack = zstack.Alignment(alignment)
	}
	if verticalAlignment != AlignmentCenter {
		zstack = zstack.VerticalAlignment(verticalAlignment)
	}

	return zstack
}

// toText converts node to Text element.
func (n *node) toText(props Props) Element {
	// Collect text from content and children
	var text string

	if n.Content != "" {
		text = strings.TrimSpace(n.Content)
	}

	// Also collect text from child text nodes
	for _, child := range n.Children {
		if child.XMLName.Local == "" && child.Content != "" {
			if text != "" {
				text += " "
			}
			text += strings.TrimSpace(child.Content)
		}
	}

	textElem := NewText(text)

	// Parse granular style attributes
	if fontWeight := props.Get("font-weight"); fontWeight == FontWeightBold {
		textElem = textElem.Bold()
	}

	if fontStyle := props.Get("font-style"); fontStyle == FontStyleItalic {
		textElem = textElem.Italic()
	}

	if decoration := props.Get("text-decoration"); decoration != "" {
		switch decoration {
		case DecorationUnderline:
			textElem = textElem.Underline()
		case DecorationStrikethrough:
			textElem = textElem.Strikethrough()
		}
	}

	if fgColor := props.Get("foreground-color"); fgColor != "" {
		if c, err := parseColor(fgColor); err == nil {
			textElem = textElem.ForegroundColor(c)
		}
	}

	if bgColor := props.Get("background-color"); bgColor != "" {
		if c, err := parseColor(bgColor); err == nil {
			textElem = textElem.BackgroundColor(c)
		}
	}

	if wrap := parseBoolAttr(props, "wrap", false); wrap {
		textElem = textElem.Wrap(true)
	}

	if alignment := props.Get("alignment"); alignment != "" {
		textElem = textElem.Alignment(alignment)
	}

	return textElem
}

// toBox converts node to Box element.
func (n *node) toBox(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	// Parse border color if present
	borderColorStr := props.Get("border-color")
	var borderColor color.Color
	if borderColorStr != "" {
		if c, err := parseColor(borderColorStr); err == nil {
			borderColor = c
		}
	}

	// Parse width and height constraints
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))

	// Parse padding
	padding := parseIntAttr(props, "padding", 0)

	// Parse margin
	margin := parseIntAttr(props, "margin", 0)
	marginTop := parseIntAttr(props, "margin-top", 0)
	marginRight := parseIntAttr(props, "margin-right", 0)
	marginBottom := parseIntAttr(props, "margin-bottom", 0)
	marginLeft := parseIntAttr(props, "margin-left", 0)

	box := NewBox(child)
	if border := props.Get("border"); border != "" {
		box = box.Border(border)
	}
	if borderColor != nil {
		box = box.BorderColor(borderColor)
	}
	if !width.IsAuto() {
		box = box.Width(width)
	}
	if !height.IsAuto() {
		box = box.Height(height)
	}
	if padding > 0 {
		box = box.Padding(padding)
	}
	if margin > 0 {
		box = box.Margin(margin)
	}
	if marginTop > 0 {
		box = box.MarginTop(marginTop)
	}
	if marginRight > 0 {
		box = box.MarginRight(marginRight)
	}
	if marginBottom > 0 {
		box = box.MarginBottom(marginBottom)
	}
	if marginLeft > 0 {
		box = box.MarginLeft(marginLeft)
	}

	return box
}

// toSpacer converts node to Spacer element.
func (n *node) toSpacer(props Props) Element {
	size := parseIntAttr(props, "size", 0)
	if size > 0 {
		return NewFixedSpacer(size)
	}
	return NewSpacer()
}

// toFlex converts node to Flex element.
func (n *node) toFlex(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	grow := parseIntAttr(props, "grow", 0)
	shrink := parseIntAttr(props, "shrink", 1)
	basis := parseIntAttr(props, "basis", 0)

	flex := NewFlex(child)
	if grow > 0 {
		flex = flex.Grow(grow)
	}
	if shrink != 1 {
		flex = flex.Shrink(shrink)
	}
	if basis > 0 {
		flex = flex.Basis(basis)
	}

	return flex
}

// toPositioned converts node to Positioned element.
func (n *node) toPositioned(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	x := parseIntAttr(props, "x", 0)
	y := parseIntAttr(props, "y", 0)
	right := parseIntAttr(props, "right", -1)
	bottom := parseIntAttr(props, "bottom", -1)
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))

	positioned := NewPositioned(child, x, y)
	if right >= 0 {
		positioned = positioned.Right(right)
	}
	if bottom >= 0 {
		positioned = positioned.Bottom(bottom)
	}
	if !width.IsAuto() {
		positioned = positioned.Width(width)
	}
	if !height.IsAuto() {
		positioned = positioned.Height(height)
	}

	return positioned
}

// toDivider converts node to Divider element.
func (n *node) toDivider(props Props) Element {
	vertical := parseBoolAttr(props, "vertical", false)
	char := props.Get("char")

	var divider *Divider
	if vertical {
		divider = NewVerticalDivider()
	} else {
		divider = NewDivider()
	}

	// Parse foreground color
	if fgColor := props.Get("foreground-color"); fgColor != "" {
		if c, err := parseColor(fgColor); err == nil {
			divider = divider.ForegroundColor(c)
		}
	}

	if char != "" {
		divider = divider.Char(char)
	}

	return divider
}

// toSlot converts node to Slot element.
func (n *node) toSlot(props Props) Element {
	name := props.Get("name")
	if name == "" {
		// Slot requires a name
		name = "unnamed"
	}

	return NewSlot(name)
}

// toScrollView converts node to ScrollView element.
func (n *node) toScrollView(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	// Parse dimensions
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))

	// Parse scroll options
	offsetX := parseIntAttr(props, "offset-x", 0)
	offsetY := parseIntAttr(props, "offset-y", 0)
	showScrollbar := parseBoolAttr(props, "scrollbar", true)
	vertical := parseBoolAttr(props, "vertical", true)
	horizontal := parseBoolAttr(props, "horizontal", false)

	// Parse scrollbar color
	var scrollbarColor color.Color
	if colorStr := props.Get("scrollbar-color"); colorStr != "" {
		if c, err := parseColor(colorStr); err == nil {
			scrollbarColor = c
		}
	}

	scrollView := NewScrollView(child)
	if offsetX != 0 || offsetY != 0 {
		scrollView = scrollView.Offset(offsetX, offsetY)
	}
	if !width.IsAuto() {
		scrollView = scrollView.Width(width)
	}
	if !height.IsAuto() {
		scrollView = scrollView.Height(height)
	}
	if !showScrollbar {
		scrollView = scrollView.Scrollbar(false)
	}
	if !vertical {
		scrollView = scrollView.Vertical(false)
	}
	if horizontal {
		scrollView = scrollView.Horizontal(true)
	}
	if scrollbarColor != nil {
		scrollView = scrollView.ScrollbarColor(scrollbarColor)
	}

	return scrollView
}

// Helper functions for parsing attributes

func parseIntAttr(props Props, key string, defaultValue int) int {
	if val := props.Get(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseBoolAttr(props Props, key string, defaultValue bool) bool {
	val := props.Get(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1" || val == "yes"
}
