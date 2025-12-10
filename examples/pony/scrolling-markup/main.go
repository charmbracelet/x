package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
	uv "github.com/charmbracelet/ultraviolet"
)

// ListItem is a custom component that can be registered and used in markup
type ListItem struct {
	pony.BaseElement
	Text     string
	Selected bool
}

func NewListItem(text string, selected bool) *ListItem {
	return &ListItem{Text: text, Selected: selected}
}

// Render implements the Component interface
func (l *ListItem) Render() pony.Element {
	text := pony.NewText(l.Text)

	if l.Selected {
		text = text.BackgroundColor(pony.RGB(0, 0, 255)).
			ForegroundColor(pony.RGB(255, 255, 255)).
			Bold()
	}

	box := pony.NewBox(text).
		Padding(1).
		Border("rounded")

	if l.Selected {
		box = box.BorderColor(pony.RGB(0, 0, 255))
	}

	// Pass through ID
	box.SetID(l.ID())

	return box
}

func (l *ListItem) Layout(constraints pony.Constraints) pony.Size {
	return l.Render().Layout(constraints)
}

func (l *ListItem) Draw(scr uv.Screen, area uv.Rectangle) {
	l.SetBounds(area)
	l.Render().Draw(scr, area)
}

func (l *ListItem) Children() []pony.Element {
	return nil
}

// Factory function for parser
func NewListItemFromProps(props pony.Props, children []pony.Element) pony.Element {
	text := props.Get("text")
	selected := props.Get("selected") == "true"
	return NewListItem(text, selected)
}

func init() {
	// Register the ListItem component
	pony.Register("listitem", NewListItemFromProps)
}

// Template with mostly markup
const tmpl = `
<vstack spacing="1">
	<!-- Header -->
	<box border="double" border-color="cyan" padding="1">
		<text font-weight="bold" foreground-color="yellow" alignment="center">Scroll View with Clickable Items (Markup Demo)</text>
	</box>

	<divider foreground-color="gray" />

	<!-- Info display -->
	<vstack spacing="0">
		<text font-weight="bold">Selected: {{ .SelectedItem }}</text>
		<text font-style="italic" foreground-color="gray" width="80">{{ .HitInfo }}</text>
	</vstack>

	<divider foreground-color="gray" />

	<!-- Scroll view with items (entirely in markup!) -->
	<box border="rounded" border-color="green">
		<scrollview id="main-scroll-view" height="12" offset-y="{{ .ScrollOffset }}">
			<vstack spacing="0">
				{{ range .Items }}
				<listitem id="item-{{ .ID }}" text="{{ .Text }}" selected="{{ .Selected }}" />
				{{ end }}
			</vstack>
		</scrollview>
	</box>

	<divider foreground-color="gray" />

	<!-- Instructions -->
	<text font-style="italic" foreground-color="gray">Click items to select • Mouse wheel to scroll • q to quit</text>
</vstack>
`

// Data structures
type ItemData struct {
	ID       int
	Text     string
	Selected bool
}

type TemplateData struct {
	SelectedItem string
	HitInfo      string
	ScrollOffset int
	Items        []ItemData
}

// Model
type model struct {
	template     *pony.Template[TemplateData]
	items        []ItemData
	scrollOffset int
	selectedItem string
	hitInfo      string
	width        int
	height       int
}

func initialModel() model {
	var items []ItemData
	for i := 1; i <= 20; i++ {
		items = append(items, ItemData{
			ID:       i,
			Text:     fmt.Sprintf("List Item %d - Click me!", i),
			Selected: false,
		})
	}

	return model{
		template:     pony.MustParse[TemplateData](tmpl),
		items:        items,
		selectedItem: "none",
		width:        80,
		height:       24,
	}
}

func (m model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

// Messages
type selectItemMsg int
type hitInfoMsg string
type scrollOffsetMsg int

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case selectItemMsg:
		itemID := int(msg)
		m.selectedItem = fmt.Sprintf("item-%d", itemID)

		// Clear selection from all items
		for i := range m.items {
			m.items[i].Selected = false
		}

		// Select the clicked item
		if itemID > 0 && itemID <= len(m.items) {
			m.items[itemID-1].Selected = true
		}

	case hitInfoMsg:
		m.hitInfo = string(msg)

	case scrollOffsetMsg:
		m.scrollOffset = int(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	data := TemplateData{
		SelectedItem: m.selectedItem,
		HitInfo:      m.hitInfo,
		ScrollOffset: m.scrollOffset,
		Items:        m.items,
	}

	// Render with bounds
	scr, boundsMap := m.template.RenderWithBounds(data, nil, m.width, m.height)

	view := tea.NewView(scr.Render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeAllMotion

	// Set up callback using HitTestAll
	view.Callback = func(msg tea.Msg) tea.Cmd {
		switch msg := msg.(type) {
		case tea.MouseWheelMsg:
			// Get the ScrollView to calculate actual content size
			if sv, ok := boundsMap.GetByID("main-scroll-view"); ok {
				if scrollView, ok := sv.(*pony.ScrollView); ok {
					// Get actual content size from ScrollView
					contentSize := scrollView.ContentSize()
					viewportHeight := 12 // Height specified in template
					maxOffset := max(0, contentSize.Height-viewportHeight)

					mouse := msg.Mouse()
					newOffset := m.scrollOffset

					switch mouse.Button {
					case tea.MouseWheelUp:
						newOffset = max(0, m.scrollOffset-3)
					case tea.MouseWheelDown:
						newOffset = min(maxOffset, m.scrollOffset+3)
					}

					if newOffset != m.scrollOffset {
						return func() tea.Msg {
							return scrollOffsetMsg(newOffset)
						}
					}
				}
			}

		case tea.MouseClickMsg:
			mouse := msg.Mouse()

			// Use HitTestAll to get all elements at click position
			hits := boundsMap.HitTestAll(mouse.X, mouse.Y)

			if len(hits) == 0 {
				return nil
			}

			// Log what we hit for debugging
			var hitIDs []string
			for _, elem := range hits {
				hitIDs = append(hitIDs, elem.ID())
			}
			hitInfo := fmt.Sprintf("Hit %d elements: %v", len(hits), hitIDs)

			// Check if we clicked a list item
			for _, elem := range hits {
				id := elem.ID()
				// Parse item-N format
				var itemNum int
				if n, err := fmt.Sscanf(id, "item-%d", &itemNum); err == nil && n == 1 {
					// Found a list item! Return batch of commands
					return tea.Batch(
						func() tea.Msg {
							return hitInfoMsg(hitInfo)
						},
						func() tea.Msg {
							return selectItemMsg(itemNum)
						},
					)
				}
			}

			// If no item found, just update hit info
			return func() tea.Msg {
				return hitInfoMsg(hitInfo)
			}
		}
		return nil
	}

	return view
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nThanks for trying the markup-based scroll view demo!")
}
