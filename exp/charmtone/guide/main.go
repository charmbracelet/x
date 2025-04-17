package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

func main() {
	// Find the longest key name.
	var width int
	for k := range charmtone.Tones {
		if w := lipgloss.Width(k.String()); w > width {
			width = w
		}
	}

	fg := lipgloss.NewStyle().
		MarginLeft(2).
		Width(width).
		Align(lipgloss.Right)
	bg := lipgloss.NewStyle().
		Width(8)

	var b strings.Builder
	for i, k := range charmtone.Keys {
		block := fmt.Sprintf(
			"%s %s",
			fg.Foreground(charmtone.Tones[k]).Render(k.String()),
			bg.Background(charmtone.Tones[k]).Render(""),
		)

		fmt.Fprintf(&b, "%s", block)
		if i == int(charmtone.Pepper)-1 {
			b.WriteRune('\n')
		}
		if i%3 == 2 || i >= int(charmtone.Pepper) {
			b.WriteRune('\n')
		} else {
			b.WriteRune(' ')
		}
	}

	fmt.Println("\n" + b.String())
}
