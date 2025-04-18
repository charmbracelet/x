package main

import (
	"fmt"
	"io"
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
	legend := subdued.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(tones[charmtone.Charcoal]).
		Padding(0, 2).
		MarginLeft(2).
		MarginTop(1)
	primaryMark := lipgloss.NewStyle().Foreground(tones[charmtone.Squid]).SetString("◉")
	secondaryMark := primaryMark.SetString("○")

	var b strings.Builder

	// Render title and description.
	fmt.Fprintf(
		&b,
		"\n  %s %s %s\n\n",
		logo.String(),
		title.Render("CharmTone"),
		subdued.Render("• Formula Guide"),
	)

	// Render a swatch and its metadata.
	renderSwatch := func(w io.Writer, k charmtone.Key) {
		mark := " "
		if charmtone.IsPrimary(k) {
			mark = primaryMark.String()
		} else if charmtone.IsSecondary(k) {
			mark = secondaryMark.String()
		}
		_, _ = fmt.Fprintf(w,
			"%s %s %s %s",
			fg.Foreground(tones[k]).Render(k.String()),
			mark,
			bg.Background(tones[k]).Render(),
			hex.Render(hexes[k]),
		)
	}

	// Render main color block.
	for i := charmtone.Cumin; i < charmtone.Pepper; i++ {
		k := keys[i]
		renderSwatch(&b, k)
		if i%3 == 2 {
			b.WriteRune('\n')
		} else {
			b.WriteRune(' ')
		}
	}

	// Grayscale block.
	var grays strings.Builder
	for i := charmtone.Pepper; i <= charmtone.Butter; i++ {
		k := keys[i]
		renderSwatch(&grays, k)
		if i < charmtone.Butter {
			grays.WriteRune('\n')
		}
	}

	// Build legend.
	legendBlock := legend.Render(
		primaryMark.String() + subdued.Render(" Primary") + "\n" +
			secondaryMark.String() + subdued.Render(" Secondary"),
	)

	// Join Greys and legend.
	fmt.Fprint(&b, lipgloss.JoinHorizontal(lipgloss.Top, grays.String(), " ", legendBlock))

	// Flush.
	fmt.Println(b.String() + "\n")
}
