// Package pony provides a declarative, type-safe markup language for building
// terminal user interfaces using Ultraviolet as the rendering engine.
//
// ⚠️ EXPERIMENTAL: This is an experimental project, primarily AI-generated as
// an exploration of declarative TUI frameworks. Use at your own risk.
//
// pony allows you to define TUI layouts using familiar XML-like markup syntax
// combined with Go templates for dynamic content. It integrates seamlessly with
// Bubble Tea for application lifecycle management while leveraging Ultraviolet's
// efficient cell-based rendering.
//
// # Basic Example
//
//	type ViewData struct {
//	    Title   string
//	    Content string
//	}
//
//	const tmpl = `
//	<vstack gap="1">
//	  <box border="rounded">
//	    <text style="bold; fg:cyan">{{ .Title }}</text>
//	  </box>
//	  <text>{{ .Content }}</text>
//	</vstack>
//	`
//
//	t := pony.MustParse[ViewData](tmpl)
//	data := ViewData{
//	    Title:   "Hello World",
//	    Content: "Welcome to pony!",
//	}
//	output := t.Render(data, 80, 24)
//
// # Elements
//
//   - vstack: Vertical stack container with gap and alignment
//   - hstack: Horizontal stack container with gap and alignment
//   - text: Text content with styling and alignment
//   - box: Container with borders and padding
//   - scrollview: Scrollable viewport with scrollbars
//   - divider: Horizontal or vertical separator line
//   - spacer: Flexible or fixed empty space
//   - slot: Placeholder for dynamic content
//
// # Styling
//
// Elements support CSS-like inline styling:
//
//	<text style="fg:cyan; bg:#1a1b26; bold; italic">Styled text</text>
//
// For programmatic styling, use the StyleBuilder:
//
//	style := pony.NewStyle().
//	    Fg(pony.Hex("#FF5555")).
//	    Bold().
//	    Build()
//
// # Custom Components
//
// Register custom components with the component registry:
//
//	pony.Register("card", func(props Props, children []Element) Element {
//	    return pony.NewBox(
//	        pony.NewVStack(children...),
//	    ).WithBorder("rounded").WithPadding(1)
//	})
//
// Use in markup:
//
//	<card><text>Content</text></card>
//
// # Stateful Components
//
// Use slots for stateful components that manage their own state:
//
//	type Input struct {
//	    value string
//	}
//
//	func (i *Input) Update(msg tea.Msg) { /* handle events */ }
//
//	func (i *Input) Render() pony.Element {
//	    return pony.NewBox(pony.NewText(i.value)).WithBorder("rounded")
//	}
//
// Template with slot:
//
//	<vstack>
//	  <text>Enter name:</text>
//	  <slot name="input" />
//	</vstack>
//
// Render with slots:
//
//	slots := map[string]pony.Element{
//	    "input": m.inputComp.Render(),
//	}
//	output := tmpl.RenderWithSlots(data, slots, width, height)
//
// # Bubble Tea Integration
//
//	type model struct {
//	    template *pony.Template[ViewData]
//	    width    int
//	    height   int
//	}
//
//	func (m model) Init() tea.Cmd {
//	    return tea.RequestWindowSize
//	}
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.WindowSizeMsg:
//	        m.width = msg.Width
//	        m.height = msg.Height
//	    }
//	    return m, nil
//	}
//
//	func (m model) View() tea.View {
//	    data := ViewData{...}
//	    output := m.template.Render(data, m.width, m.height)
//	    return tea.NewView(output)
//	}
package pony
