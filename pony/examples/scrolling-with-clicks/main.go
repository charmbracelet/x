package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// Clickable list item
type ListItem struct {
	pony.BaseElement
	id       string
	text     string
	selected bool
}

func NewListItem(id, text string) *ListItem {
	return &ListItem{id: id, text: text}
}

func (l *ListItem) Render() pony.Element {
	text := pony.NewText(l.text)
	
	if l.selected {
		if style, err := pony.ParseStyle("bg:blue; fg:white; bold"); err == nil {
			text.Style = style
		}
	}

	box := pony.NewBox(text).
		WithPadding(1).
		WithBorder("rounded")

	if l.selected {
		if style, err := pony.ParseStyle("fg:blue; bold"); err == nil {
			box.BorderStyle = style
		}
	}

	// Set the component ID on the root element
	box.SetID(l.id)

	return box
}

// Template
const tmpl = `
<vstack gap="1">
	<box border="double" border-style="fg:cyan; bold" padding="1">
		<text style="bold; fg:yellow" align="center">Scroll View with Clickable Items</text>
	</box>

	<divider style="fg:gray" />

	<text style="fg:gray">Selected: {{ .SelectedItem }}</text>
	<text style="fg:gray; italic">{{ .HitInfo }}</text>

	<divider style="fg:gray" />

	<box border="rounded" border-style="fg:green">
		<slot name="scrollview" />
	</box>

	<divider style="fg:gray" />

	<text style="fg:gray; italic">Click items to select • Mouse wheel to scroll • q to quit</text>
</vstack>
`

type TemplateData struct {
	SelectedItem string
	HitInfo      string
}

type model struct {
	template     *pony.Template[TemplateData]
	items        []*ListItem
	scrollView   *pony.ScrollView
	scrollOffset int
	selectedItem string
	hitInfo      string
	width        int
	height       int
}

func initialModel() model {
	var items []*ListItem
	for i := 1; i <= 20; i++ {
		id := fmt.Sprintf("item-%d", i)
		text := fmt.Sprintf("List Item %d - Click me!", i)
		items = append(items, NewListItem(id, text))
	}

	// Build list of items for scroll view
	var itemElements []pony.Element
	for _, item := range items {
		itemElements = append(itemElements, item.Render())
	}

	// Create scroll view
	scrollView := pony.NewScrollView(pony.NewVStack(itemElements...))
	scrollView.SetID("main-scroll-view")
	scrollView.WithHeight(pony.NewFixedConstraint(12))
	scrollView.WithVertical(true)
	scrollView.WithScrollbar(true)

	return model{
		template:     pony.MustParse[TemplateData](tmpl),
		items:        items,
		scrollView:   scrollView,
		selectedItem: "none",
		width:        80,
		height:       24,
	}
}

func (m model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

type selectItemMsg string
type hitInfoMsg string

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.MouseWheelMsg:
		mouse := msg.Mouse()
		contentSize := m.scrollView.ContentSize()
		viewportHeight := 12 // Fixed height from WithHeight
		maxOffset := max(0, contentSize.Height-viewportHeight)
		
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.scrollOffset = max(0, m.scrollOffset-3)
			m.scrollView.OffsetY = m.scrollOffset
		case tea.MouseWheelDown:
			m.scrollOffset = min(maxOffset, m.scrollOffset+3)
			m.scrollView.OffsetY = m.scrollOffset
		}

	case selectItemMsg:
		m.selectedItem = string(msg)
		// Clear selection from all items
		for _, item := range m.items {
			item.selected = false
		}
		// Select the clicked item
		for _, item := range m.items {
			if item.id == string(msg) {
				item.selected = true
				break
			}
		}

	case hitInfoMsg:
		m.hitInfo = string(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	// Build list of items (they may have changed due to selection)
	var itemElements []pony.Element
	for _, item := range m.items {
		itemElements = append(itemElements, item.Render())
	}

	// Update scroll view child with current items
	m.scrollView.Child = pony.NewVStack(itemElements...)
	m.scrollView.OffsetY = m.scrollOffset

	slots := map[string]pony.Element{
		"scrollview": m.scrollView,
	}

	data := TemplateData{
		SelectedItem: m.selectedItem,
		HitInfo:      m.hitInfo,
	}

	// Render with bounds
	scr, boundsMap := m.template.RenderWithBounds(data, slots, m.width, m.height)

	view := tea.NewView(scr.Render())
	view.AltScreen = true
	view.MouseMode = tea.MouseModeAllMotion

	// Set up callback using HitTestAll
	view.Callback = func(msg tea.Msg) tea.Cmd {
		switch msg := msg.(type) {
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
				if len(id) > 5 && id[:5] == "item-" {
					// Found a list item! Return batch of commands
					return tea.Batch(
						func() tea.Msg {
							return hitInfoMsg(hitInfo)
						},
						func() tea.Msg {
							return selectItemMsg(id)
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
	fmt.Println("\nThanks for trying HitTestAll!")
}
