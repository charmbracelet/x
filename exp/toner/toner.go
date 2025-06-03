package toner

import (
	"bytes"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/charmtone"
)

// Strings returns a colorized string representation of the input byte slice or
// string.
func Strings(s string) string {
	var buf bytes.Buffer
	_, _ = writeColorize(&buf, s)
	return buf.String()
}

// Bytes returns a colorized byte slice representation of the input byte slice.
func Bytes(b []byte) []byte {
	var buf bytes.Buffer
	_, _ = writeColorize(&buf, b)
	return buf.Bytes()
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

	return writeColorize(w, p)
}

// WriteString writes the colorized output of the input string to the underlying writer.
func (w Writer) WriteString(s string) (n int, err error) {
	if len(s) == 0 {
		return 0, nil
	}

	return writeColorize(w, s)
}

var colors = func() []charmtone.Key {
	cols := charmtone.Keys()
	// Filter out these colors.
	filterOut := []charmtone.Key{
		charmtone.Cumin,
		charmtone.Tang,
		charmtone.Paprika,
		charmtone.Pepper,
		charmtone.Charcoal,
		charmtone.Iron,
		charmtone.Oyster,
		charmtone.Squid,
		charmtone.Smoke,
		charmtone.Ash,
		charmtone.Salt,
		charmtone.Butter,
	}
	for _, k := range filterOut {
		cols = slices.DeleteFunc(cols, func(c charmtone.Key) bool {
			return c == k
		})
	}
	return cols
}()

func writeColorize[T []byte | string](w io.Writer, p T) (n int, err error) {
	pa := ansi.NewParser()

	var state byte
	for len(p) > 0 {
		seq, width, nr, newState := ansi.DecodeSequence(p, state, pa)
		cmd := pa.Command()

		var st ansi.Style
		var s string
		if cmd == 0 && width > 0 {
			s = string(seq)
		} else {
			st = st.ForegroundColor(colors[cmd%len(colors)])
			s = string(seq)
		}

		s = strconv.Quote(s)
		s = strings.TrimPrefix(s, "\"")
		s = strings.TrimSuffix(s, "\"")
		if len(st) > 0 {
			s = st.Styled(s)
		}

		m, err := io.WriteString(w, s)
		if err != nil {
			return n, err
		}

		n += m

		p = p[nr:]
		state = newState
	}

	return n, nil
}
