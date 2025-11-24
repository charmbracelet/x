# pony Examples

This directory contains working examples demonstrating pony features.

## Running Examples

Each example is a standalone Go module. To run:

```bash
cd <example-name>
go run main.go
```

## Examples

### Basic Examples

**hello/** - Hello World
- Basic pony usage
- Simple layouts
- Demonstrates core elements

**layout/** - Layout Showcase
- Responsive layouts with percentages
- Fixed sizes and auto sizing
- Nested containers
- Different border styles

**styled/** - Styling Showcase
- All color formats (named, hex, RGB, ANSI)
- Text attributes (bold, italic, etc.)
- Underline styles
- Background colors
- Border styling

**alignment/** - Alignment & Padding
- Text alignment (left, center, right)
- Container alignment (VStack align, HStack valign)
- Padding demonstration
- Combined features

### Template Examples

**dynamic/** - Type-Safe Templates
- Generic `Template[T]` with typed data
- Variables, loops, conditionals
- Template functions
- Dynamic styling

### Component Examples

**components/** - Built-in Components
- Badge component for status indicators
- Progress bars with styling
- Header component with underlines
- Combining components

**custom/** - Custom Components
- Creating custom components
- Component registry usage
- Composition from primitives
- Using style helpers

**helpers/** - Style & Layout Helpers
- StyleBuilder for type-safe styling
- Color helpers (Hex, RGB)
- Layout helpers (Panel, Card, Section)
- Comparison: before/after helpers

### Advanced Examples

**stateful/** - Stateful Components
- Slot system for dynamic content
- Stateful text input component
- Focus management
- Event routing
- State preservation across renders

**scrolling/** - Scrollable Views
- ScrollView element
- Stateful scrollable component
- Keyboard and mouse wheel scrolling
- Scrollbar rendering
- Large content in small viewport

### Bubble Tea Integration

**simple-bubbletea/** - Minimal Integration
- Basic Bubble Tea app with pony
- Window resize handling
- Simple counter example
- Clean integration pattern

**bubbletea/** - Full Interactive App
- Complete interactive application
- Multiple stateful components
- Focus management
- Event log
- Toggle panels
- Real-world patterns

## Example Categories

### For Learning

Start with these in order:
1. hello
2. layout
3. styled
4. dynamic

### For Reference

When you need to see how to:
- **Style elements** → styled, helpers
- **Create layouts** → layout, alignment
- **Use templates** → dynamic
- **Build components** → components, custom
- **Handle state** → stateful, scrolling
- **Integrate with Bubble Tea** → simple-bubbletea, bubbletea

## Common Patterns

### Type-Safe Template

```go
type ViewData struct {
    Title string
    Count int
}

tmpl := pony.MustParse[ViewData](markup)
output := tmpl.Render(ViewData{Title: "App", Count: 42}, 80, 24)
```

### Bubble Tea Integration

```go
type model struct {
    template *pony.Template[ViewData]
    width    int
    height   int
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
    }
    return m, nil
}

func (m model) View() tea.View {
    output := m.template.Render(data, m.width, m.height)
    return tea.NewView(output)
}
```

### Custom Component

```go
func NewMyComponent(props Props, children []Element) Element {
    return pony.NewBox(
        pony.NewVStack(children...),
    ).WithBorder("rounded").WithPadding(1)
}

pony.Register("mycomp", NewMyComponent)
```

### Stateful Component

```go
type MyComp struct {
    state string
}

func (c *MyComp) Update(msg tea.Msg) { /* handle events */ }

func (c *MyComp) Render() pony.Element {
    return pony.NewText(c.state)
}

// In template: <slot name="comp" />
// In View: slots["comp"] = m.comp.Render()
```

## Tips

- Use type parameters for template data: `Template[YourType]`
- Always pass terminal size to `Render(data, width, height)`
- Enable mouse for scrolling: `view.MouseMode = tea.MouseModeCellMotion`
- Use helpers instead of string parsing for styles
- Compose custom components from pony primitives
- Keep state in Bubble Tea model, render via slots

## Need Help?

- Check the example code
- Read the main [README.md](../README.md)
- Review component source code in parent directory
