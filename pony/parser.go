package pony

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
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
				elem = &Text{Content: content}
			} else {
				return nil
			}
		default:
			// Unknown element, treat as a container
			elem = &VStack{Items: n.childElements()}
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
			elements = append(elements, &Text{Content: strings.TrimSpace(child.Content)})
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
	gap := parseIntAttr(props, "gap", 0)
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	align := props.GetOr("align", AlignLeft)

	return &VStack{
		Gap:    gap,
		Items:  n.childElements(),
		Width:  width,
		Height: height,
		Align:  align,
	}
}

// toHStack converts node to HStack element.
func (n *node) toHStack(props Props) Element {
	gap := parseIntAttr(props, "gap", 0)
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	valign := props.GetOr("valign", AlignTop)

	return &HStack{
		Gap:    gap,
		Items:  n.childElements(),
		Width:  width,
		Height: height,
		Valign: valign,
	}
}

// toZStack converts node to ZStack element.
func (n *node) toZStack(props Props) Element {
	width := parseSizeConstraint(props.Get("width"))
	height := parseSizeConstraint(props.Get("height"))
	align := props.GetOr("align", AlignCenter)
	valign := props.GetOr("valign", AlignMiddle)

	return &ZStack{
		Items:  n.childElements(),
		Width:  width,
		Height: height,
		Align:  align,
		Valign: valign,
	}
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

	// Parse style if present
	style := parseStyleAttr(props)

	return &Text{
		Content: text,
		Wrap:    parseBoolAttr(props, "wrap", false),
		Align:   props.GetOr("align", AlignLeft),
		Style:   style,
	}
}

// toBox converts node to Box element.
func (n *node) toBox(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	// Parse border style if present
	borderStyleStr := props.Get("border-style")
	var borderStyle uv.Style
	if borderStyleStr != "" {
		if s, err := ParseStyle(borderStyleStr); err == nil {
			borderStyle = s
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

	return &Box{
		Child:        child,
		Border:       props.GetOr("border", BorderNone),
		BorderStyle:  borderStyle,
		Width:        width,
		Height:       height,
		Padding:      padding,
		Margin:       margin,
		MarginTop:    marginTop,
		MarginRight:  marginRight,
		MarginBottom: marginBottom,
		MarginLeft:   marginLeft,
	}
}

// toSpacer converts node to Spacer element.
func (n *node) toSpacer(props Props) Element {
	return &Spacer{
		Size: parseIntAttr(props, "size", 0),
	}
}

// toFlex converts node to Flex element.
func (n *node) toFlex(props Props) Element {
	var child Element
	children := n.childElements()
	if len(children) > 0 {
		child = children[0]
	}

	return &Flex{
		Child:  child,
		Grow:   parseIntAttr(props, "grow", 0),
		Shrink: parseIntAttr(props, "shrink", 1),
		Basis:  parseIntAttr(props, "basis", 0),
	}
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

	return &Positioned{
		Child:  child,
		X:      x,
		Y:      y,
		Right:  right,
		Bottom: bottom,
		Width:  width,
		Height: height,
	}
}

// toDivider converts node to Divider element.
func (n *node) toDivider(props Props) Element {
	style := parseStyleAttr(props)

	return &Divider{
		Vertical: parseBoolAttr(props, "vertical", false),
		Char:     props.Get("char"),
		Style:    style,
	}
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

	// Parse scrollbar style
	var scrollbarStyle uv.Style
	if styleStr := props.Get("scrollbar-style"); styleStr != "" {
		if s, err := ParseStyle(styleStr); err == nil {
			scrollbarStyle = s
		}
	}

	return &ScrollView{
		Child:          child,
		OffsetX:        offsetX,
		OffsetY:        offsetY,
		Width:          width,
		Height:         height,
		ShowScrollbar:  showScrollbar,
		ScrollbarStyle: scrollbarStyle,
		Vertical:       vertical,
		Horizontal:     horizontal,
	}
}

// Helper functions for parsing attributes

func parseStyleAttr(props Props) uv.Style {
	styleStr := props.Get("style")
	if styleStr == "" {
		return uv.Style{}
	}

	style, err := ParseStyle(styleStr)
	if err != nil {
		// Log error but don't fail, just return empty style
		return uv.Style{}
	}
	return style
}

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
