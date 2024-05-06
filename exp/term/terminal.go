package term

import (
	"image/color"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

// BackgroundColor queries the terminal for the background color.
// If the terminal does not support querying the background color, nil is
// returned.
func BackgroundColor(in, out *os.File) (c color.Color) {
	state, err := MakeRaw(in.Fd())
	if err != nil {
		return
	}

	defer Restore(in.Fd(), state) // nolint: errcheck

	// nolint: errcheck
	QueryTerminal(in, out, func(events []input.Event) bool {
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
	state, err := MakeRaw(in.Fd())
	if err != nil {
		return
	}

	defer Restore(in.Fd(), state) // nolint: errcheck

	// nolint: errcheck
	QueryTerminal(in, out, func(events []input.Event) bool {
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
	state, err := MakeRaw(in.Fd())
	if err != nil {
		return
	}

	defer Restore(in.Fd(), state) // nolint: errcheck

	// nolint: errcheck
	QueryTerminal(in, out, func(events []input.Event) bool {
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
	state, err := MakeRaw(in.Fd())
	if err != nil {
		return
	}

	defer Restore(in.Fd(), state) // nolint: errcheck

	// nolint: errcheck
	QueryTerminal(in, out, func(events []input.Event) bool {
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

// QueryTerminalFilter is a function that filters input events using a type
// switch. If false is returned, the QueryTerminal function will stop reading
// input.
type QueryTerminalFilter func(events []input.Event) bool

// QueryTerminal queries the terminal for support of various features and
// returns a list of response events.
// Most of the time, you will need to set stdin to raw mode before calling this
// function.
// Note: This function will block until the terminal responds or the timeout
// is reached.
func QueryTerminal(
	in io.Reader,
	out io.Writer,
	filter QueryTerminalFilter,
	query string,
) error {
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

	for {
		events, err := rd.ReadEvents()
		if err != nil {
			return err
		}

		if !filter(events) {
			break
		}
	}

	return nil
}
