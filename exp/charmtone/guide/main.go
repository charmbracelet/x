package main

import (
	"fmt"
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

	// Styles.
	hasDarkBG := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(hasDarkBG)
	logo := lipgloss.NewStyle().
		Foreground(tones[charmtone.Ash]).
		Background(tones[charmtone.Charple]).
		Padding(0, 1).
		SetString("Charm™")
	title := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Charcoal], tones[charmtone.Smoke]))
	subdued := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Squid], tones[charmtone.Oyster]))
	fg := lipgloss.NewStyle().
		MarginLeft(2).
		Width(width).
		Align(lipgloss.Right)
	bg := lipgloss.NewStyle().
		Width(8)
	hex := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Smoke], tones[charmtone.Charcoal]))

	var b strings.Builder

	// Render title and description.
	fmt.Fprintf(
		&b,
		"\n  %s %s %s\n\n",
		logo.String(),
		title.Render("CharmTone"),
		subdued.Render("• Formula Guide"),
	)

	// Render swatches.
	for i, k := range keys {
		block := fmt.Sprintf(
			"%s %s %s",
			fg.Foreground(tones[k]).Underline(charmtone.IsCore(k)).Render(k.String()),
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

	fmt.Fprintf(&b, "\n  %s\n", subdued.Render("Underline: Core System"))

	fmt.Println(b.String())
}
