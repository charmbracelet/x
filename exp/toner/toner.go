package toner

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
)

// Strings returns a colorized string representation of the input byte slice or
// string.
func Strings(s string) string {
	return colorize(s)
}

// Bytes returns a colorized byte slice representation of the input byte slice.
func Bytes(b []byte) []byte {
	return colorize(b)
}

// Writer encapsulates a [io.Writer] that colorizes the output using charm tones.
type Writer struct {
	io.Writer
}

// Write writes the colorized output to the underlying writer.
func (w Writer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	colored := colorize(p)
	n, err = w.Writer.Write(colored)
	return n, err
}

// WriteString writes the colorized output of the input string to the underlying writer.
func (w Writer) WriteString(s string) (n int, err error) {
	if len(s) == 0 {
		return 0, nil
	}

	colored := colorize(s)
	n, err = io.WriteString(w.Writer, colored)
	return n, err
}

const (
	startTone = charmtone.Cumin
	endTone   = charmtone.Zest
)

func colorize[T []byte | string](b T) T {
	var buf bytes.Buffer

	p := ansi.NewParser()

	var state byte
	for len(b) > 0 {
		seq, w, n, newState := ansi.DecodeSequence(b, state, p)
		cmd := p.Command()

		var st ansi.Style
		var s string
		if cmd == 0 && w > 0 {
			s = string(seq)
		} else {
			st = st.ForegroundColor(charmtone.Key(cmd % int(endTone)))
			s = string(seq)
		}

		s = strconv.Quote(s)
		s = strings.TrimPrefix(s, "\"")
		s = strings.TrimSuffix(s, "\"")
		if len(st) > 0 {
			s = st.Styled(s)
		}
		buf.WriteString(s)

		b = b[n:]
		state = newState
	}

	switch any(b).(type) {
	case []byte:
		return T(buf.Bytes())
	default:
		return T(buf.String())
	}
}
