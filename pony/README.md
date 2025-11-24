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
<vstack spacing="1">
    <box border="rounded" border-color="cyan">
        <text font-weight="bold" foreground-color="yellow">{{ .Title }}</text>
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
    <vstack spacing="1">
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
<vstack spacing="1" alignment="center" width="50%" height="20">
    <!-- children -->
</vstack>
```
Attributes: `spacing`, `alignment` (leading|center|trailing), `width`, `height`

**HStack** - Horizontal stack
```xml
<hstack spacing="2" alignment="center" width="100%">
    <!-- children -->
</hstack>
```
Attributes: `spacing`, `alignment` (top|center|bottom), `width`, `height`

**ZStack** - Layered stack (overlays)
```xml
<zstack alignment="center" vertical-alignment="center">
    <box border="rounded">Background</box>
    <text font-weight="bold">Overlay</text>
</zstack>
```
Attributes: `alignment` (leading|center|trailing), `vertical-alignment` (top|center|bottom), `width`, `height`

Children are drawn on top of each other (later children on top).

### Content

**Text**
```xml
<text foreground-color="cyan" font-weight="bold" alignment="center" wrap="true">
    Content
</text>
```
Attributes: `foreground-color`, `background-color`, `font-weight` (bold), `font-style` (italic), `text-decoration` (underline|strikethrough), `alignment` (leading|center|trailing), `wrap`

**Box** - Container (like a div)
```xml
<box padding="2" margin="1" width="50%">
    <!-- child -->
</box>

<!-- With border -->
<box border="rounded" border-color="cyan" padding="2">
    <!-- child -->
</box>

<!-- Individual margins -->
<box margin-top="2" margin-left="3" margin-right="1" margin-bottom="2">
    <!-- child -->
</box>
```
Attributes: `border`, `border-color`, `padding`, `margin`, `margin-top`, `margin-right`, `margin-bottom`, `margin-left`, `width`, `height`

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

**ProgressView** - Progress bar
```xml
<progressview value="75" max="100" width="40" style="fg:green" />
```

**Button** - Clickable button
```xml
<button id="submit-btn" text="Submit" border="rounded" padding="1" />
```
Attributes: `id`, `text`, `border`, `padding`, `width`, `height`

## Styling

### In Markup

```xml
<text foreground-color="red" background-color="black" font-weight="bold" font-style="italic">Styled text</text>
```

**Text Attributes:**
- `foreground-color` - Text color (named, hex, rgb)
- `background-color` - Background color  
- `font-weight` - `bold` or omit for normal
- `font-style` - `italic` or omit for normal
- `text-decoration` - `underline` or `strikethrough`
- `alignment` - `leading`, `center`, `trailing`

**Colors:**
- Named: `red`, `blue`, `green`, `cyan`, `yellow`, `magenta`, `white`, `black`, `gray`
- Hex: `#FF5555`, `#282a36`
- RGB: `rgb(255,85,85)`
- ANSI: `196`

### In Code (Fluent API)

```go
text := pony.NewText("Hello").
    ForegroundColor(pony.Hex("#FF5555")).
    BackgroundColor(pony.RGB(40, 42, 54)).
    Bold().
    Italic().
    Alignment(pony.AlignmentCenter)
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
<text alignment="leading|center|trailing">...</text>
```

**VStack children:**
```xml
<vstack alignment="leading|center|trailing">...</vstack>
```

**HStack children:**
```xml
<hstack alignment="top|center|bottom">...</hstack>
```

**ZStack children:**
```xml
<zstack alignment="leading|center|trailing" vertical-alignment="top|center|bottom">...</zstack>
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
    return pony.NewBox(
        pony.NewVStack(
            pony.NewText(props.Get("title")).Bold(),
            pony.NewDivider(),
            pony.NewVStack(children...),
        ),
    ).Border("rounded").Padding(1)
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
    themeColor := pony.Hex("#00FFFF")
    card := pony.NewBox(
        pony.NewVStack(
            pony.NewText(c.Title).ForegroundColor(themeColor).Bold(),
            pony.NewDivider(),
            pony.NewVStack(c.Content...),
        ),
    ).Border("rounded").BorderColor(themeColor).Padding(1)

    card.Draw(scr, area)
}

func (c *Card) Layout(constraints pony.Constraints) pony.Size {
    // Delegate to composed structure
    themeColor := pony.Hex("#00FFFF")
    card := pony.NewBox(
        pony.NewVStack(
            pony.NewText(c.Title).ForegroundColor(themeColor).Bold(),
            pony.NewDivider(),
            pony.NewVStack(c.Content...),
        ),
    ).Border("rounded").BorderColor(themeColor).Padding(1)

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
    ).Border("rounded")
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
        pony.NewBox(pony.NewText(i.value)).Border("rounded"),
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
    Border("rounded").
    Padding(2).
    Margin(1).
    MarginTop(2).
    Width(pony.NewFixedConstraint(50)).
    BorderColor(pony.Hex("#00FFFF"))

text := pony.NewText("Hello").
    ForegroundColor(pony.Hex("#FF5555")).
    Bold().
    Italic().
    Alignment(pony.AlignmentCenter).
    Wrap(true)

button := pony.NewButton("Click Me").
    Border("rounded").
    Padding(1).
    Style(style).
    Width(pony.NewFixedConstraint(20))
button.SetID("my-button")

flex := pony.NewFlex(child).
    Grow(1).
    Shrink(0).
    Basis(20)

positioned := pony.NewPositioned(child, 10, 5).
    Right(2).
    Bottom(1)
```

### Style Builder

StyleBuilder is now deprecated. Use granular Text methods instead:

```go
// Old way (deprecated)
style := pony.NewStyle().Fg(...).Bold().Build()
text.Style(style)

// New way (SwiftUI-style)
text := pony.NewText("Hello").
    ForegroundColor(pony.Hex("#FF5555")).
    BackgroundColor(pony.RGB(40, 42, 54)).
    Bold().
    Italic().
    Underline()
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
pony.Card(title, titleColor, borderColor, children...)
pony.Section(header, headerColor, children...)
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
