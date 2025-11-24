package pony

import (
	"bytes"
	"fmt"
	"maps"
	"strings"
	"text/template"

	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Template is a type-safe pony template that can be rendered with data of type T.
type Template[T any] struct {
	markup   string
	goTmpl   *template.Template
	cacheKey string
}

// Parse parses pony markup into a type-safe template.
// The markup can contain Go template syntax like {{ .Variable }}.
func Parse[T any](markup string) (*Template[T], error) {
	return ParseWithFuncs[T](markup, nil)
}

// ParseWithFuncs parses pony markup with custom template functions.
func ParseWithFuncs[T any](markup string, funcs template.FuncMap) (*Template[T], error) {
	t := &Template[T]{
		markup:   markup,
		cacheKey: markup,
	}

	// Create Go template with builtin functions
	tmplFuncs := defaultTemplateFuncs()
	maps.Copy(tmplFuncs, funcs)

	goTmpl, err := template.New("pony").Funcs(tmplFuncs).Parse(markup)
	if err != nil {
		return nil, fmt.Errorf("template parse: %w", err)
	}
	t.goTmpl = goTmpl

	return t, nil
}

// MustParse parses pony markup and panics on error.
func MustParse[T any](markup string) *Template[T] {
	t, err := Parse[T](markup)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParseWithFuncs parses pony markup with custom functions and panics on error.
func MustParseWithFuncs[T any](markup string, funcs template.FuncMap) *Template[T] {
	t, err := ParseWithFuncs[T](markup, funcs)
	if err != nil {
		panic(err)
	}
	return t
}

// Render renders the template with the given data to the specified viewport size.
func (t *Template[T]) Render(data T, width, height int) string {
	scr, _ := t.RenderWithBounds(data, nil, width, height)
	str := scr.Render()
	return strings.ReplaceAll(str, "\r\n", "\n")
}

// RenderWithBounds renders the template and returns both the screen buffer and bounds map.
// The bounds map can be used for mouse hit testing in event handlers.
func (t *Template[T]) RenderWithBounds(data T, slots map[string]Element, width, height int) (uv.ScreenBuffer, *BoundsMap) {
	// Execute Go template first
	var buf bytes.Buffer
	if err := t.goTmpl.Execute(&buf, data); err != nil {
		errScreen := uv.NewScreenBuffer(width, 1)
		return errScreen, NewBoundsMap()
	}

	processedMarkup := buf.String()

	// Parse the processed markup
	root, err := parse(processedMarkup)
	if err != nil {
		errScreen := uv.NewScreenBuffer(width, 1)
		return errScreen, NewBoundsMap()
	}

	// Convert to element tree
	elem := root.toElement()
	if elem == nil {
		emptyScreen := uv.NewScreenBuffer(width, height)
		return emptyScreen, NewBoundsMap()
	}

	// Fill slots with provided elements
	if slots != nil {
		fillSlots(elem, slots)
	}

	// Layout the element
	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: height,
	}
	size := elem.Layout(constraints)

	// Use the smaller of calculated vs requested size
	if size.Width > width {
		size.Width = width
	}
	if size.Height > height {
		size.Height = height
	}

	// Create buffer and render
	uvBuf := uv.NewScreenBuffer(size.Width, size.Height)
	area := uv.Rect(0, 0, size.Width, size.Height)
	elem.Draw(uvBuf, area)

	// Build bounds map
	boundsMap := NewBoundsMap()
	walkAndRegister(elem, boundsMap)

	return uvBuf, boundsMap
}

// RenderWithSlots renders the template with data and slot elements.
// Slots allow injecting stateful components into the template.
func (t *Template[T]) RenderWithSlots(data T, slots map[string]Element, width, height int) string {
	scr, _ := t.RenderWithBounds(data, slots, width, height)
	str := scr.Render()
	return strings.ReplaceAll(str, "\r\n", "\n")
}

// fillSlots recursively fills slot elements with their corresponding elements.
func fillSlots(elem Element, slots map[string]Element) {
	if slot, ok := elem.(*Slot); ok {
		if slotElem, found := slots[slot.Name]; found {
			slot.setElement(slotElem)
		}
		return
	}

	// Recursively fill slots in children
	for _, child := range elem.Children() {
		fillSlots(child, slots)
	}
}

// defaultTemplateFuncs returns the default template functions.
func defaultTemplateFuncs() template.FuncMap {
	titleCaser := cases.Title(language.English)
	return template.FuncMap{
		// String functions
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": titleCaser.String,
		"trim":  strings.TrimSpace,
		"join":  strings.Join,

		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},

		// Formatting
		"printf": fmt.Sprintf,
		"repeat": strings.Repeat,

		// Color helpers
		"colorHex": func(hex string) string {
			return fmt.Sprintf("fg:%s", hex)
		},
		"bgHex": func(hex string) string {
			return fmt.Sprintf("bg:%s", hex)
		},
	}
}
