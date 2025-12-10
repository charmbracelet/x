# pony + Bubble Tea Integration Example

This example demonstrates how to use pony with Bubble Tea v2 for building interactive terminal applications.

## How It Works

pony handles **view rendering** using declarative markup, while Bubble Tea manages:
- Application lifecycle (Init/Update/View pattern)
- Event handling (keyboard, mouse, window resize)
- Commands and side effects

## Key Integration Points

### 1. Store Template in Model

```go
type model struct {
    template *pony.Template
    // ... your app state
}

func initialModel() model {
    return model{
        template: pony.MustParse(markup),
    }
}
```

### 2. Render with Data in View()

```go
func (m model) View() tea.View {
    // Prepare data for template
    data := map[string]interface{}{
        "Count": m.count,
        "Items": m.items,
    }
    
    // Render pony with window size
    output := m.template.RenderToSize(data, m.width, m.height)
    
    // Return as Bubble Tea View
    return tea.NewView(output)
}
```

### 3. Handle Window Resize

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

### 4. Request Initial Size

```go
func (m model) Init() tea.Cmd {
    return tea.RequestWindowSize
}
```

## This Example

The demo shows:
- **Dynamic counter** that increments with spacebar
- **Live timer** updating every second
- **Event log** showing recent actions
- **Toggle help** with 'h' key
- **Conditional rendering** (help panel, celebration message)
- **Window size tracking** in the UI

## Running

```bash
go run main.go
```

### Controls

- `Space` - Increment counter
- `r` - Reset counter
- `h` - Toggle help
- `q` or `Ctrl+C` - Quit

## Benefits

✅ **Declarative UI** - Define layout in markup, not code  
✅ **Type-safe** - Go templates with compile-time checking  
✅ **Reactive** - Update data, UI updates automatically  
✅ **Styled** - Full color and styling support  
✅ **BubbleTea Compatible** - Works seamlessly with BT's event system
