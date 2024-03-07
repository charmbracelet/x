package term

import (
	"image/color"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

// Environ represents the terminal environment.
type Environ interface {
	Getenv(string) string
	LookupEnv(string) (string, bool)
	Environ() []string
}

// OsEnviron is an implementation of Environ that uses os.Environ.
type OsEnviron struct{}

var _ Environ = OsEnviron{}

// Environ implements Environ.
func (OsEnviron) Environ() []string {
	return os.Environ()
}

// Getenv implements Environ.
func (OsEnviron) Getenv(key string) string {
	return os.Getenv(key)
}

// LookupEnv implements Environ.
func (OsEnviron) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// BackgroundColor queries the terminal for the background color.
// If the terminal does not support querying the background color, nil is
// returned.
func BackgroundColor(in, out *os.File) (c color.Color) {
	// nolint: errcheck
	queryTerminal(in, out, func(events []input.Event) bool {
		for _, e := range events {
			switch e := e.(type) {
			case input.BackgroundColorEvent:
				c = e.Color
				continue // we need to consume the next DA1 event
			case input.PrimaryDeviceAttributesEvent:
				return false
			}
		}
		return true
	}, ansi.RequestBackgroundColor+ansi.RequestPrimaryDeviceAttributes)
	return
}

// ForegroundColor queries the terminal for the foreground color.
// If the terminal does not support querying the foreground color, nil is
// returned.
func ForegroundColor(in, out *os.File) (c color.Color) {
	// nolint: errcheck
	queryTerminal(in, out, func(events []input.Event) bool {
		for _, e := range events {
			switch e := e.(type) {
			case input.ForegroundColorEvent:
				c = e.Color
				continue // we need to consume the next DA1 event
			case input.PrimaryDeviceAttributesEvent:
				return false
			}
		}
		return true
	}, ansi.RequestForegroundColor+ansi.RequestPrimaryDeviceAttributes)
	return
}

// CursorColor queries the terminal for the cursor color.
// If the terminal does not support querying the cursor color, nil is returned.
func CursorColor(in, out *os.File) (c color.Color) {
	// nolint: errcheck
	queryTerminal(in, out, func(events []input.Event) bool {
		for _, e := range events {
			switch e := e.(type) {
			case input.CursorColorEvent:
				c = e.Color
				continue // we need to consume the next DA1 event
			case input.PrimaryDeviceAttributesEvent:
				return false
			}
		}
		return true
	}, ansi.RequestCursorColor+ansi.RequestPrimaryDeviceAttributes)
	return
}

// SupportsKittyKeyboard returns true if the terminal supports the Kitty
// keyboard protocol.
func SupportsKittyKeyboard(in, out *os.File) (supported bool) {
	// nolint: errcheck
	queryTerminal(in, out, func(events []input.Event) bool {
		for _, e := range events {
			switch e.(type) {
			case input.KittyKeyboardEvent:
				supported = true
				continue // we need to consume the next DA1 event
			case input.PrimaryDeviceAttributesEvent:
				return false
			}
		}
		return true
	}, ansi.RequestKittyKeyboard+ansi.RequestPrimaryDeviceAttributes)
	return
}

// queryTerminalFunc is a function that filters input events using a type
// switch. If false is returned, the queryTerminal function will stop reading
// input.
type queryTerminalFunc func(events []input.Event) bool

// queryTerminal queries the terminal for support of various features and
// returns a list of response events.
func queryTerminal(
	in *os.File,
	out *os.File,
	filter queryTerminalFunc,
	query string,
) error {
	state, err := MakeRaw(in.Fd())
	if err != nil {
		return err
	}

	defer Restore(in.Fd(), state) // nolint: errcheck

	rd, err := input.NewDriver(in, "", 0)
	if err != nil {
		return err
	}

	defer rd.Close() // nolint: errcheck

	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			rd.Cancel()
		}
	}()

	if _, err := io.WriteString(out, query); err != nil {
		return err
	}

	events := make([]input.Event, 2)

	for {
		n, err := rd.ReadInput(events)
		if err != nil {
			return err
		}

		if !filter(events[:n]) {
			break
		}
	}

	return nil
}
