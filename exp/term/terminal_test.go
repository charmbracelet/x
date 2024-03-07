package term_test

import (
	"os"
	"testing"

	"github.com/charmbracelet/x/exp/term"
)

func TestTerminalQueries(t *testing.T) {
	in, out := os.Stdin, os.Stdout
	_ = term.BackgroundColor(in, out)
	_ = term.ForegroundColor(in, out)
	_ = term.CursorColor(in, out)
	_ = term.SupportsKittyKeyboard(in, out)
}
