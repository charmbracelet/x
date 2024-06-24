package colors

import "github.com/charmbracelet/lipgloss"

var (
	WhiteBright = lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}

	Normal    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#dddddd"}
	NormalDim = lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}

	Gray          = lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"}
	GrayMid       = lipgloss.AdaptiveColor{Light: "#B2B2B2", Dark: "#4A4A4A"}
	GrayDark      = lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#222222"}
	GrayBright    = lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"}
	GrayBrightDim = lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}

	Indigo          = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	IndigoDim       = lipgloss.AdaptiveColor{Light: "#9498FF", Dark: "#494690"}
	IndigoSubtle    = lipgloss.AdaptiveColor{Light: "#7D79F6", Dark: "#514DC1"}
	IndigoSubtleDim = lipgloss.AdaptiveColor{Light: "#BBBDFF", Dark: "#383584"}

	YellowGreen     = lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#ECFD65"}
	YellowGreenDull = lipgloss.AdaptiveColor{Light: "#6BCB94", Dark: "#9BA92F"}

	Fuschia        = lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}
	FuchsiaDim     = lipgloss.AdaptiveColor{Light: "#F1A8FF", Dark: "#99519E"}
	FuchsiaDull    = lipgloss.AdaptiveColor{Dark: "#AD58B4", Light: "#F793FF"}
	FuchsiaDullDim = lipgloss.AdaptiveColor{Light: "#F6C9FF", Dark: "#6B3A6F"}

	Green    = lipgloss.Color("#04B575")
	GreenDim = lipgloss.AdaptiveColor{Light: "#72D2B0", Dark: "#0B5137"}

	Red     = lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}
	RedDull = lipgloss.AdaptiveColor{Light: "#FF6F91", Dark: "#C74665"}
)
