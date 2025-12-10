package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// ViewData is the type-safe data structure for our template.
type ViewData struct {
	Title  string
	Count  int
	Width  int
	Height int
}

const tmpl = `
<vstack spacing="1">
	<box border="rounded" border-color="cyan">
		<text font-weight="bold" foreground-color="yellow">{{ .Title }}</text>
	</box>
	<text>Counter: {{ .Count }}</text>
	<text foreground-color="gray">Window: {{ .Width }}x{{ .Height }}</text>
	<divider foreground-color="gray" />
	<text font-style="italic" foreground-color="gray">Press space to increment, q to quit</text>
</vstack>
`

type model struct {
	template *pony.Template[ViewData]
	count    int
	width    int
	height   int
}

func initialModel() model {
	return model{
		template: pony.MustParse[ViewData](tmpl),
		count:    0,
		width:    80,
		height:   24,
	}
}

func (m model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ", "space":
			m.count++
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	// Type-safe data structure!
	data := ViewData{
		Title:  "Simple pony + Bubble Tea",
		Count:  m.count,
		Width:  m.width,
		Height: m.height,
	}

	output := m.template.Render(data, m.width, m.height)
	return tea.NewView(output)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nThanks for trying pony!")
}
