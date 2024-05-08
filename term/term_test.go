package term_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/charmbracelet/x/term"
)

func TestIsTerminalTempFile(t *testing.T) {
	file, err := os.CreateTemp("", "TestIsTerminalTempFile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if term.IsTerminal(file.Fd()) {
		t.Fatalf("IsTerminal unexpectedly returned true for temporary file %s", file.Name())
	}
}

func TestIsTerminalTerm(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
	}
	file, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if !term.IsTerminal(file.Fd()) {
		t.Fatalf("IsTerminal unexpectedly returned false for terminal file %s", file.Name())
	}
}
