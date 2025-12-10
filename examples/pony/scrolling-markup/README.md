# Scrolling with Markup Demo

This example demonstrates how to build interactive scroll views using **mostly markup** with pony.

## Key Features

### 1. **Custom Components in Markup**

Define a custom `ListItem` component in Go and register it:

```go
type ListItem struct {
    pony.BaseElement
    Text     string
    Selected bool
}

func NewListItemFromProps(props pony.Props, children []pony.Element) pony.Element {
    text := props.Get("text")
    selected := props.Get("selected") == "true"
    return NewListItem(text, selected)
}

func init() {
    pony.Register("listitem", NewListItemFromProps)
}
```

Then use it in markup:

```xml
<listitem id="item-1" text="My Item" selected="true" />
```

### 2. **Scroll View with Props**

The scroll view now accepts props including `offset-y` for dynamic scrolling:

```xml
<scrollview id="main-scroll-view" height="12" offset-y="{{ .ScrollOffset }}">
    <vstack gap="0">
        {{ range .Items }}
        <listitem id="item-{{ .ID }}" text="{{ .Text }}" selected="{{ .Selected }}" />
        {{ end }}
    </vstack>
</scrollview>
```

### 3. **HitTestAll for Nested Clicks**

Use `HitTestAll()` to detect clicks on items inside scroll views:

```go
hits := boundsMap.HitTestAll(mouse.X, mouse.Y)

for _, elem := range hits {
    id := elem.ID()
    if strings.HasPrefix(id, "item-") {
        // Handle item click
    }
}
```

## Template Structure

The entire UI is defined in markup:

```xml
<vstack gap="1">
    <!-- Header box -->
    <box border="double" border-style="fg:cyan; bold" padding="1">
        <text>Title</text>
    </box>

    <!-- Scrollable content -->
    <scrollview height="12" offset-y="{{ .ScrollOffset }}">
        <vstack>
            {{ range .Items }}
            <listitem id="item-{{ .ID }}" text="{{ .Text }}" />
            {{ end }}
        </vstack>
    </scrollview>

    <!-- Instructions -->
    <text>Press q to quit</text>
</vstack>
```

## Data Flow

1. **Model holds state**: scroll offset, selected item, items list
2. **Template receives data**: `TemplateData` with all state
3. **Markup renders UI**: components are created from template
4. **HitTestAll handles clicks**: finds which item was clicked
5. **Messages update model**: selection and scroll offset change

## Running

```bash
go run main.go
```

## Controls

- **Click** items to select them
- **Mouse Wheel** to scroll
- **q** to quit

## What's Different from scrolling-with-clicks?

- **Markup-first**: UI is defined in templates, not Go code
- **Custom component**: `ListItem` is registered and used in markup
- **Props for state**: `offsety` prop controls scroll position
- **Cleaner code**: Less programmatic UI building, more declarative
