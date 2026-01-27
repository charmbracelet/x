package ansi

import (
	"strconv"
	"strings"
)

// ModeSetting represents a mode setting.
type ModeSetting byte

// ModeSetting constants.
const (
	ModeNotRecognized ModeSetting = iota
	ModeSet
	ModeReset
	ModePermanentlySet
	ModePermanentlyReset
)

// IsNotRecognized returns true if the mode is not recognized.
func (m ModeSetting) IsNotRecognized() bool {
	return m == ModeNotRecognized
}

// IsSet returns true if the mode is set or permanently set.
func (m ModeSetting) IsSet() bool {
	return m == ModeSet || m == ModePermanentlySet
}

// IsReset returns true if the mode is reset or permanently reset.
func (m ModeSetting) IsReset() bool {
	return m == ModeReset || m == ModePermanentlyReset
}

// IsPermanentlySet returns true if the mode is permanently set.
func (m ModeSetting) IsPermanentlySet() bool {
	return m == ModePermanentlySet
}

// IsPermanentlyReset returns true if the mode is permanently reset.
func (m ModeSetting) IsPermanentlyReset() bool {
	return m == ModePermanentlyReset
}

// Mode represents an interface for terminal modes.
// Modes can be set, reset, and requested.
type Mode interface {
	Mode() int
}

// SetMode (SM) or (DECSET) returns a sequence to set a mode.
// The mode arguments are a list of modes to set.
//
// If one of the modes is a [DECMode], the function will returns two escape
// sequences.
//
// ANSI format:
//
//	CSI Pd ; ... ; Pd h
//
// DEC format:
//
//	CSI ? Pd ; ... ; Pd h
//
// See: https://vt100.net/docs/vt510-rm/SM.html
func SetMode(modes ...Mode) string {
	return setMode(false, modes...)
}

// SM is an alias for [SetMode].
func SM(modes ...Mode) string {
	return SetMode(modes...)
}

// DECSET is an alias for [SetMode].
func DECSET(modes ...Mode) string {
	return SetMode(modes...)
}

// ResetMode (RM) or (DECRST) returns a sequence to reset a mode.
// The mode arguments are a list of modes to reset.
//
// If one of the modes is a [DECMode], the function will returns two escape
// sequences.
//
// ANSI format:
//
//	CSI Pd ; ... ; Pd l
//
// DEC format:
//
//	CSI ? Pd ; ... ; Pd l
//
// See: https://vt100.net/docs/vt510-rm/RM.html
func ResetMode(modes ...Mode) string {
	return setMode(true, modes...)
}

// RM is an alias for [ResetMode].
func RM(modes ...Mode) string {
	return ResetMode(modes...)
}

// DECRST is an alias for [ResetMode].
func DECRST(modes ...Mode) string {
	return ResetMode(modes...)
}

func setMode(reset bool, modes ...Mode) (s string) {
	if len(modes) == 0 {
		return s
	}

	cmd := "h"
	if reset {
		cmd = "l"
	}

	seq := "\x1b["
	if len(modes) == 1 {
		switch modes[0].(type) {
		case DECMode:
			seq += "?"
		}
		return seq + strconv.Itoa(modes[0].Mode()) + cmd
	}

	dec := make([]string, 0, len(modes)/2)
	ansi := make([]string, 0, len(modes)/2)
	for _, m := range modes {
		switch m.(type) {
		case DECMode:
			dec = append(dec, strconv.Itoa(m.Mode()))
		case ANSIMode:
			ansi = append(ansi, strconv.Itoa(m.Mode()))
		}
	}

	if len(ansi) > 0 {
		s += seq + strings.Join(ansi, ";") + cmd
	}
	if len(dec) > 0 {
		s += seq + "?" + strings.Join(dec, ";") + cmd
	}
	return s
}

// RequestMode (DECRQM) returns a sequence to request a mode from the terminal.
// The terminal responds with a report mode function [DECRPM].
//
// ANSI format:
//
//	CSI Pa $ p
//
// DEC format:
//
//	CSI ? Pa $ p
//
// See: https://vt100.net/docs/vt510-rm/DECRQM.html
func RequestMode(m Mode) string {
	seq := "\x1b["
	switch m.(type) {
	case DECMode:
		seq += "?"
	}
	return seq + strconv.Itoa(m.Mode()) + "$p"
}

// DECRQM is an alias for [RequestMode].
func DECRQM(m Mode) string {
	return RequestMode(m)
}

// ReportMode (DECRPM) returns a sequence that the terminal sends to the host
// in response to a mode request [DECRQM].
//
// ANSI format:
//
//	CSI Pa ; Ps ; $ y
//
// DEC format:
//
//	CSI ? Pa ; Ps $ y
//
// Where Pa is the mode number, and Ps is the mode value.
//
//	0: Not recognized
//	1: Set
//	2: Reset
//	3: Permanent set
//	4: Permanent reset
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
func ReportMode(mode Mode, value ModeSetting) string {
	if value > 4 {
		value = 0
	}
	switch mode.(type) {
	case DECMode:
		return "\x1b[?" + strconv.Itoa(mode.Mode()) + ";" + strconv.Itoa(int(value)) + "$y"
	}
	return "\x1b[" + strconv.Itoa(mode.Mode()) + ";" + strconv.Itoa(int(value)) + "$y"
}

// DECRPM is an alias for [ReportMode].
func DECRPM(mode Mode, value ModeSetting) string {
	return ReportMode(mode, value)
}

// ANSIMode represents an ANSI terminal mode.
type ANSIMode int //nolint:revive

// Mode returns the ANSI mode as an integer.
func (m ANSIMode) Mode() int {
	return int(m)
}

// DECMode represents a private DEC terminal mode.
type DECMode int

// Mode returns the DEC mode as an integer.
func (m DECMode) Mode() int {
	return int(m)
}


// These are aliases for [SetModeTextCursorEnable] and [ResetModeTextCursorEnable].
const (
	ShowCursor = SetModeTextCursorEnable
	HideCursor = ResetModeTextCursorEnable
)
