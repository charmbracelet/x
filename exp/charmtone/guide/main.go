package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

func main() {
	keys := charmtone.Keys()
	tones := charmtone.Tones()
	hexes := charmtone.Hexes()

	// Find the longest key name.
	var width int
	for k := range tones {
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
	hex := lipgloss.NewStyle().
		Foreground(func() color.Color {
			if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
				return tones[charmtone.Charcoal]
			}
			return tones[charmtone.Smoke]
		}())

	var b strings.Builder
	for i, k := range keys {
		block := fmt.Sprintf(
			"%s %s %s",
			fg.Foreground(tones[k]).Render(k.String()),
			bg.Background(tones[k]).Render(),
			hex.Render(hexes[k]),
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
