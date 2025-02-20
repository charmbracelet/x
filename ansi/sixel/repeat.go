package sixel

import (
	"fmt"
	"io"
	"strings"
)

// ErrInvalidRepeat is returned when a Repeat is invalid
var ErrInvalidRepeat = fmt.Errorf("invalid repeat")

// WriteRepeat writes a Repeat to a writer. A repeat character is in the range
// of '?' (0x3F) to '~' (0x7E).
func WriteRepeat(w io.Writer, count int, char byte) (int, error) {
	return fmt.Fprintf(w, "%c%d%c", RepeatIntroducer, count, char)
}

// Repeat represents a Sixel repeat introducer.
type Repeat struct {
	Count int
	Char  byte
}

// WriteTo writes a Repeat to a writer.
func (r Repeat) WriteTo(w io.Writer) (int64, error) {
	n, err := WriteRepeat(w, r.Count, r.Char)
	return int64(n), err
}

// String returns the Repeat as a string.
func (r Repeat) String() string {
	var b strings.Builder
	r.WriteTo(&b) //nolint:errcheck
	return b.String()
}

// DecodeRepeat decodes a Repeat from a byte slice. It returns the Repeat and
// the number of bytes read.
func DecodeRepeat(data []byte) (r Repeat, n int) {
	if len(data) == 0 || data[0] != RepeatIntroducer {
		return
	}

	if len(data) < 3 { // The minimum length is 3: the introducer, a digit, and a character.
		return
	}

	for n = 1; n < len(data); n++ {
		if data[n] >= '0' && data[n] <= '9' {
			r.Count = r.Count*10 + int(data[n]-'0')
		} else {
			r.Char = data[n]
			n++ // Include the character in the count.
			break
		}
	}

	return
}
