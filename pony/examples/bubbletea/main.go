package main

import (
	"fmt"
	"log"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// TemplateData represents the data passed to the template
type TemplateData struct {
	Title    string
	Count    int
	Time     string
	Events   []string
	ShowHelp bool
	Width    int
	Height   int
}

// Define our template with dynamic content
const tmpl = `
<vstack gap="1">
	<box border="double" border-style="fg:cyan; bold">
		<text style="bold; fg:yellow">{{ .Title }}</text>
	</box>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Counter Demo:</text>
		<text style="fg:cyan">Count: {{ .Count }}</text>
		<text style="fg:magenta">Time: {{ .Time }}</text>
		<text style="fg:gray">Window: {{ .Width }}x{{ .Height }}</text>
	</vstack>

	<divider style="fg:gray" />

	<vstack gap="0">
		<text style="bold">Recent Events:</text>
		{{ range .Events }}
		<text style="fg:green">• {{ . }}</text>
		{{ end }}
	</vstack>

	<divider style="fg:gray" />

	{{ if .ShowHelp }}
	<box border="rounded" border-style="fg:blue">
		<vstack>
			<text style="bold; fg:blue">Help:</text>
			<text>Press 'space' to increment counter</text>
			<text>Press 'r' to reset</text>
			<text>Press 'h' to toggle help</text>
			<text>Press 'q' or 'ctrl+c' to quit</text>
		</vstack>
	</box>
	{{ else }}
	<text style="italic; fg:gray">Press 'h' for help</text>
	{{ end }}

	{{ if gt .Count 10 }}
	<text style="fg:yellow; bold">🎉 You reached {{ .Count }}!</text>
	{{ end }}
</vstack>
`

type model struct {
	template  *pony.Template[TemplateData]
	count     int
	events    []string
	showHelp  bool
	startTime time.Time
	width     int
	height    int
}

func initialModel() model {
	return model{
		template:  pony.MustParse[TemplateData](tmpl),
		count:     0,
		events:    []string{"Application started"},
		showHelp:  true,
		startTime: time.Now(),
		width:     80,
		height:    24,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), tea.RequestWindowSize)
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
			m.events = append([]string{fmt.Sprintf("Counter incremented to %d", m.count)}, m.events...)
			if len(m.events) > 5 {
				m.events = m.events[:5]
			}
		case "r":
			m.count = 0
			m.events = append([]string{"Counter reset"}, m.events...)
			if len(m.events) > 5 {
				m.events = m.events[:5]
			}
		case "h":
			m.showHelp = !m.showHelp
			action := "shown"
			if !m.showHelp {
				action = "hidden"
			}
			m.events = append([]string{fmt.Sprintf("Help %s", action)}, m.events...)
			if len(m.events) > 5 {
				m.events = m.events[:5]
			}
		}

	case tickMsg:
		return m, tick()
	}

	return m, nil
}

func (m model) View() tea.View {
	elapsed := time.Since(m.startTime)

	// Prepare data for template (type-safe!)
	data := TemplateData{
		Title:    "pony + Bubble Tea Demo",
		Count:    m.count,
		Time:     elapsed.Round(time.Second).String(),
		Events:   m.events,
		ShowHelp: m.showHelp,
		Width:    m.width,
		Height:   m.height,
	}

	// Render pony template with data to fit terminal size
	output := m.template.Render(data, m.width, m.height)

	// Return as Bubble Tea View
	view := tea.NewView(output)
	view.AltScreen = true
	return view
}

// Messages
type tickMsg time.Time

func tick() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second)
		return tickMsg(time.Now())
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
