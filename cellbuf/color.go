package cellbuf

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
)

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
			R: uint8(ansi.Param(params[i+2])),
			G: uint8(ansi.Param(params[i+3])),
			B: uint8(ansi.Param(params[i+4])),
			A: 0xff,
		}
		*idxp += 4
	case 5: // 256 colors
		if i > paramsLen-2 {
			return
		}
		c = ansi.ExtendedColor(ansi.Param(params[i+2]))
		*idxp += 2
	}
	return
}
