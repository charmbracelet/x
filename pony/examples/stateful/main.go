package main

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// Simple stateful input component
type Input struct {
	value   string
	focused bool
}

func NewInput(placeholder string) *Input {
	return &Input{}
}

// Update handles events
func (i *Input) Update(msg tea.Msg) {
	if !i.focused {
		return
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "backspace":
			if len(i.value) > 0 {
				i.value = i.value[:len(i.value)-1]
			}
		default:
			if len(key.String()) == 1 {
				i.value += key.String()
			}
		}
	}
}

// Render returns pony elements
func (i *Input) Render() pony.Element {
	displayText := i.value
	if displayText == "" {
		displayText = "Type something..."
	}

	// Build UI using helpers
	style := pony.NewStyle().Fg(pony.RGB(255, 255, 255)).Build()
	borderColor := pony.Hex("#00FFFF")
	if !i.focused {
		borderColor = pony.RGB(128, 128, 128)
	}
	borderStyle := pony.NewStyle().Fg(borderColor).Build()

	return pony.NewBox(
		pony.NewText(displayText).WithStyle(style),
	).WithBorder("rounded").
		WithBorderStyle(borderStyle).
		WithPadding(1).
		WithWidth(pony.NewFixedConstraint(40))
}

func (i *Input) Value() string   { return i.value }
func (i *Input) SetFocus(f bool) { i.focused = f }

// Template
const tmpl = `
<vstack gap="1">
	<box border="rounded" border-style="fg:yellow; bold" padding="1">
		<text style="bold; fg:yellow" align="center">Stateful Components Demo</text>
	</box>

	<divider style="fg:gray" />

	<vstack gap="1">
		<text style="bold">Enter your name:</text>
		<slot name="input" />
	</vstack>

	<divider style="fg:gray" />

	<box padding="1">
		<vstack gap="0">
			<text style="bold">Live Value:</text>
			<text style="fg:cyan">{{ .InputValue }}</text>
		</vstack>
	</box>

	<text style="fg:gray; italic">Tab to focus, type to edit, Esc to quit</text>
</vstack>
`

type ViewData struct {
	InputValue string
}

type model struct {
	template *pony.Template[ViewData]
	input    *Input
	width    int
	height   int
}

func initialModel() model {
	m := model{
		template: pony.MustParse[ViewData](tmpl),
		input:    NewInput("Enter text..."),
		width:    80,
		height:   24,
	}
	m.input.SetFocus(true)
	return m
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
		if msg.String() == "esc" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Route events to input
		m.input.Update(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	// Prepare data
	data := ViewData{
		InputValue: m.input.Value(),
	}

	// Fill slots with stateful component
	slots := map[string]pony.Element{
		"input": m.input.Render(),
	}

	output := m.template.RenderWithSlots(data, slots, m.width, m.height)
	return tea.NewView(output)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nThanks!")
}
