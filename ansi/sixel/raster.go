package sixel

import (
	"fmt"
	"io"
	"strings"
)

// ErrInvalidRaster is returned when Raster Attributes are invalid.
var ErrInvalidRaster = fmt.Errorf("invalid raster attributes")

// WriteRaster writes Raster attributes to a writer. If ph and pv are 0, they
// are omitted.
func WriteRaster(w io.Writer, pan, pad, ph, pv int) (n int, err error) {
	if pad == 0 {
		return WriteRaster(w, 1, 1, ph, pv)
	}

	if ph <= 0 && pv <= 0 {
		return fmt.Fprintf(w, "%c%d;%d", RasterAttribute, pan, pad) //nolint:wrapcheck
	}

	return fmt.Fprintf(w, "%c%d;%d;%d;%d", RasterAttribute, pan, pad, ph, pv) //nolint:wrapcheck
}

// Raster represents Sixel raster attributes.
type Raster struct {
	Pan, Pad, Ph, Pv int
}

// WriteTo writes Raster attributes to a writer.
func (r Raster) WriteTo(w io.Writer) (int64, error) {
	n, err := WriteRaster(w, r.Pan, r.Pad, r.Ph, r.Pv)
	return int64(n), err
}

// String returns the Raster as a string.
func (r Raster) String() string {
	var b strings.Builder
	r.WriteTo(&b) //nolint:errcheck,gosec
	return b.String()
}

// DecodeRaster decodes a Raster from a byte slice. It returns the Raster and
// the number of bytes read.
func DecodeRaster(data []byte) (r Raster, n int) {
	if len(data) == 0 || data[0] != RasterAttribute {
		return r, n
	}

	ptr := &r.Pan
	for n = 1; n < len(data); n++ {
		if data[n] == ';' { //nolint:nestif
			if ptr == &r.Pan {
				ptr = &r.Pad
			} else if ptr == &r.Pad {
				ptr = &r.Ph
			} else if ptr == &r.Ph {
				ptr = &r.Pv
			} else {
				n++
				break
			}
		} else if data[n] >= '0' && data[n] <= '9' {
			*ptr = (*ptr)*10 + int(data[n]-'0')
		} else {
			break
		}
	}

	return r, n
}
