# `<pony />`

> **⚠️ EXPERIMENTAL**: This is an experimental project, primarily AI-generated as an exploration of declarative TUI frameworks. Use at your own risk.

[![Go Reference](https://pkg.go.dev/badge/github.com/charmbracelet/x/pony.svg)](https://pkg.go.dev/github.com/charmbracelet/x/pony)
[![Build Status](https://github.com/charmbracelet/x/actions/workflows/pony.yml/badge.svg)](https://github.com/charmbracelet/x/actions/workflows/pony.yml)

A declarative, type-safe markup language for building terminal user interfaces with [Ultraviolet](../ultraviolet) as the rendering engine.

```go
type ViewData struct {
    Title string
    Count int
}

const tmpl = `
<vstack gap="1">
    <box border="rounded" border-style="fg:cyan; bold">
        <text style="bold; fg:yellow">{{ .Title }}</text>
    </box>
    <text>Count: {{ .Count }}</text>
</vstack>
`

tmpl := pony.MustParse[ViewData](tmpl)
output := tmpl.Render(ViewData{Title: "My App", Count: 42}, 80, 24)
```

## Features

- ✅ **Type-safe templates** with Go generics
- ✅ **Full styling system** (colors, attributes, borders)
- ✅ **Responsive layouts** (%, auto, fixed sizes)
- ✅ **Advanced layout** (flex-grow, absolute positioning, layering)
- ✅ **Custom components** with clean API
- ✅ **Stateful components** via slots
- ✅ **Scrolling** with scrollbars
- ✅ **Bubble Tea integration**

## Quick Start

### Installation

```bash
go get github.com/charmbracelet/x/pony
```

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/charmbracelet/x/pony"
)

func main() {
    const tmpl = `
    <vstack gap="1">
        <box border="rounded">
            <text>Hello, World!</text>
        </box>
        <text>Welcome to pony!</text>
    </vstack>
    `

    t := pony.MustParse[interface{}](tmpl)
    output := t.Render(nil, 80, 24)
    fmt.Print(output)
}
```

## Elements

### Containers

**VStack** - Vertical stack
```xml
<vstack gap="1" align="center" width="50%" height="20">
    <!-- children -->
</vstack>
```
Attributes: `gap`, `align` (left|center|right), `width`, `height`

**HStack** - Horizontal stack
```xml
<hstack gap="2" valign="middle" width="100%">
    <!-- children -->
</hstack>
```
Attributes: `gap`, `valign` (top|middle|bottom), `width`, `height`

**ZStack** - Layered stack (overlays)
```xml
<zstack align="center" valign="middle">
    <box border="rounded">Background</box>
    <text style="bold">Overlay</text>
</zstack>
```
Attributes: `align` (left|center|right), `valign` (top|middle|bottom), `width`, `height`

Children are drawn on top of each other (later children on top).

### Content

**Text**
```xml
<text style="fg:cyan; bold" align="center" wrap="true">
    Content
</text>
```
Attributes: `style`, `align` (left|center|right), `wrap`

**Box** - Container (like a div)
```xml
<box padding="2" margin="1" width="50%">
    <!-- child -->
</box>

<!-- With border -->
<box border="rounded" border-style="fg:cyan" padding="2">
    <!-- child -->
</box>

<!-- Individual margins -->
<box margin-top="2" margin-left="3" margin-right="1" margin-bottom="2">
    <!-- child -->
</box>
```
Attributes: `border`, `border-style`, `padding`, `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left`, `width`, `height`

Border styles: `normal`, `rounded`, `thick`, `double`, `hidden`, `none` (default)

**Divider** - Separator line
```xml
<divider style="fg:gray" />
<divider vertical="true" char="|" />
```

**Spacer** - Empty space
```xml
<!-- Fixed spacer -->
<spacer size="2" />

<!-- Flexible spacer (grows to fill available space) -->
<spacer />
```

**ScrollView** - Scrollable viewport
```xml
<scrollview height="10" scrollbar="true">
    <!-- large content -->
</scrollview>
```

**Slot** - Dynamic content placeholder
```xml
<slot name="content" />
```

### Advanced Layout

**Flex** - Flexible sizing wrapper
```xml
<hstack>
    <box width="20">Fixed</box>
    <flex grow="1">
        <box>Grows 1x</box>
    </flex>
    <flex grow="2">
        <box>Grows 2x</box>
    </flex>
</hstack>
```
Attributes: `grow` (flex-grow factor), `shrink` (flex-shrink factor), `basis` (initial size)

**Positioned** - Absolute positioning
```xml
<zstack>
    <box>Background</box>

    <!-- Position from top-left -->
    <positioned x="10" y="5">
        <text>At (10,5)</text>
    </positioned>

    <!-- Position from edges -->
    <positioned right="2" bottom="1">
        <text>Bottom-right corner</text>
    </positioned>
</zstack>
```
Attributes: `x`, `y` (position from top-left), `right`, `bottom` (position from edges), `width`, `height`

Positioned elements don't affect parent layout (out of flow).

### Built-in Components

**Badge** - Status indicator
```xml
<badge text="NEW" style="fg:green; bold" />
```

**Progress** - Progress bar
```xml
<progress value="75" max="100" width="40" style="fg:green" />
```

**Header** - Section header
```xml
<header text="Title" style="fg:cyan; bold" border="true" />
```

**Button** - Clickable button
```xml
<button id="submit-btn" text="Submit" border="rounded" padding="1" style="fg:green" />
```
Attributes: `id`, `text`, `border`, `padding`, `style`, `width`, `height`

## Styling

### In Markup

```xml
<text style="fg:red; bg:black; bold; italic">Styled text</text>
```

**Colors:**
- Named: `fg:red`, `bg:blue`
- Hex: `fg:#FF5555`, `bg:#282a36`
- RGB: `fg:rgb(255,85,85)`
- ANSI: `fg:196`

**Attributes:**
`bold`, `italic`, `underline`, `strikethrough`, `faint`, `blink`, `reverse`

**Underline styles:**
`underline:single|double|curly|dotted|dashed`

### In Code (Helpers)

```go
style := pony.NewStyle().
    Fg(pony.Hex("#FF5555")).
    Bg(pony.RGB(40, 42, 54)).
    Bold().
    Italic().
    Build()

text := &pony.Text{
    Content: "Hello",
    Style:   style,
}
```

## Layout

### Sizing

```xml
<box width="50%">...</box>    <!-- Percentage -->
<box width="20">...</box>      <!-- Fixed cells -->
<box width="auto">...</box>    <!-- Content size (default) -->
<box width="min">...</box>     <!-- Minimum content size -->
<box width="max">...</box>     <!-- Maximum available -->
```

### Alignment

**Text:**
```xml
<text align="left|center|right">...</text>
```

**VStack children:**
```xml
<vstack align="left|center|right">...</vstack>
```

**HStack children:**
```xml
<hstack valign="top|middle|bottom">...</hstack>
```

**ZStack children:**
```xml
<zstack align="left|center|right" valign="top|middle|bottom">...</zstack>
```

### Flexible Sizing

Use `<spacer />` or `<flex>` for flexible layouts:

```xml
<vstack>
    <text>Header</text>
    <spacer />  <!-- Grows to fill space -->
    <text>Footer</text>
</vstack>

<hstack>
    <box width="20">Fixed sidebar</box>
    <flex grow="1">
        <box>Main content (grows)</box>
    </flex>
</hstack>
```

## Go Templates

### Variables

```xml
<text>Hello, {{ .Username }}!</text>
```

### Conditionals

```xml
{{ if .IsOnline }}
<text style="fg:green">● Online</text>
{{ else }}
<text style="fg:red">○ Offline</text>
{{ end }}
```

### Loops

```xml
{{ range .Items }}
<text>• {{ . }}</text>
{{ end }}
```

### Functions

Built-in: `upper`, `lower`, `title`, `trim`, `join`, `printf`, `add`, `sub`, `mul`, `div`, `repeat`

```xml
<text>{{ upper .Title }}</text>
<text>{{ printf "Count: %d" .Count }}</text>
```

## Custom Components

### Register a Component

```go
// Simple functional component
pony.Register("card", func(props pony.Props, children []pony.Element) pony.Element {
    titleStyle := pony.NewStyle().Bold().Build()

    return pony.NewBox(
        pony.NewVStack(
            pony.NewText(props.Get("title")).WithStyle(titleStyle),
            pony.NewDivider(),
            pony.NewVStack(children...),
        ),
    ).WithBorder("rounded").WithPadding(1)
})

// Or create a custom type for more control
type Card struct {
    pony.BaseElement  // Required for ID and bounds tracking
    Title   string
    Color   string
    Content []pony.Element
}

func NewCard(props pony.Props, children []pony.Element) pony.Element {
    return &Card{
        Title:   props.Get("title"),
        Color:   props.GetOr("color", "blue"),
        Content: children,
    }
}

func (c *Card) Draw(scr uv.Screen, area uv.Rectangle) {
    c.SetBounds(area) // Track bounds for mouse interaction

    // Build composed structure
    style := pony.NewStyle().Fg(pony.Hex("#00FFFF")).Bold().Build()
    card := pony.NewBox(
        pony.NewVStack(
            pony.NewText(c.Title).WithStyle(style),
            pony.NewDivider(),
            pony.NewVStack(c.Content...),
        ),
    ).WithBorder("rounded").WithPadding(1)

    card.Draw(scr, area)
}

func (c *Card) Layout(constraints pony.Constraints) pony.Size {
    // Delegate to composed structure
    style := pony.NewStyle().Fg(pony.Hex("#00FFFF")).Bold().Build()
    card := pony.NewBox(
        pony.NewVStack(
            pony.NewText(c.Title).WithStyle(style),
            pony.NewDivider(),
            pony.NewVStack(c.Content...),
        ),
    ).WithBorder("rounded").WithPadding(1)

    return card.Layout(constraints)
}

func (c *Card) Children() []pony.Element {
    return c.Content
}

// Register it
pony.Register("card", NewCard)
```

### Use in Markup

```xml
<card title="Profile">
    <text>Name: Alice</text>
    <text>Role: Developer</text>
</card>
```

## Stateful Components

Components with state use the slot system:

```go
type Input struct {
    value  string
    cursor int
}

func (i *Input) Update(msg tea.Msg) {
    // Handle input
}

func (i *Input) Render() pony.Element {
    return pony.NewBox(
        pony.NewText(i.value),
    ).WithBorder("rounded")
}
```

**Template with slots:**
```xml
<vstack>
    <text>Username:</text>
    <slot name="input" />
</vstack>
```

**Render with slots:**
```go
func (m model) View() tea.View {
    slots := map[string]pony.Element{
        "input": m.inputComp.Render(),
    }

    output := m.template.RenderWithSlots(data, slots, m.width, m.height)
    return tea.NewView(output)
}
```

## Bubble Tea Integration

```go
import (
    tea "charm.land/bubbletea/v2"
    "github.com/charmbracelet/x/pony"
)

type ViewData struct {
    Count int
}

type model struct {
    template *pony.Template[ViewData]
    count    int
    width    int
    height   int
}

func (m model) Init() tea.Cmd {
    return tea.RequestWindowSize
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    case tea.KeyPressMsg:
        if msg.String() == "space" {
            m.count++
        }
    }
    return m, nil
}

func (m model) View() tea.View {
    data := ViewData{Count: m.count}
    output := m.template.Render(data, m.width, m.height)
    return tea.NewView(output)
}
```

## Mouse Click Handling

pony provides stateless mouse click handling through bounds tracking and hit testing. All elements are interactive by default with no state mutation in View.

### Quick Example

```go
type buttonClickMsg string

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case buttonClickMsg:
        switch msg {
        case "submit-btn":
            return m, m.submitForm()
        case "cancel-btn":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() tea.View {
    data := ViewData{...}

    // Render and get bounds map
    scr, boundsMap := m.template.RenderWithBounds(data, nil, m.width, m.height)

    view := tea.NewView(scr.Render())
    view.MouseMode = tea.MouseModeAllMotion  // Enable mouse events

    // Callback captures boundsMap - no model mutation!
    view.Callback = func(msg tea.Msg) tea.Cmd {
        if click, ok := msg.(tea.MouseClickMsg); ok {
            mouse := click.Mouse()

            // Find which element was clicked
            if elem := boundsMap.HitTest(mouse.X, mouse.Y); elem != nil {
                return func() tea.Msg {
                    return buttonClickMsg(elem.ID())
                }
            }
        }
        return nil
    }

    return view
}
```

### Set Element IDs

**In markup:**
```xml
<button id="submit-btn" text="Submit" />
<button id="cancel-btn" text="Cancel" />
```

**Programmatically:**
```go
btn := pony.NewButton("Submit")
btn.SetID("submit-btn")
```

**For custom components that render other elements:**
```go
// IMPORTANT: Pass through your component's ID to the rendered element
// so clicks anywhere in the component return your component's ID

func (i *Input) Render() pony.Element {
    vstack := pony.NewVStack(
        pony.NewText(i.label),
        pony.NewBox(pony.NewText(i.value)).WithBorder("rounded"),
    )

    // Set the input's ID on the VStack so clicks return "my-input", not child IDs
    vstack.SetID(i.ID())

    return vstack
}

// Usage
input := NewInput("Name:")
input.SetID("name-input")  // When clicked, returns "name-input"
```

### Button Component

```xml
<!-- Basic button -->
<button id="my-btn" text="Click Me" />

<!-- Styled button -->
<button id="submit" text="Submit"
        border="rounded"
        padding="1"
        style="fg:green; bold" />
```

### BoundsMap API

```go
// Hit test - find element at coordinates
// Prefers elements with explicit IDs over auto-generated ones
elem := boundsMap.HitTest(x, y) // Returns Element or nil

// Get element by ID
elem, ok := boundsMap.GetByID("button-id")

// Get bounds for element ID
bounds, ok := boundsMap.GetBounds("button-id")

// Get all elements with bounds
elements := boundsMap.AllElements() // []ElementWithBounds
```

**Important:** `HitTest()` prefers elements with explicit IDs when multiple elements overlap. This ensures that clicking inside a component returns the component's ID, not its children's IDs. Always use `SetID()` on the root element your component returns (see example above).

### Hover Detection

```go
type hoverMsg string

view.MouseMode = tea.MouseModeAllMotion  // Enable all mouse motion

view.Callback = func(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.MouseClickMsg:
        // Handle clicks

    case tea.MouseMotionMsg:
        // Handle hover
        mouse := msg.Mouse()
        if elem := boundsMap.HitTest(mouse.X, mouse.Y); elem != nil {
            return func() tea.Msg {
                return hoverMsg(elem.ID())
            }
        }
    }
    return nil
}
```

### Requirements

Mouse handling requires Bubble Tea PR #1549 (View callback support). Pin to this commit:

```go
require (
    charm.land/bubbletea/v2 v2.0.0-20250120210912-18cfb8c3ccb3
)
```

### How It Works

1. **RenderWithBounds()** returns screen + BoundsMap (every element's position)
2. **View.Callback** captures BoundsMap via closure (no model mutation!)
3. **HitTest()** finds which element is at mouse coordinates
   - **Prefers elements with explicit IDs** over auto-generated IDs
   - This means clicking inside a component returns the component's ID, not child IDs
4. **Callback returns Cmd** with element ID
5. **Update()** handles custom messages based on element ID

**Benefits:**
- ✅ Pure View() - no state mutation
- ✅ Stateless - BoundsMap is immutable
- ✅ Universal - all elements interactive by default
- ✅ Type-safe - element IDs are strings
- ✅ Single render - no double-rendering needed
- ✅ Smart hit testing - prefers meaningful IDs

## Examples

See [examples/](./examples) for complete working examples:

- **hello** - Basic hello world
- **layout** - Responsive layouts
- **styled** - Styling showcase
- **dynamic** - Type-safe templates
- **alignment** - Alignment & padding
- **components** - Built-in components
- **custom** - Custom components
- **stateful** - Stateful components
- **scrolling** - Scrollable views
- **helpers** - Style helpers
- **advanced** - Advanced layout (ZStack, Flex, Positioned, Margin)
- **simple-bubbletea** - Minimal Bubble Tea app
- **bubbletea** - Full interactive app
- **buttons** - Mouse click handling with interactive buttons
- **interactive-form** - Complex form with slots, validation, and mouse interactions

## API Reference

### Template

```go
// Parse with type safety
tmpl, err := pony.Parse[YourDataType](markup)
tmpl := pony.MustParse[YourDataType](markup)

// Render
output := tmpl.Render(data, width, height)
output := tmpl.RenderWithSlots(data, slots, width, height)

// Render with bounds for mouse handling
scr, boundsMap := tmpl.RenderWithBounds(data, slots, width, height)
```

### Element Constructors

```go
pony.NewText(content)
pony.NewBox(child)
pony.NewVStack(children...)
pony.NewHStack(children...)
pony.NewZStack(children...)
pony.NewButton(text)
pony.NewDivider()
pony.NewSpacer()
pony.NewFlex(child)
pony.NewPositioned(child, x, y)
pony.NewSlot(name)
pony.NewScrollView(child)
```

### Fluent API

```go
box := pony.NewBox(child).
    WithBorder("rounded").
    WithPadding(2).
    WithMargin(1).
    WithMarginTop(2).
    WithWidth(pony.NewFixedConstraint(50)).
    WithBorderStyle(style)

text := pony.NewText("Hello").
    WithStyle(style).
    WithAlign("center").
    WithWrap(true)

button := pony.NewButton("Click Me").
    WithBorder("rounded").
    WithPadding(1).
    WithStyle(style).
    WithWidth(pony.NewFixedConstraint(20))
button.SetID("my-button")

flex := pony.NewFlex(child).
    WithGrow(1).
    WithShrink(0).
    WithBasis(20)

positioned := pony.NewPositioned(child, 10, 5).
    WithRight(2).
    WithBottom(1)
```

### Style Builder

```go
style := pony.NewStyle().
    Fg(pony.Hex("#FF5555")).
    Bg(pony.RGB(40, 42, 54)).
    Bold().
    Italic().
    Underline().
    Build()
```

### Component Registry

```go
pony.Register(name, factory)
pony.Unregister(name)
pony.GetComponent(name)
pony.RegisteredComponents()
```

### Layout Helpers

```go
// Basic layouts
pony.Panel(child, border, padding)
pony.PanelWithMargin(child, border, padding, margin)
pony.Card(title, titleStyle, borderStyle, children...)
pony.Section(header, headerStyle, children...)
pony.Separated(children...)

// Advanced layouts
pony.Overlay(children...)  // ZStack with default alignment
pony.FlexGrow(child, grow)
pony.Position(child, x, y)
pony.PositionRight(child, right, y)
pony.PositionBottom(child, x, bottom)
pony.PositionCorner(child, right, bottom)
```

## Architecture

```
Template[T] → Go Template → XML Parse → Element Tree → Fill Slots → Layout → UV Render
```
