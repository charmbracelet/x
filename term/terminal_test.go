package term_test

import (
	"os"
	"testing"

	"github.com/charmbracelet/x/term"
)

func TestTerminalQueries(t *testing.T) {
	in, out := os.Stdin, os.Stdout
	term.QueryBackgroundColor(in, out)
	term.QueryKittyKeyboard(in, out)
}
