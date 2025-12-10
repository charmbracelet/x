package main

import (
	"fmt"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/pony"
)

// Input is a stateful text input component.
type Input struct {
	pony.BaseElement
	label       string
	value       string
	placeholder string
	focused     bool
	cursorPos   int
}

func NewInput(label, placeholder string) *Input {
	return &Input{
		label:       label,
		placeholder: placeholder,
		focused:     false,
	}
}

func (i *Input) Update(msg tea.Msg) {
	if !i.focused {
		return
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "backspace":
			if len(i.value) > 0 && i.cursorPos > 0 {
				i.value = i.value[:i.cursorPos-1] + i.value[i.cursorPos:]
				i.cursorPos--
			}
		case "left":
			if i.cursorPos > 0 {
				i.cursorPos--
			}
		case "right":
			if i.cursorPos < len(i.value) {
				i.cursorPos++
			}
		case "home":
			i.cursorPos = 0
		case "end":
			i.cursorPos = len(i.value)
		default:
			// Handle printable characters
			if len(key.String()) == 1 {
				i.value = i.value[:i.cursorPos] + key.String() + i.value[i.cursorPos:]
				i.cursorPos++
			}
		}
	}
}

func (i *Input) Render() pony.Element {
	displayText := i.value
	if displayText == "" {
		displayText = i.placeholder
	}

	// Add cursor if focused
	if i.focused && i.cursorPos <= len(displayText) {
		if i.cursorPos == len(displayText) {
			displayText += "│"
		} else {
			displayText = displayText[:i.cursorPos] + "│" + displayText[i.cursorPos:]
		}
	}

	borderColor := pony.RGB(100, 100, 100)
	textColor := pony.RGB(255, 255, 255)
	if i.focused {
		borderColor = pony.Hex("#00FFFF")
		textColor = pony.Hex("#00FFFF")
	} else if i.value == "" {
		textColor = pony.RGB(128, 128, 128)
	}

	// Create the VStack and set the input's ID on it
	// so clicks anywhere in the input will register as clicking this input
	vstack := pony.NewVStack(
		pony.NewText(i.label).Bold(),
		pony.NewBox(
			pony.NewText(displayText).ForegroundColor(textColor),
		).Border("rounded").
			BorderColor(borderColor).
			Padding(1).
			Width(pony.NewFixedConstraint(50)),
	)
	vstack.SetID(i.ID()) // Set the input's ID on the rendered element

	return vstack
}

func (i *Input) Value() string   { return i.value }
func (i *Input) SetFocus(f bool) { i.focused = f }
func (i *Input) IsFocused() bool { return i.focused }

// ButtonBar is a component that renders action buttons.
type ButtonBar struct {
	pony.BaseElement
	showSubmit bool
	showClear  bool
	showQuit   bool
}

func NewButtonBar(showSubmit, showClear, showQuit bool) *ButtonBar {
	return &ButtonBar{
		showSubmit: showSubmit,
		showClear:  showClear,
		showQuit:   showQuit,
	}
}

func (b *ButtonBar) Render() pony.Element {
	buttons := []pony.Element{}

	if b.showSubmit {
		submitBtn := pony.NewButton("Submit")
		submitBtn.SetID("submit-btn")
		submitBtn = submitBtn.Border("rounded").
			Padding(1).
			Style(pony.NewStyle().Fg(pony.Hex("#00FF00")).Bold().Build())
		buttons = append(buttons, submitBtn)
	}

	if b.showClear {
		clearBtn := pony.NewButton("Clear")
		clearBtn.SetID("clear-btn")
		clearBtn = clearBtn.Border("rounded").
			Padding(1).
			Style(pony.NewStyle().Fg(pony.Hex("#FFFF00")).Build())
		buttons = append(buttons, clearBtn)
	}

	if b.showQuit {
		quitBtn := pony.NewButton("Quit")
		quitBtn.SetID("quit-btn")
		quitBtn = quitBtn.Border("rounded").
			Padding(1).
			Style(pony.NewStyle().Fg(pony.Hex("#FF0000")).Build())
		buttons = append(buttons, quitBtn)
	}

	return pony.NewHStack(buttons...).Spacing(2)
}

// Template.
const tmpl = `
<vstack spacing="1">
	<box border="double" border-color="yellow" padding="1">
		<text font-weight="bold" foreground-color="yellow" alignment="center">✨ Interactive Form Demo</text>
	</box>

	<divider foreground-color="gray" />

	<vstack spacing="1">
		<text font-weight="bold" foreground-color="cyan">User Registration Form</text>
		<text foreground-color="gray" font-style="italic">Click inputs to focus, type to fill, click buttons to submit</text>
	</vstack>

	<divider foreground-color="gray" />

	<vstack spacing="2">
		<slot name="name-input" />
		<slot name="email-input" />
		<slot name="username-input" />
	</vstack>

	<divider foreground-color="gray" />

	<slot name="button-bar" />

	{{ if .ShowStatus }}
	<divider foreground-color="gray" />

	<box border="rounded" border-color="{{ .StatusColor }}" padding="1">
		<text foreground-color="{{ .StatusColor }}" font-weight="bold">{{ .StatusMessage }}</text>
	</box>
	{{ end }}

	{{ if .ShowData }}
	<divider foreground-color="gray" />

	<box border="rounded" border-color="cyan" padding="1">
		<vstack spacing="0">
			<text font-weight="bold" foreground-color="cyan">Submitted Data:</text>
			<divider />
			<text>Name: {{ .Name }}</text>
			<text>Email: {{ .Email }}</text>
			<text>Username: {{ .Username }}</text>
		</vstack>
	</box>
	{{ end }}

	<divider foreground-color="gray" />

	<text foreground-color="gray" font-style="italic">Focused: {{ .FocusedInput }}</text>
	<text foreground-color="gray" font-style="italic">Hover: {{ .HoveredElement }}</text>
</vstack>
`

type ViewData struct {
	Name           string
	Email          string
	Username       string
	FocusedInput   string
	HoveredElement string
	ShowStatus     bool
	StatusMessage  string
	StatusColor    string
	ShowData       bool
}

type model struct {
	template       *pony.Template[ViewData]
	nameInput      *Input
	emailInput     *Input
	usernameInput  *Input
	buttonBar      *ButtonBar
	width          int
	height         int
	hoveredElement string
	showStatus     bool
	statusMessage  string
	statusColor    string
	showData       bool
	submittedName  string
	submittedEmail string
	submittedUser  string
}

func initialModel() model {
	nameInput := NewInput("Full Name:", "Enter your full name...")
	nameInput.SetID("name-input")

	emailInput := NewInput("Email Address:", "your.email@example.com")
	emailInput.SetID("email-input")

	usernameInput := NewInput("Username:", "Choose a username...")
	usernameInput.SetID("username-input")

	// Create button bar component
	buttonBar := NewButtonBar(true, true, true)

	// Focus the first input by default
	nameInput.SetFocus(true)

	return model{
		template:      pony.MustParse[ViewData](tmpl),
		nameInput:     nameInput,
		emailInput:    emailInput,
		usernameInput: usernameInput,
		buttonBar:     buttonBar,
		width:         80,
		height:        30,
	}
}

func (m model) Init() tea.Cmd {
	return tea.RequestWindowSize
}

// Custom messages.
type (
	buttonClickMsg string
	hoverMsg       string
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab":
			// Cycle focus through inputs
			if m.nameInput.IsFocused() {
				m.nameInput.SetFocus(false)
				m.emailInput.SetFocus(true)
			} else if m.emailInput.IsFocused() {
				m.emailInput.SetFocus(false)
				m.usernameInput.SetFocus(true)
			} else if m.usernameInput.IsFocused() {
				m.usernameInput.SetFocus(false)
				m.nameInput.SetFocus(true)
			}

		default:
			// Route to focused input
			m.nameInput.Update(msg)
			m.emailInput.Update(msg)
			m.usernameInput.Update(msg)
		}

	case buttonClickMsg:
		switch msg {
		case "submit-btn":
			// Validate and submit
			if m.nameInput.Value() == "" {
				m.showStatus = true
				m.statusMessage = "❌ Please enter your name"
				m.statusColor = "red"
				m.showData = false
			} else if m.emailInput.Value() == "" {
				m.showStatus = true
				m.statusMessage = "❌ Please enter your email"
				m.statusColor = "red"
				m.showData = false
			} else if m.usernameInput.Value() == "" {
				m.showStatus = true
				m.statusMessage = "❌ Please choose a username"
				m.statusColor = "red"
				m.showData = false
			} else if !strings.Contains(m.emailInput.Value(), "@") {
				m.showStatus = true
				m.statusMessage = "❌ Please enter a valid email address"
				m.statusColor = "red"
				m.showData = false
			} else {
				// Success!
				m.showStatus = true
				m.statusMessage = "✅ Form submitted successfully!"
				m.statusColor = "green"
				m.showData = true
				m.submittedName = m.nameInput.Value()
				m.submittedEmail = m.emailInput.Value()
				m.submittedUser = m.usernameInput.Value()
			}

		case "clear-btn":
			// Clear all inputs
			m.nameInput.value = ""
			m.nameInput.cursorPos = 0
			m.emailInput.value = ""
			m.emailInput.cursorPos = 0
			m.usernameInput.value = ""
			m.usernameInput.cursorPos = 0
			m.showStatus = true
			m.statusMessage = "🗑️  Form cleared"
			m.statusColor = "yellow"
			m.showData = false

		case "quit-btn":
			return m, tea.Quit

		case "name-input", "email-input", "username-input":
			// Focus the clicked input
			m.nameInput.SetFocus(msg == "name-input")
			m.emailInput.SetFocus(msg == "email-input")
			m.usernameInput.SetFocus(msg == "username-input")
		}

	case hoverMsg:
		m.hoveredElement = string(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	// XXX: view.Callback doesn't exist.
	return tea.NewView("")

	// // Prepare data
	// focusedInput := "none"
	// if m.nameInput.IsFocused() {
	// 	focusedInput = "name"
	// } else if m.emailInput.IsFocused() {
	// 	focusedInput = "email"
	// } else if m.usernameInput.IsFocused() {
	// 	focusedInput = "username"
	// }
	//
	// data := ViewData{
	// 	Name:           m.nameInput.Value(),
	// 	Email:          m.emailInput.Value(),
	// 	Username:       m.usernameInput.Value(),
	// 	FocusedInput:   focusedInput,
	// 	HoveredElement: m.hoveredElement,
	// 	ShowStatus:     m.showStatus,
	// 	StatusMessage:  m.statusMessage,
	// 	StatusColor:    m.statusColor,
	// 	ShowData:       m.showData,
	// }
	//
	// // Fill slots with stateful components
	// slots := map[string]pony.Element{
	// 	"name-input":     m.nameInput.Render(),
	// 	"email-input":    m.emailInput.Render(),
	// 	"username-input": m.usernameInput.Render(),
	// 	"button-bar":     m.buttonBar.Render(),
	// }
	//
	// // Render with bounds
	// scr, boundsMap := m.template.RenderWithBounds(data, slots, m.width, m.height)
	//
	// view := tea.NewView(scr.Render())
	// view.AltScreen = true
	// view.MouseMode = tea.MouseModeAllMotion
	//
	// // Set up callback for mouse events
	// view.Callback = func(msg tea.Msg) tea.Cmd {
	// 	switch msg := msg.(type) {
	// 	case tea.MouseClickMsg:
	// 		mouse := msg.Mouse()
	//
	// 		// Hit test to find clicked element
	// 		if elem := boundsMap.HitTest(mouse.X, mouse.Y); elem != nil {
	// 			return func() tea.Msg {
	// 				return buttonClickMsg(elem.ID())
	// 			}
	// 		}
	//
	// 	case tea.MouseMotionMsg:
	// 		mouse := msg.Mouse()
	//
	// 		// Track hover state
	// 		if elem := boundsMap.HitTest(mouse.X, mouse.Y); elem != nil {
	// 			return func() tea.Msg {
	// 				return hoverMsg(elem.ID())
	// 			}
	// 		} else {
	// 			return func() tea.Msg {
	// 				return hoverMsg("")
	// 			}
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
	fmt.Println("\n✨ Thanks for trying the interactive form demo!")
}
