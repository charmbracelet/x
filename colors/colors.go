package colors

import "github.com/charmbracelet/lipgloss"

var (
	Light = Colors{
		WhiteBright:     lipgloss.Color("#FFFDF5"),
		Normal:          lipgloss.Color("#1A1A1A"),
		NormalDim:       lipgloss.Color("#A49FA5"),
		Gray:            lipgloss.Color("#909090"),
		GrayMid:         lipgloss.Color("#B2B2B2"),
		GrayDark:        lipgloss.Color("#DDDADA"),
		GrayBright:      lipgloss.Color("#847A85"),
		GrayBrightDim:   lipgloss.Color("#C2B8C2"),
		Indigo:          lipgloss.Color("#5A56E0"),
		IndigoDim:       lipgloss.Color("#9498FF"),
		IndigoSubtle:    lipgloss.Color("#7D79F6"),
		IndigoSubtleDim: lipgloss.Color("#BBBDFF"),
		YellowGreen:     lipgloss.Color("#04B575"),
		YellowGreenDull: lipgloss.Color("#6BCB94"),
		Fuschia:         lipgloss.Color("#EE6FF8"),
		FuchsiaDim:      lipgloss.Color("#F1A8FF"),
		FuchsiaDull:     lipgloss.Color("#AD58B4"),
		FuchsiaDullDim:  lipgloss.Color("#F6C9FF"),
		Green:           lipgloss.Color("#04B575"),
		GreenDim:        lipgloss.Color("#72D2B0"),
		Red:             lipgloss.Color("#FF4672"),
		RedDull:         lipgloss.Color("#FF6F91"),
	}

	Dark = Colors{
		WhiteBright:     lipgloss.Color("#FFFDF5"),
		Normal:          lipgloss.Color("#dddddd"),
		NormalDim:       lipgloss.Color("#777777"),
		Gray:            lipgloss.Color("626262"),
		GrayMid:         lipgloss.Color("#4A4A4A"),
		GrayDark:        lipgloss.Color("#222222"),
		GrayBright:      lipgloss.Color("#979797"),
		GrayBrightDim:   lipgloss.Color("#4D4D4D"),
		Indigo:          lipgloss.Color("#7571F9"),
		IndigoDim:       lipgloss.Color("#494690"),
		IndigoSubtle:    lipgloss.Color("#514DC1"),
		IndigoSubtleDim: lipgloss.Color("#383584"),
		YellowGreen:     lipgloss.Color("#ECFD65"),
		YellowGreenDull: lipgloss.Color("#9BA92F"),
		Fuschia:         lipgloss.Color("#EE6FF8"),
		FuchsiaDim:      lipgloss.Color("#99519E"),
		FuchsiaDull:     lipgloss.Color("#AD58B4"),
		FuchsiaDullDim:  lipgloss.Color("#6B3A6F"),
		Green:           lipgloss.Color("#04B575"),
		GreenDim:        lipgloss.Color("#0B5137"),
		Red:             lipgloss.Color("#ED567A"),
		RedDull:         lipgloss.Color("#C74665"),
	}
)

type Colors struct {
	WhiteBright lipgloss.TerminalColor

	Normal    lipgloss.TerminalColor
	NormalDim lipgloss.TerminalColor

	Gray          lipgloss.TerminalColor
	GrayMid       lipgloss.TerminalColor
	GrayDark      lipgloss.TerminalColor
	GrayBright    lipgloss.TerminalColor
	GrayBrightDim lipgloss.TerminalColor

	Indigo          lipgloss.TerminalColor
	IndigoDim       lipgloss.TerminalColor
	IndigoSubtle    lipgloss.TerminalColor
	IndigoSubtleDim lipgloss.TerminalColor

	YellowGreen     lipgloss.TerminalColor
	YellowGreenDull lipgloss.TerminalColor

	Fuschia        lipgloss.TerminalColor
	FuchsiaDim     lipgloss.TerminalColor
	FuchsiaDull    lipgloss.TerminalColor
	FuchsiaDullDim lipgloss.TerminalColor

	Green    lipgloss.TerminalColor
	GreenDim lipgloss.TerminalColor

	Red     lipgloss.TerminalColor
	RedDull lipgloss.TerminalColor
}
