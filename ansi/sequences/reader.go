package sequences

import (
	"bufio"
	"io"

	"github.com/charmbracelet/x/ansi"
)

// Scanner is a [bufio.Scanner] wrapped to use ANSI escape sequence splitting.
type Scanner struct {
	*bufio.Scanner
	state state[[]byte]
}

// FromReader returns a new [Scanner] that splits ANSI escape sequences and
// grapheme clusters.
//
// It embeds a [bufio.Scanner], so you can use [Scanner.Scan] and
// [Scanner.Text] as usual.
func FromReader(r io.Reader) *Scanner {
	s := new(Scanner)
	s.state = newState[[]byte]()
	s.Scanner = bufio.NewScanner(r)
	s.Scanner.Split(s.state.splitFunc)
	return s
}

// Width returns the display width of the most recently scanned token.
func (s *Scanner) Width() int {
	return s.state.width
}

// State returns the current state of the scanner.
func (s *Scanner) State() ansi.State {
	return s.state.state
}

// Exec sets the control code callback function that is called when a control
// character is encountered.
func (s *Scanner) Exec(f ExecFunc) {
	s.state.execFunc = f
}

// Cmd sets the command callback function that is called when an escape
// sequence is finished.
func (s *Scanner) Cmd(f CmdFunc) {
	s.state.cmdFunc = f
}

// Param sets the parameter callback function that is called when a parameter
// is parsed.
func (s *Scanner) Param(f ParamFunc) {
	s.state.paramFunc = f
}

// Data sets the data callback function that is called when data is parsed in
// string sequences.
func (s *Scanner) Data(f DataFunc[[]byte]) {
	s.state.dataFunc = f
}

// Print sets the grapheme cluster callback function that is called when a
// grapheme cluster is parsed.
func (s *Scanner) Print(f PrintFunc[[]byte]) {
	s.state.printFunc = f
}
