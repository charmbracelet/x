# Scrolling Example

This example demonstrates pony's scrolling system with the `<scrollview>` element and stateful scrollable components.

## Two Ways to Scroll

### 1. Markup-Based (Static)

Use `<scrollview>` directly in markup:

```xml
<scrollview height="10">
    <vstack>
        <text>Line 1</text>
        <text>Line 2</text>
        <!-- ... many lines ... -->
    </vstack>
</scrollview>
```

**Pros**: Simple, declarative  
**Cons**: Static offset (set via `offset-y` attribute)

### 2. Stateful Component (Interactive)

Create a stateful component that manages scroll position:

```go
type ScrollableLog struct {
    content []string
    offset  int
    height  int
}

func (s *ScrollableLog) Update(msg tea.Msg) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if msg.String() == "down" {
            s.offset++
        }
    }
}

func (s *ScrollableLog) Render() pony.Element {
    items := /* build elements from content */
    
    return &pony.ScrollView{
        Child:   pony.NewVStack(items...),
        OffsetY: s.offset,  // Dynamic offset!
        Height:  pony.NewFixedConstraint(s.height),
    }
}

// In template
<slot name="log" />

// In View
slots["log"] = m.log.Render()
```

**Pros**: Interactive, dynamic scrolling  
**Cons**: Requires state management

## ScrollView Features

### Attributes

- `height="10"` - Viewport height (required for vertical scroll)
- `width="40"` - Viewport width (required for horizontal scroll)
- `offset-y="5"` - Initial vertical scroll position
- `offset-x="0"` - Initial horizontal scroll position
- `scrollbar="true"` - Show/hide scrollbar (default: true)
- `vertical="true"` - Enable vertical scrolling (default: true)
- `horizontal="false"` - Enable horizontal scrolling
- `scrollbar-style="fg:cyan"` - Style the scrollbar

### Programmatic API

```go
scroll := pony.NewScrollView(content).
    WithHeight(pony.NewFixedConstraint(20)).
    WithScrollbar(true).
    WithOffset(0, 10)

// Scroll methods
scroll.ScrollUp(1)
scroll.ScrollDown(5, contentHeight, viewportHeight)
scroll.ScrollLeft(3)
scroll.ScrollRight(10, contentWidth, viewportWidth)

// Get content size
size := scroll.ContentSize()
```

## Mouse Support

Enable mouse in Bubble Tea:

```go
func (m model) View() tea.View {
    output := m.template.Render(data, m.width, m.height)
    
    view := tea.NewView(output)
    view.MouseMode = tea.MouseModeCellMotion  // Enable mouse!
    
    return view
}
```

Then handle mouse events:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseWheelMsg:
        m := msg.Mouse()
        if m.Button == tea.MouseWheelUp {
            m.scrollable.ScrollUp(3)
        } else if m.Button == tea.MouseWheelDown {
            m.scrollable.ScrollDown(3)
        }
    }
}
```

## Running This Example

```bash
go run main.go
```

### Controls

**Keyboard:**
- `↑` or `k` - Scroll up
- `↓` or `j` - Scroll down
- `PgUp` - Page up
- `PgDn` - Page down
- `Home` or `g` - Go to top
- `End` or `G` - Go to bottom
- `q`, `Esc`, or `Ctrl+C` - Quit

**Mouse:**
- Scroll wheel to scroll up/down

## What This Demonstrates

✅ **Stateful scrolling** - Interactive scroll with keyboard/mouse  
✅ **ScrollView element** - Viewport with automatic clipping  
✅ **Scrollbars** - Visual indication of position  
✅ **Efficient rendering** - Only visible content rendered  
✅ **Slots integration** - Stateful scroll components via slots  
✅ **Mouse support** - Scroll wheel events  

## Key Insight

Components don't manage viewport - they just set offset:

```go
func (s *ScrollableLog) Render() pony.Element {
    return &pony.ScrollView{
        Child:   content,
        OffsetY: s.offset,  // Component controls offset
        Height:  fixedHeight,  // pony controls viewport
    }
}
```

pony handles:
- Viewport clipping
- Scrollbar rendering
- Content layout
- Size calculations

Component handles:
- Offset changes
- Event handling
- State management
