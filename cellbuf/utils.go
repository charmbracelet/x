package cellbuf

import (
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Height returns the height of a string.
func Height(s string) int {
	return strings.Count(s, "\n") + 1
}

func readColor(idxp *int, params []int) (c ansi.Color) {
	i := *idxp
	paramsLen := len(params)
	if i > paramsLen-1 {
		return
	}
	// Note: we accept both main and subparams here
	switch param := ansi.Param(params[i+1]); param {
	case 2: // RGB
		if i > paramsLen-4 {
			return
		}
		c = color.RGBA{
			R: uint8(ansi.Param(params[i+2])), //nolint:gosec
			G: uint8(ansi.Param(params[i+3])), //nolint:gosec
			B: uint8(ansi.Param(params[i+4])), //nolint:gosec
			A: 0xff,
		}
		*idxp += 4
	case 5: // 256 colors
		if i > paramsLen-2 {
			return
		}
		c = ansi.ExtendedColor(ansi.Param(params[i+2])) //nolint:gosec
		*idxp += 2
	}
	return
}
