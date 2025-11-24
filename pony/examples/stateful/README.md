# Stateful Components with Slots

This example demonstrates pony's stateful component system using **slots**.

## How It Works

### The Pattern

1. **Component manages state** - handles Update, stores data
2. **Component renders to Element** - returns pony elements (stateless!)
3. **Slots inject components** - template has `<slot>` placeholders
4. **BT routes events** - updates focused component

### Key Insight

**Components don't need to know their size!**

They return declarative `Element` structures, and pony's layout engine handles all sizing.

## Creating a Stateful Component

```go
type Input struct {
    value   string
    cursor  int
    focused bool
}

// Update handles events (state management)
func (i *Input) Update(msg tea.Msg) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        i.value += msg.String()
        i.cursor++
    }
}

// Render returns pony Element (no size needed!)
func (i *Input) Render() pony.Element {
    style, _ := pony.ParseStyle("fg:white")
    
    return pony.NewBox(
        pony.NewText(i.value).WithStyle(style),
    ).WithBorder("rounded").WithPadding(1)
}
```

## Using in Bubble Tea

### 1. Template with Slots

```xml
<vstack gap="1">
    <text>Username:</text>
    <slot name="username" />
    
    <text>Email:</text>
    <slot name="email" />
</vstack>
```

### 2. Model with Components

```go
type model struct {
    template *pony.Template[ViewData]
    username *Input  // Stateful component
    email    *Input  // Stateful component
    focused  string
}
```

### 3. Update Routes to Components

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        // Route to focused component
        switch m.focused {
        case "username":
            m.username.Update(msg)
        case "email":
            m.email.Update(msg)
        }
    }
    return m, nil
}
```

### 4. View Fills Slots

```go
func (m model) View() tea.View {
    data := ViewData{Title: "My Form"}
    
    slots := map[string]pony.Element{
        "username": m.username.Render(), // Render to Element!
        "email":    m.email.Render(),
    }
    
    output := m.template.RenderWithSlots(data, slots, m.width, m.height)
    return tea.NewView(output)
}
```

## Why This DX Is Great

✅ **Clean separation**: State vs rendering  
✅ **Component encapsulation**: Update + Render in one place  
✅ **No size management**: pony layouts everything  
✅ **Composable**: Components built from pony elements  
✅ **Type-safe**: Slots are `map[string]Element`  
✅ **Explicit**: You see slot names in template  
✅ **Flexible**: Any Element can be slotted  

## Running This Example

```bash
go run main.go
```

### Controls

- `Tab` / `Shift+Tab` - Switch focus between inputs
- `Type` - Enter text in focused input
- `Backspace` / `Delete` - Edit text
- `Enter` - Submit (increments counter)
- `+` / `-` - Increment/decrement counter
- `Ctrl+R` - Reset counter
- `Esc` / `Ctrl+C` - Quit

## What This Demonstrates

1. ✅ Stateful text inputs with cursor
2. ✅ Focus management across components
3. ✅ Event routing to focused component
4. ✅ State preservation across re-renders
5. ✅ Combining template data + slot elements
6. ✅ Components composed from pony primitives
7. ✅ Clean DX - components just implement Update + Render
