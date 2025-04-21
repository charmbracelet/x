package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

const (
	blackCircle = "●"
	whiteCircle = "○"
	rightArrow  = "→"
)

func main() {
	keys := charmtone.Keys()
	tones := charmtone.Tones()
	hexes := charmtone.Hexes()

	// Find the longest key name.
	var widestKeyName int
	for k := range tones {
		if w := lipgloss.Width(k.String()); w > widestKeyName {
			widestKeyName = w
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
		Width(widestKeyName).
		Align(lipgloss.Right)
	bg := lipgloss.NewStyle().
		Width(8)
	hex := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Smoke], tones[charmtone.Charcoal]))
	legend := subdued.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(tones[charmtone.Charcoal]).
		Padding(0, 2).
		MarginLeft(2)
	primaryMark := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Squid], tones[charmtone.Smoke])).
		SetString(blackCircle)
	secondaryMark := primaryMark.
		Foreground(lightDark(tones[charmtone.Squid], tones[charmtone.Oyster])).
		SetString(blackCircle)
	tertiaryMark := primaryMark.
		Foreground(lightDark(tones[charmtone.Squid], tones[charmtone.Oyster])).
		SetString(whiteCircle)
	rightArrowMark := lipgloss.NewStyle().
		Foreground(lightDark(tones[charmtone.Squid], tones[charmtone.Oyster])).
		Margin(0, 1).
		SetString(rightArrow)

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
		switch {
		case charmtone.IsPrimary(k):
			mark = primaryMark.String()
		case charmtone.IsSecondary(k):
			mark = secondaryMark.String()
		case charmtone.IsTertiary(k):
			mark = tertiaryMark.String()
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

	// Get total block width so far.
	var totalWidth int
	for l := range strings.SplitSeq(b.String(), "\n") {
		if w := lipgloss.Width(l); w > totalWidth {
			totalWidth = w
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

	// Get width of grayscale block.
	var grayWidth int
	for l := range strings.SplitSeq(grays.String(), "\n") {
		if w := lipgloss.Width(l); w > grayWidth {
			grayWidth = w
		}
	}

	fmt.Fprint(&b, "\n")

	// Build legend.
	legendBlock := legend.Render(
		strings.Join([]string{
			primaryMark.String() + subdued.Render(" Primary"),
			secondaryMark.String() + subdued.Render(" Secondary"),
			tertiaryMark.String() + subdued.Render(" Tertiary"),
		}, "  "),
	)

	// Build gradients.
	var grads strings.Builder
	gap := "  "
	gapWidth := lipgloss.Width(gap)
	{
		fullWidth := (totalWidth - grayWidth) - lipgloss.Width(gap)
		if fullWidth%2 != 0 {
			fullWidth--
		}
		halfWidth := fullWidth / gapWidth

		block := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(halfWidth)
		s := subdued.
			Foreground(tones[charmtone.Squid])

		left := blendKeys(halfWidth, charmtone.Hazy, charmtone.Blush)
		left += "\n" + block.Render(s.Render("Hazy")+rightArrowMark.String()+s.Render("Blush"))
		right := blendKeys(halfWidth, charmtone.Bok, charmtone.Zest)
		right += "\n" + block.Render(s.Render("Bok")+rightArrowMark.String()+s.Render("Zest"))
		fmt.Fprint(&grads, "\n", lipgloss.JoinHorizontal(lipgloss.Top, gap, left, right))

		block = block.Width(fullWidth)
		buf := strings.Builder{}
		fmt.Fprint(&buf, blendKeys(fullWidth, charmtone.Uni,
			charmtone.Coral, charmtone.Tuna, charmtone.Violet,
			charmtone.Malibu, charmtone.Turtle,
		))
		fmt.Fprint(&buf, "\n",
			block.Render(
				s.Render("Uni")+rightArrowMark.String()+
					s.Render("Coral")+rightArrowMark.String()+
					s.Render("Tuna")+rightArrowMark.String()+
					s.Render("Violet")+rightArrowMark.String()+
					s.Render("Malibu")+rightArrowMark.String()+
					s.Render("Turtle"),
			),
		)
		fmt.Fprint(&grads, "\n\n", lipgloss.JoinHorizontal(lipgloss.Top, gap, buf.String()))
	}

	// Join Greys and legend.
	fmt.Fprint(&b, lipgloss.JoinHorizontal(lipgloss.Top, grays.String(), " ", grads.String()))

	fmt.Fprint(&b, "\n\n", legendBlock, "\n\n")

	// Flush.
	lipgloss.Print(b.String())
}

func blendKeys(width int, keys ...charmtone.Key) string {
	var w strings.Builder
	for _, c := range charmtone.BlendColors(width, keys...) {
		fmt.Fprint(&w, lipgloss.NewStyle().Background(c).Render(" "))
	}
	return w.String()
}
