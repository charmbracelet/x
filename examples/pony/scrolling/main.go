// Package main example.
package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// Scrollable log component.
type ScrollLog struct {
	lines  []string
	offset int
	height int
}

func NewScrollLog(height int) *ScrollLog {
	var lines []string
	for i := 1; i <= 100; i++ {
		lines = append(lines, fmt.Sprintf("[%02d] Log entry %d", i, i))
	}
	return &ScrollLog{lines: lines, height: height}
}

func (s *ScrollLog) Update(msg tea.Msg) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "down", "j":
			s.offset = min(s.offset+1, len(s.lines)-s.height)
		case "up", "k":
			s.offset = max(0, s.offset-1)
		}
	}
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		m := wheel.Mouse()
		switch m.Button {
		case tea.MouseWheelUp:
			s.offset = max(0, s.offset-3)
		case tea.MouseWheelDown:
			s.offset = min(s.offset+3, len(s.lines)-s.height)
		}
	}
}

func (s *ScrollLog) Render() pony.Element {
	items := make([]pony.Element, 0, len(s.lines))
	for _, line := range s.lines {
		items = append(items, pony.NewText(line))
	}

	return pony.NewScrollView(pony.NewVStack(items...)).
		Offset(0, s.offset).
		Height(pony.NewFixedConstraint(s.height)).
		Vertical(true).
		Scrollbar(true)
}

// Template.
const tmpl = `
<vstack spacing="1">
	<box border="rounded" border-color="cyan" padding="1">
		<text font-weight="bold" foreground-color="cyan" alignment="center">Scrolling Demo</text>
	</box>

	<divider foreground-color="gray" />

	<box border="rounded" border-color="green">
		<slot name="log" />
	</box>

	<spacer />

	<text font-style="italic" foreground-color="gray">↑/↓ or j/k to scroll, mouse wheel works too • q to quit</text>
</vstack>
`

type model struct {
	template *pony.Template[any]
	log      *ScrollLog
	width    int
	height   int
}

func (m model) Init() tea.Cmd { return tea.RequestWindowSize }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.log.Update(msg)
	case tea.MouseWheelMsg:
		m.log.Update(msg)
	}
	return m, nil
}

func (m model) View() tea.View {
	slots := map[string]pony.Element{
		"log": m.log.Render(),
	}

	output := m.template.RenderWithSlots(nil, slots, m.width, m.height)
	view := tea.NewView(output)
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func main() {
	m := model{
		template: pony.MustParse[any](tmpl),
		log:      NewScrollLog(15),
		width:    80,
		height:   24,
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nThanks!")
}
