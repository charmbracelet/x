# AGENTS.md - Guide for AI Agents Working on pony

This file documents essential information for AI agents working in the pony codebase.

## Project Overview

**pony** is a declarative, type-safe markup language for building terminal user interfaces. It uses [Ultraviolet](https://github.com/charmbracelet/ultraviolet) as the rendering engine and integrates with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

- **Language**: Go 1.24.2
- **Module**: `github.com/charmbracelet/x/pony`
- **Primary dependencies**: 
  - `github.com/charmbracelet/ultraviolet` (UV rendering)
  - `github.com/charmbracelet/x/ansi` (ANSI parsing)
  - `golang.org/x/text` (text processing)
- **Experimental**: This is primarily AI-generated and experimental

## Essential Commands

All commands use Task (Taskfile.yaml) or standard Go tooling.

### Testing
```bash
# Run all tests
task test
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestRender

# Run tests with coverage
task test:coverage

# Update golden test files (IMPORTANT for rendering tests)
task test:update
go test -update ./...
```

### Linting & Formatting
```bash
# Run linter (uses .golangci.yml config)
task lint

# Run linter with auto-fix
task lint:fix

# Format code (gofumpt + goimports)
task fmt

# Install golangci-lint
task lint:install
```

### Maintenance
```bash
# Clean build artifacts and test cache
task clean

# Tidy dependencies
task tidy
go mod tidy

# List all available tasks
task
```

## Code Organization

### Main Source Files (Root)

Core element types and primitives:
- `element.go` - Element interface, Constraints, Size types
- `box.go` - Box container with borders, padding, margin
- `text.go` - Text element with styling and alignment
- `container.go` - VStack, HStack container implementations
- `zstack.go` - ZStack (layering/overlay)
- `flex.go` - Flex wrapper for flexible sizing
- `positioned.go` - Positioned element (absolute positioning)
- `spacer.go` - Spacer element (fixed/flexible)
- `divider.go` - Divider element
- `scrollview.go` - ScrollView with scrollbar support
- `button.go` - Button component with click handling

Template and parsing:
- `template.go` - Template[T] type with Go template integration
- `parser.go` - XML parser (markup → element tree)
- `slot.go` - Slot system for dynamic content injection

Styling and layout:
- `style.go` - Style parsing and builder API
- `layout.go` - Size constraint helpers
- `constants.go` - Border, alignment, unit constants

Component system:
- `registry.go` - Global component registry
- `components.go` - Built-in components (Badge, Progress, Header)

Interactivity:
- `bounds.go` - BoundsMap, hit testing, BaseElement
- `doc_interactive.go` - Mouse interaction documentation

Helpers:
- `helpers.go` - Layout and style helper functions

### Test Files

Tests follow `*_test.go` naming convention. Key test files:
- `element_test.go` - Element layout and rendering
- `parser_test.go` - XML parsing
- `style_test.go` - Style parsing and rendering
- `template_test.go` - Go template integration
- `layout_test.go` - Size constraints
- `alignment_test.go` - Alignment
- `bounds_test.go`, `bounds_hitall_test.go` - Hit testing
- `scrollview_test.go`, `scrollview_props_test.go`, `scrollview_scrollbar_test.go` - Scrolling
- `slot_test.go`, `slots_bounds_test.go` - Slot system
- `input_click_test.go`, `nested_interactive_test.go` - Mouse interactions

### Test Data

- `testdata/` - Golden files for rendering tests (`.golden` extension)
- Golden files are named after test functions: `TestMyFunc.golden` or `TestMyFunc/subtest.golden`

### Examples

- `examples/` - Complete working examples
- Each example has its own `go.mod` and `main.go`
- Examples demonstrate features: hello, layout, styled, dynamic, components, custom, stateful, scrolling, bubbletea, buttons, etc.

## Code Patterns & Conventions

### Element Pattern

All elements implement the `Element` interface:
```go
type Element interface {
    uv.Drawable                          // Draw(scr uv.Screen, area uv.Rectangle)
    Layout(constraints Constraints) Size // Calculate size
    Children() []Element                 // Return child elements (or nil)
    ID() string                          // Element identifier
    SetID(id string)                     // Set identifier
    Bounds() uv.Rectangle                // Last rendered bounds
    SetBounds(bounds uv.Rectangle)       // Set bounds (call in Draw)
}
```

### BaseElement Embedding

**All element types embed `BaseElement`** to get ID and bounds tracking:
```go
type MyElement struct {
    BaseElement  // Required for ID and bounds
    // ... other fields
}
```

`BaseElement` provides:
- `ID()` - Returns explicit ID or pointer-based ID (`elem_%p`)
- `SetID(string)` - Set explicit ID
- `Bounds()` - Get last rendered bounds
- `SetBounds(uv.Rectangle)` - Record bounds (call at start of `Draw()`)

### Fluent API Pattern

**All elements use method chaining** for configuration:
```go
box := NewBox(child).
    WithBorder("rounded").
    WithPadding(2).
    WithMargin(1)

text := NewText("Hello").
    WithStyle(style).
    WithAlign("center")
```

Method naming: `With<Property>(*Element) *Element`

### Draw() Implementation

**Always call `SetBounds()` first** in `Draw()`:
```go
func (e *MyElement) Draw(scr uv.Screen, area uv.Rectangle) {
    e.SetBounds(area)  // REQUIRED for hit testing
    // ... rest of drawing logic
}
```

### Custom Components

Two approaches:

**1. Functional component** (simple):
```go
pony.Register("card", func(props pony.Props, children []pony.Element) pony.Element {
    return pony.NewBox(
        pony.NewVStack(children...),
    ).WithBorder("rounded").WithPadding(1)
})
```

**2. Type-based component** (more control):
```go
type Card struct {
    BaseElement  // Required
    Title   string
    Content []Element
}

func NewCard(props Props, children []Element) Element {
    return &Card{
        Title:   props.Get("title"),
        Content: children,
    }
}

func (c *Card) Draw(scr uv.Screen, area uv.Rectangle) {
    c.SetBounds(area)  // Required
    // Build composed structure and draw
    card := NewBox(NewVStack(c.Content...)).WithBorder("rounded")
    card.Draw(scr, area)
}

func (c *Card) Layout(constraints Constraints) Size {
    // Delegate to composed structure
    card := NewBox(NewVStack(c.Content...)).WithBorder("rounded")
    return card.Layout(constraints)
}

func (c *Card) Children() []Element {
    return c.Content
}

// Register
pony.Register("card", NewCard)
```

### Props Helper Type

`Props` is a `map[string]string` with helpers:
```go
type Props map[string]string

props.Get("key")           // Get value (empty if missing)
props.GetOr("key", "def")  // Get with default
props.GetInt("key", 10)    // Parse int with default
props.GetBool("key")       // Parse bool (default false)
```

### Style Parsing

Styles can be:
1. **Parsed from string** (markup): `"fg:red; bg:blue; bold"`
2. **Built with API** (code):
```go
style := pony.NewStyle().
    Fg(pony.Hex("#FF5555")).
    Bg(pony.RGB(40, 42, 54)).
    Bold().
    Italic().
    Build()
```

Style string format:
- Colors: `fg:<color>`, `bg:<color>`
- Color formats: named (`red`), hex (`#FF5555`), rgb (`rgb(255,85,85)`), ansi (`196`)
- Attributes: `bold`, `italic`, `underline`, `strikethrough`, `faint`, `blink`, `reverse`
- Underline styles: `underline:single|double|curly|dotted|dashed`

### Size Constraints

Size constraints use special types:
- `FixedConstraint` - Fixed size in cells
- `PercentConstraint` - Percentage of available space (0-100)
- `AutoConstraint` - Content-based size (default)
- `MinConstraint` - Minimum content size
- `MaxConstraint` - Maximum available space

Parse from string with `parseSizeConstraint(value)` (handles `"50%"`, `"20"`, `"auto"`, etc.)

### Constants

Use constants from `constants.go`:
- Borders: `BorderNone`, `BorderRounded`, `BorderNormal`, `BorderThick`, `BorderDouble`, `BorderHidden`
- Alignment: `AlignLeft`, `AlignCenter`, `AlignRight`, `AlignTop`, `AlignMiddle`, `AlignBottom`
- Units: `UnitAuto`, `UnitMin`, `UnitMax`, `UnitPercent`

## Testing Approach

### Golden File Testing

**Most rendering tests use golden files** stored in `testdata/`:
```go
func TestMyRender(t *testing.T) {
    output := tmpl.Render(data, 80, 24)
    golden.RequireEqual(t, output)  // Compares against testdata/TestMyRender.golden
}
```

**CRITICAL**: When rendering output changes intentionally:
```bash
go test -update ./...  # Regenerate all golden files
```

Golden files preserve:
- Exact ANSI escape codes
- Terminal formatting
- Spacing and alignment

**Never manually edit golden files** - always regenerate with `-update`.

### Test Organization

- Use table-driven tests for multiple cases
- Use subtests: `t.Run(name, func(t *testing.T) { ... })`
- Test both logic (Layout) and rendering (Draw/golden)

Example:
```go
func TestElement(t *testing.T) {
    tests := []struct {
        name string
        // ... fields
    }{
        {name: "case1", ...},
        {name: "case2", ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Important Gotchas

### 1. Golden Files Are Critical

**Always run `go test -update` after changing rendering logic.** Failing to do so will cause tests to fail. Golden files are the source of truth for expected output.

### 2. SetBounds() Must Be First

**Always call `SetBounds(area)` at the start of `Draw()`** for mouse hit testing to work. Forgetting this breaks mouse interactivity.

### 3. BaseElement Is Required

**All custom elements must embed `BaseElement`** to satisfy the Element interface and enable ID/bounds tracking.

### 4. Component Registry Is Global

`Register()` modifies global state. In tests that register components:
- Register in `init()` or test setup
- Consider cleanup with `Unregister()` in test teardown
- Or accept that components are globally available after registration

### 5. Template Type Safety

Templates are type-safe via generics:
```go
type ViewData struct { ... }
tmpl := pony.Parse[ViewData](markup)  // Type-checked at compile time
tmpl.Render(ViewData{...}, w, h)       // Only accepts ViewData
```

Use `interface{}` or `any` for templates without data: `Parse[any](markup)`

### 6. Markup Is XML

pony uses XML parsing under the hood:
- Self-closing tags: `<divider />` not `<divider>`
- Attributes in quotes: `width="50%"` not `width=50%`
- Special characters must be escaped: `&lt;`, `&gt;`, `&amp;`
- Go template syntax is processed before XML parsing

### 7. Mouse Interaction Requires Bubble Tea PR

Mouse handling requires Bubble Tea PR #1549 (View.Callback support). Pin to specific commit:
```go
require (
    charm.land/bubbletea/v2 v2.0.0-20250120210912-18cfb8c3ccb3
)
```

### 8. HitTest() Prefers Explicit IDs

`BoundsMap.HitTest(x, y)` prefers elements with explicit IDs over auto-generated IDs. For custom components, **set the ID on the root element** you return:
```go
func (c *MyComponent) Render() Element {
    vstack := NewVStack(...)
    vstack.SetID(c.ID())  // Pass through component ID
    return vstack
}
```

### 9. UV Screen Coordinates

Ultraviolet uses `uv.Rectangle` for screen areas:
- `area.Min.X`, `area.Min.Y` - Top-left corner
- `area.Max.X`, `area.Max.Y` - Bottom-right corner (exclusive)
- `area.Dx()` - Width
- `area.Dy()` - Height

### 10. Style Is UV Style

`uv.Style` from Ultraviolet is used throughout. It's **not** a string - use `NewStyle()` builder or `parseStyle()` to create.

### 11. Children() Returns Nil for Leaves

Leaf elements (Text, Spacer, Divider) return `nil` from `Children()`, not an empty slice.

### 12. Constraints Are Immutable

`Constraints` and `Size` are value types. Methods like `Constrain()` return new values - they don't modify in place.

## Working with Rendering

### UV Screen Buffer

Templates render to `uv.ScreenBuffer`:
```go
scr, boundsMap := tmpl.RenderWithBounds(data, slots, width, height)
output := scr.Render()  // Convert to string
```

### BoundsMap for Mouse Events

`BoundsMap` tracks element positions for hit testing:
```go
elem := boundsMap.HitTest(x, y)  // Find element at coordinates
if elem != nil {
    id := elem.ID()  // Get element ID for event handling
}
```

### Slots for Stateful Components

Slots allow dynamic content injection:
```xml
<slot name="input" />
```

```go
slots := map[string]Element{
    "input": myInputComponent.Render(),
}
output := tmpl.RenderWithSlots(data, slots, w, h)
```

## File Naming Conventions

- Source files: `<element>.go` (e.g., `box.go`, `text.go`)
- Test files: `<element>_test.go` (e.g., `box_test.go`)
- Multi-word: lowercase with underscores (e.g., `scroll_view.go` → NO, `scrollview.go` → YES)
- Constants: `constants.go`
- Helpers: `helpers.go`
- Documentation: `doc.go`, `doc_<topic>.go`

## Documentation

- Package docs in `doc.go`
- Interactive docs in `doc_interactive.go`
- User guide in `README.md`
- Testing guide in `TESTING.md`
- This file: `AGENTS.md`

## CI/CD

GitHub Actions workflows:
- `.github/workflows/build.yml` - Run tests and build
- `.github/workflows/lint.yml` - Run golangci-lint

Both use Charmbracelet's reusable workflows from `charmbracelet/meta`.

## Linting Configuration

`.golangci.yml` enables:
- bodyclose, exhaustive, goconst, godot, godox, gomoddirectives, goprintffuncname, gosec
- misspell, nakedret, nilerr, noctx, nolintlint, prealloc, revive
- rowserrcheck, sqlclosecheck, staticcheck, tparallel, unconvert, unparam
- whitespace, wrapcheck

Formatters:
- gofumpt (stricter gofmt)
- goimports (auto-import management)

## Common Tasks

### Adding a New Element Type

1. Create `<element>.go` with struct embedding `BaseElement`
2. Implement Element interface: `Draw()`, `Layout()`, `Children()` (from BaseElement: ID, SetID, Bounds, SetBounds)
3. Add constructor: `New<Element>()`
4. Add fluent API methods: `With<Property>(*Element) *Element`
5. Add to parser in `parser.go`: case in `toElement()` switch
6. Add tests in `<element>_test.go`
7. Update golden files: `go test -update`
8. Add example in `examples/`

### Adding a Built-in Component

1. Add constructor in `components.go`: `func NewBadge(props Props) Element`
2. Or register functional: `Register("badge", func(props Props, children []Element) Element { ... })`
3. Add to `parser.go` if it should be available in markup
4. Add tests in `registry_test.go` or `components_test.go`

### Changing Rendering Logic

1. Make changes to `Draw()` or `Layout()`
2. Run tests: `task test` - they will fail
3. Review failures to ensure changes are intentional
4. Update golden files: `task test:update`
5. Verify golden file diffs in git: `git diff testdata/`
6. Run tests again: `task test` - should pass

### Adding Template Functions

Add to `defaultTemplateFuncs()` in `template.go`:
```go
func defaultTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "upper": strings.ToUpper,
        "myFunc": func(arg string) string { ... },
    }
}
```

## Dependencies

Main dependencies are managed in `go.mod`:
- Ultraviolet (UV) - Core rendering engine
- Charmbracelet packages (ansi, term, golden testing)
- golang.org/x/text - Text processing

Examples have separate `go.mod` files that depend on the main module.

## Project-Specific Context

- **Experimental**: Code is primarily AI-generated, treat carefully
- **UV-centric**: Rendering heavily depends on Ultraviolet primitives
- **Type-safe**: Heavy use of Go generics for compile-time safety
- **Declarative**: XML-based markup with Go template preprocessing
- **Stateless rendering**: View functions are pure (no model mutation)
- **Mouse handling**: Uses unreleased Bubble Tea feature (specific commit required)

## Questions/Debugging

When debugging:
1. Check golden files in `testdata/` for expected output
2. Run single test: `go test -run TestName -v`
3. Update golden files if needed: `go test -run TestName -update`
4. Check examples in `examples/` for working patterns
5. Review `README.md` for API documentation

When adding features:
1. Look for similar existing elements/components
2. Follow established patterns (BaseElement, fluent API, etc.)
3. Add comprehensive tests
4. Update golden files
5. Add example if it's a major feature

## Summary

Key things to remember:
- ✅ Always embed `BaseElement` in custom elements
- ✅ Always call `SetBounds()` first in `Draw()`
- ✅ Update golden files after changing rendering: `go test -update`
- ✅ Use fluent API pattern: `With<Property>() *Element`
- ✅ Use constants from `constants.go`
- ✅ Write tests for both logic and rendering
- ✅ Check examples for working patterns
