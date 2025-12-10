// Package main example.
package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// ListItem is a clickable list item.
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
		text = text.BackgroundColor(pony.RGB(0, 0, 255)).
			ForegroundColor(pony.RGB(255, 255, 255)).
			Bold()
	}

	box := pony.NewBox(text).
		Padding(1).
		Border("rounded")

	if l.selected {
		box = box.BorderColor(pony.RGB(0, 0, 255))
	}

	// Set the component ID on the root element
	box.SetID(l.id)

	return box
}

// Template.
const tmpl = `
<vstack spacing="1">
	<box border="double" border-color="cyan" padding="1">
		<text font-weight="bold" foreground-color="yellow" alignment="center">Scroll View with Clickable Items</text>
	</box>

	<divider foreground-color="gray" />

	<text foreground-color="gray">Selected: {{ .SelectedItem }}</text>
	<text font-style="italic" foreground-color="gray">{{ .HitInfo }}</text>

	<divider foreground-color="gray" />

	<box border="rounded" border-color="green">
		<slot name="scrollview" />
	</box>

	<divider foreground-color="gray" />

	<text font-style="italic" foreground-color="gray">Click items to select • Mouse wheel to scroll • q to quit</text>
</vstack>
`

type TemplateData struct {
	SelectedItem string
	HitInfo      string
}

type model struct {
	template     *pony.Template[TemplateData]
	items        []*ListItem
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

type (
	selectItemMsg string
	hitInfoMsg    string
)

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
		// Use fixed viewport height
		viewportHeight := 12
		// Calculate content height (20 items, each rendered as box with padding/border)
		// Approximate: each item is ~5 lines tall
		contentHeight := len(m.items) * 5
		maxOffset := max(0, contentHeight-viewportHeight)

		switch mouse.Button {
		case tea.MouseWheelUp:
			m.scrollOffset = max(0, m.scrollOffset-3)
		case tea.MouseWheelDown:
			m.scrollOffset = min(maxOffset, m.scrollOffset+3)
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
	// XXX: view.Callback doesn't exist.
	return tea.NewView("")

	// // Build list of items (they may have changed due to selection)
	// var itemElements []pony.Element
	// for _, item := range m.items {
	// 	itemElements = append(itemElements, item.Render())
	// }
	//
	// // Create scroll view with current offset
	// scrollView := pony.NewScrollView(pony.NewVStack(itemElements...))
	// scrollView.SetID("main-scroll-view")
	// scrollView = scrollView.
	// 	Height(pony.NewFixedConstraint(12)).
	// 	Vertical(true).
	// 	Scrollbar(true).
	// 	Offset(0, m.scrollOffset)
	//
	// slots := map[string]pony.Element{
	// 	"scrollview": scrollView,
	// }
	//
	// data := TemplateData{
	// 	SelectedItem: m.selectedItem,
	// 	HitInfo:      m.hitInfo,
	// }
	//
	// // Render with bounds
	// scr, boundsMap := m.template.RenderWithBounds(data, slots, m.width, m.height)
	//
	// view := tea.NewView(scr.Render())
	// view.AltScreen = true
	// view.MouseMode = tea.MouseModeAllMotion
	//
	// // Set up callback using HitTestAll
	// view.Callback = func(msg tea.Msg) tea.Cmd {
	// 	switch msg := msg.(type) {
	// 	case tea.MouseClickMsg:
	// 		mouse := msg.Mouse()
	//
	// 		// Use HitTestAll to get all elements at click position
	// 		hits := boundsMap.HitTestAll(mouse.X, mouse.Y)
	//
	// 		if len(hits) == 0 {
	// 			return nil
	// 		}
	//
	// 		// Log what we hit for debugging
	// 		var hitIDs []string
	// 		for _, elem := range hits {
	// 			hitIDs = append(hitIDs, elem.ID())
	// 		}
	// 		hitInfo := fmt.Sprintf("Hit %d elements: %v", len(hits), hitIDs)
	//
	// 		// Check if we clicked a list item
	// 		for _, elem := range hits {
	// 			id := elem.ID()
	// 			if len(id) > 5 && id[:5] == "item-" {
	// 				// Found a list item! Return batch of commands
	// 				return tea.Batch(
	// 					func() tea.Msg {
	// 						return hitInfoMsg(hitInfo)
	// 					},
	// 					func() tea.Msg {
	// 						return selectItemMsg(id)
	// 					},
	// 				)
	// 			}
	// 		}
	//
	// 		// If no item found, just update hit info
	// 		return func() tea.Msg {
	// 			return hitInfoMsg(hitInfo)
	// 		}
	// 	}
	// 	return nil
	// }
	//
	// return view
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nThanks for trying HitTestAll!")
}
