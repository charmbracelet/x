// Package color contains utilities for working with colors.
package color

import (
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

// Blend returns a slice of colors blended between the given
// colors. Blending is done as Hcl to stay in gamut.
//
// Example:
//
//	red := color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
//	blue := color.RGBA{R: 0x00, G: 0x00, B: 0xff, A: 0xff}
//
//	blend := Blend(10, red, blue)
//	var b strings.Builder
//
//	for _, c := range blend {
//		b.WriteString(lipgloss.NewStyle().Background(c).Render(" "))
//	}
//
//	lipgloss.Println(b.String())
func Blend(size int, points ...color.Color) []color.Color {
	if size <= 0 || len(points) < 2 {
		return nil
	}
	if size == 1 {
		return []color.Color{points[0]}
	}

	stops := make([]colorful.Color, len(points))
	for i, c := range points {
		stops[i], _ = colorful.MakeColor(c)
	}

	numSegments := len(stops) - 1
	blended := make([]color.Color, 0, size)

	// Calculate how many colors each segment should have.
	segmentSizes := make([]int, numSegments)
	baseSize := size / numSegments
	remainder := size % numSegments

	// Distribute the remainder across segments.
	for i := range numSegments {
		segmentSizes[i] = baseSize
		if i < remainder {
			segmentSizes[i]++
		}
	}

	// Generate colors for each segment.
	for i := range numSegments {
		c1 := stops[i]
		c2 := stops[i+1]
		segmentSize := segmentSizes[i]

		for j := range segmentSize {
			t := float64(j) / float64(segmentSize)
			c := c1.BlendHcl(c2, t)
			blended = append(blended, c)
		}
	}

	return blended
}
