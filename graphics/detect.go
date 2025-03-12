package graphics

import (
	"bytes"
	"os"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
)

// TODO: Verify if it's running with tmux for Kitty and ITerm2
// TODO: Write a func for preferred protocol for the terminal
// TODO: Additional check if the terminal supports cell-size by `[16t` or `[14t`
// TODO: Write tests by mocking a terminal context ?

const (
	termProgramVariable = "TERM_PROGRAM"
	lcTerminalVariable  = "LC_TERMINAL"
)

// Returns the availability of each image protocol.
type ImageProtocols struct {
	Sixel  bool
	ITerm2 bool
	Kitty  bool
	// Mosaic (Halfblocks) should work in all terminals,
	// even if the font size could not be detected, with a 4:8 pixel ratio.
	Mosaic bool
}

// Detect all availables image protocols and return as [ImageProtocols].
func DetectImageProtocols() ImageProtocols {
	return ImageProtocols{
		Sixel: detectSixel(),
		// TODO: `_Gi=...`: Kitty graphics support.
		Kitty: detectKitty(),
		// TODO: `[1337n`: iTerm2 (some terminals implement the protocol but sadly not this custom CSI)
		ITerm2: detectIterm2() || detectIterm2FromEnv(),
		Mosaic: true,
	}
}

func detectKitty() bool {
	return false
}

func detectIterm2() bool {
	return false
}

// This function detects iTerm2 protocol from environment variable.
func detectIterm2FromEnv() bool {
	termProgram := os.Getenv(termProgramVariable)
	if termProgram == "iTerm" ||
		termProgram == "WezTerm" ||
		termProgram == "mintty" ||
		termProgram == "vscode" ||
		termProgram == "Tabby" ||
		termProgram == "Hyper" ||
		termProgram == "rio" {
		return true
	}

	lcTerminal := os.Getenv(lcTerminalVariable)
	return lcTerminal == "iTerm"
}

func detectSixel() bool {
	sixelSupportedTerminals := []string{
		"\x1b[?62;", // VT240
		"\x1b[?63;", // wsltty
		"\x1b[?64;", // mintty
		"\x1b[?65;", // RLogin
		// NOTE: tmux does not return VT name.
		"\x1b[?1;2;4c", // Tmux
	}

	if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return true
	}
	s, err := term.MakeRaw(1)
	if err == nil {
		defer term.Restore(1, s) // nolint:errcheck
	}
	_, err = os.Stdout.Write([]byte("\x1b[c"))
	if err != nil {
		return false
	}
	defer os.Stdout.SetReadDeadline(time.Time{}) // nolint:errcheck

	var b [100]byte
	n, err := os.Stdout.Read(b[:])
	if err != nil {
		return false
	}

	for _, t := range bytes.Split(b[6:n], []byte(";")) {
		// Check if 4 is present in terminal capabilities.
		if len(t) == 1 && t[0] == '4' {
			return true
		}
	}

	for _, supportedTerminal := range sixelSupportedTerminals {
		if bytes.HasPrefix(b[:n], []byte(supportedTerminal)) {
			return true
		}
	}

	return false
}
