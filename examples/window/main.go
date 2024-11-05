package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"
)

type dialog struct {
	win     *cellbuf.Window
	yes, no *cellbuf.Window
	hit     bool
	hitYes  bool
	hitNo   bool
}

var _ tea.Model = dialog{}

// Init implements tea.Model.
func (d dialog) Init() (tea.Model, tea.Cmd) {
	return d, nil
}

// Update implements tea.Model.
func (d dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.yes = d.win.Child(d.win.Width()/2-13, 4, 9, 1)
		d.no = d.win.Child(d.win.Width()/2+2, 4, 11, 1)
	case tea.MouseMsg:
		m := msg.Mouse()
		d.hit = d.win.InBounds(m.X, m.Y)
		d.hitYes = d.yes.InBounds(m.X, m.Y)
		d.hitNo = d.no.InBounds(m.X, m.Y)
	}
	return d, nil
}

// View implements tea.Model.
func (d dialog) View() string {
	if d.win == nil {
		return ""
	}

	var (
		buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#888B7E")).
				Padding(0, 3)

		activeButtonStyle = buttonStyle.
					Foreground(lipgloss.Color("#FFF7DB")).
					Background(lipgloss.Color("#F25D94"))
	)

	// Dialog.

	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	okButton := activeButtonStyle.Reverse(d.hitYes).Render("Yes")
	cancelButton := buttonStyle.Reverse(d.hitNo).Render("Maybe")

	grad := applyGradient(
		lipgloss.NewStyle().Reverse(d.hit),
		"Are you sure you want to eat marmalade?",
		lipgloss.Color("#EDFF82"),
		lipgloss.Color("#F25D94"),
	)

	question := lipgloss.NewStyle().
		Reverse(d.hit).
		Width(50).
		Align(lipgloss.Center).
		Render(grad)

	// buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	// ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)
	question += "\n\n"

	main := dialogBoxStyle.Reverse(d.hit).Render(question)

	cellbuf.SetContent(d.win, cellbuf.WcWidth, main)
	cellbuf.SetContent(d.yes, cellbuf.WcWidth, okButton)
	cellbuf.SetContent(d.no, cellbuf.WcWidth, cancelButton)

	return main
}

type model struct {
	win            *cellbuf.Window
	buf            *cellbuf.Buffer
	dialog         dialog
	dragDialog     bool
	hasDarkBg      bool
	clickX, clickY int
	width          int
}

var _ tea.Model = &model{}

// Init implements tea.Model.
func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.BackgroundColor
}

const (
	dialogWidth  = 52
	dialogHeight = 7
)

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.win.Resize(msg.Width, msg.Height)
		m.buf.Resize(msg.Width, msg.Height)
		cellbuf.Fill(m.buf, cellbuf.Cell{Content: " ", Width: 1})
		m.dialog.win = m.win.Child((m.width/2)-dialogWidth/2, 12, dialogWidth, dialogHeight)

		cmds = append(cmds, tea.ClearScreen)
	case tea.BackgroundColorMsg:
		m.hasDarkBg = msg.IsDark()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.MouseClickMsg:
		m.dragDialog = msg.Button == tea.MouseLeft && m.dialog.win.InBounds(msg.X, msg.Y)
		if m.dragDialog {
			m.clickX, m.clickY = msg.X, msg.Y
		}
	case tea.MouseReleaseMsg:
		m.dragDialog = false
	case tea.MouseMotionMsg:
		if msg.Button == tea.MouseLeft &&
			m.dragDialog &&
			m.dialog.win != nil {
			posX, posY := m.dialog.win.X(), m.dialog.win.Y()
			m.dialog.win.Move(posX+msg.X-m.clickX, posY+msg.Y-m.clickY)
			m.clickX, m.clickY = msg.X, msg.Y
		}
	}

	diag, cmd := m.dialog.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.dialog = diag.(dialog)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m model) View() string {
	lightDark := lipgloss.LightDark(m.hasDarkBg)
	columnWidth := (m.width / 3) - 3

	// Style definitions.
	var (

		// General.

		subtle    = lightDark("#D9DCCF", "#383838")
		highlight = lightDark("#874BFD", "#7D56F4")
		special   = lightDark("#43BF6D", "#73F59F")

		divider = lipgloss.NewStyle().
			SetString("•").
			Padding(0, 1).
			Foreground(subtle).
			String()

		url = lipgloss.NewStyle().Foreground(special).Render

		// Tabs.

		activeTabBorder = lipgloss.Border{
			Top:         "─",
			Bottom:      " ",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┘",
			BottomRight: "└",
		}

		tabBorder = lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┴",
			BottomRight: "┴",
		}

		tab = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

		activeTab = tab.Border(activeTabBorder, true)

		tabGap = tab.
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false)

		// Title.

		titleStyle = lipgloss.NewStyle().
				MarginLeft(1).
				MarginRight(5).
				Padding(0, 1).
				Italic(true).
				Foreground(lipgloss.Color("#FFF7DB")).
				SetString("Lip Gloss")

		descStyle = lipgloss.NewStyle().MarginTop(1)

		infoStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(subtle)

		// List.

		list = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(subtle).
			MarginRight(2).
			Height(8).
			Width(columnWidth + 1)

		listHeader = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(subtle).
				MarginRight(2).
				Render

		listItem = lipgloss.NewStyle().PaddingLeft(2).Render

		checkMark = lipgloss.NewStyle().SetString("✓").
				Foreground(special).
				PaddingRight(1).
				String()

		listDone = func(s string) string {
			return checkMark + lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lightDark("#969B86", "#696969")).
				Render(s)
		}

		// Paragraphs/History.

		historyStyle = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(highlight).
				Margin(1, 3, 0, 0).
				Padding(1, 2).
				Height(19).
				Width(columnWidth)

		// Status Bar.

		statusNugget = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Padding(0, 1)

		statusBarStyle = lipgloss.NewStyle().
				Foreground(lightDark("#343433", "#C1C6B2")).
				Background(lightDark("#D9DCCF", "#353533"))

		statusStyle = lipgloss.NewStyle().
				Inherit(statusBarStyle).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#FF5F87")).
				Padding(0, 1).
				MarginRight(1)

		encodingStyle = statusNugget.
				Background(lipgloss.Color("#A550DF")).
				Align(lipgloss.Right)

		statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

		fishCakeStyle = statusNugget.Background(lipgloss.Color("#6124DF"))

		// Page.

		docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	)

	doc := strings.Builder{}

	// Tabs.
	{
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			activeTab.Render("Lip Gloss"),
			tab.Render("Blush"),
			tab.Render("Eye Shadow"),
			tab.Render("Mascara"),
			tab.Render("Foundation"),
		)
		gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
		doc.WriteString(row + "\n\n")
	}

	// Title.
	{
		var (
			colors = colorGrid(1, 5)
			title  strings.Builder
		)

		for i, v := range colors {
			const offset = 2
			c := lipgloss.Color(v[0])
			fmt.Fprint(&title, titleStyle.MarginLeft(i*offset).Background(c))
			if i < len(colors)-1 {
				title.WriteRune('\n')
			}
		}

		desc := lipgloss.JoinVertical(lipgloss.Left,
			descStyle.Render("Style Definitions for Nice Terminal Layouts"),
			infoStyle.Render("From Charm"+divider+url("https://github.com/charmbracelet/lipgloss")),
		)

		row := lipgloss.JoinHorizontal(lipgloss.Top, title.String(), desc)
		doc.WriteString(row + "\n\n")
	}

	// Dialog.
	{
		dialog := lipgloss.Place(m.width-docStyle.GetHorizontalFrameSize(), 9,
			lipgloss.Center, lipgloss.Center,
			"",
			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(subtle)),
		)

		doc.WriteString(dialog + "\n\n")
	}

	// Color grid.
	colors := func() string {
		colors := colorGrid(14, 8)

		b := strings.Builder{}
		for _, x := range colors {
			for _, y := range x {
				s := lipgloss.NewStyle().SetString("  ").Background(lipgloss.Color(y))
				b.WriteString(s.String())
			}
			b.WriteRune('\n')
		}

		return b.String()
	}()

	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		list.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				listHeader("Citrus Fruits to Try"),
				listDone("Grapefruit"),
				listDone("Yuzu"),
				listItem("Citron"),
				listItem("Kumquat"),
				listItem("Pomelo"),
			),
		),
		list.Width(columnWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				listHeader("Actual Lip Gloss Vendors"),
				listItem("Glossier"),
				listItem("Claire‘s Boutique"),
				listDone("Nyx"),
				listItem("Mac"),
				listDone("Milk"),
			),
		),
	)

	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, lists, colors))

	// Marmalade history.
	{
		const (
			historyA = "The Romans learned from the Greeks that quinces slowly cooked with honey would “set” when cool. The Apicius gives a recipe for preserving whole quinces, stems and leaves attached, in a bath of honey diluted with defrutum: Roman marmalade. Preserves of quince and lemon appear (along with rose, apple, plum and pear) in the Book of ceremonies of the Byzantine Emperor Constantine VII Porphyrogennetos."
			historyB = "Medieval quince preserves, which went by the French name cotignac, produced in a clear version and a fruit pulp version, began to lose their medieval seasoning of spices in the 16th century. In the 17th century, La Varenne provided recipes for both thick and clear cotignac."
			historyC = "In 1524, Henry VIII, King of England, received a “box of marmalade” from Mr. Hull of Exeter. This was probably marmelada, a solid quince paste from Portugal, still made and sold in southern Europe today. It became a favourite treat of Anne Boleyn and her ladies in waiting."
		)

		doc.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			historyStyle.Align(lipgloss.Right).Render(historyA),
			historyStyle.Align(lipgloss.Center).Render(historyB),
			historyStyle.MarginRight(0).Render(historyC),
		))

		doc.WriteString("\n\n")
	}

	// Status bar.
	{
		w := lipgloss.Width

		lightDarkState := "Light"
		if m.hasDarkBg {
			lightDarkState = "Dark"
		}

		statusKey := statusStyle.Render("STATUS")
		encoding := encodingStyle.Render("UTF-8")
		fishCake := fishCakeStyle.Render("🍥 Fish Cake")
		statusVal := statusText.
			Width(m.width - w(statusKey) - w(encoding) - w(fishCake) - docStyle.GetHorizontalFrameSize()).
			Render("Ravishingly " + lightDarkState + "!")

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			encoding,
			fishCake,
		)

		doc.WriteString(statusBarStyle.Width(m.width).Render(bar))
	}

	// Okay, let's print it. We use a special Lipgloss writer to downsample
	// colors to the terminal's color palette. And, if output's not a TTY, we
	// will remove color entirely.
	main := docStyle.Render(doc.String())
	cellbuf.SetContent(m.win, cellbuf.WcWidth, main)
	if m.dialog.win != nil {
		m.dialog.View()
	}

	return cellbuf.Render(m.win)
}

func main() {
	var buf cellbuf.Buffer
	win := cellbuf.NewRootWindow(&buf)
	p := tea.NewProgram(model{win: win, buf: &buf}, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// applyGradient applies a gradient to the given string string.
func applyGradient(base lipgloss.Style, input string, from, to color.Color) string {
	// We want to get the graphemes of the input string, which is the number of
	// characters as a human would see them.
	//
	// We definitely don't want to use len(), because that returns the
	// bytes. The rune count would get us closer but there are times, like with
	// emojis, where the rune count is greater than the number of actual
	// characters.
	g := uniseg.NewGraphemes(input)
	var chars []string
	for g.Next() {
		chars = append(chars, g.Str())
	}

	// Genrate the blend.
	a, _ := colorful.MakeColor(to)
	b, _ := colorful.MakeColor(from)
	var output strings.Builder
	var hex string
	for i := 0; i < len(chars); i++ {
		hex = a.BlendLuv(b, float64(i)/float64(len(chars)-1)).Hex()
		output.WriteString(base.Foreground(lipgloss.Color(hex)).Render(chars[i]))
	}

	return output.String()
}
