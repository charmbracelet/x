package ansi

import (
	"strings"
)

// CsiSequence represents a control sequence introducer (CSI) sequence.
//
// The sequence starts with a CSI sequence, CSI (0x9B) in a 8-bit environment
// or ESC [ (0x1B 0x5B) in a 7-bit environment, followed by any number of
// parameters in the range of 0x30-0x3F, then by any number of intermediate
// byte in the range of 0x20-0x2F, then finally with a single final byte in the
// range of 0x20-0x7E.
//
//	CSI P..P I..I F
//
// See ECMA-48 § 5.4.
type CsiSequence string

// IsValid reports whether the control sequence is valid.
func (c CsiSequence) IsValid() bool {
	if len(c) == 0 {
		return false
	}

	var i int
	if c[0] == CSI {
		i++
	} else if len(c) > 1 && c[0] == ESC && c[1] == '[' {
		i += 2
	} else {
		return false
	}

	// Parameters in the range 0x30-0x3F.
	for ; i < len(c) && c[i] >= 0x30 && c[i] <= 0x3F; i++ { // nolint: revive
	}

	// Intermediate bytes in the range 0x20-0x2F.
	for ; i < len(c) && c[i] >= 0x20 && c[i] <= 0x2F; i++ { // nolint: revive
	}

	// Final byte in the range 0x40-0x7E.
	return i < len(c) && c[i] >= 0x40 && c[i] <= 0x7E
}

// HasInitial reports whether the control sequence has an initial byte.
// This indicater a private sequence.
func (c CsiSequence) HasInitial() bool {
	i := c.Initial()
	return i >= 0x3C && i <= 0x3F
}

// Initial returns the initial byte of the control sequence.
func (c CsiSequence) Initial() byte {
	if len(c) == 0 {
		return 0
	}

	var i int
	for i = 0; i < len(c); i++ {
		if c[i] >= 0x30 && c[i] <= 0x3F {
			break
		}
	}

	if i >= len(c) {
		return 0
	}

	init := c[i]
	if init < 0x3C || init > 0x3F {
		return 0
	}

	return init
}

// Params returns the parameters of the control sequence.
func (c CsiSequence) Params() []byte {
	if len(c) == 0 {
		return []byte{}
	}

	start := strings.IndexFunc(string(c), func(r rune) bool {
		return r >= 0x30 && r <= 0x3F
	})
	if start == -1 {
		return []byte{}
	}

	end := strings.IndexFunc(string(c[start:]), func(r rune) bool {
		return r < 0x30 || r > 0x3F
	})
	if end == -1 {
		return []byte{}
	}

	return []byte(c[start : start+end])
}

// Intermediates returns the intermediate bytes of the control sequence.
func (c CsiSequence) Intermediates() []byte {
	if len(c) == 0 {
		return []byte{}
	}

	start := strings.IndexFunc(string(c), func(r rune) bool {
		return r >= 0x20 && r <= 0x2F
	})
	if start == -1 {
		return []byte{}
	}

	end := strings.IndexFunc(string(c[start:]), func(r rune) bool {
		return r < 0x20 || r > 0x2F
	})
	if end == -1 {
		return []byte{}
	}

	return []byte(c[start : start+end])
}

// Command returns the command byte of the control sequence.
// A CSI command byte is in the range of 0x40-0x7E. This includes ASCII
//   - @
//   - A-Z
//   - [ \ ]
//   - ^ _ `
//   - a-z
//   - { | }
//   - ~
func (c CsiSequence) Command() byte {
	i := strings.LastIndexFunc(string(c), func(r rune) bool {
		return r >= 0x40 && r <= 0x7E
	})
	if i == -1 {
		return 0
	}

	return c[i]
}

// IsPrivate reports whether the control sequence is a private sequence.
// This means either the first parameter byte is in the range of 0x3C-0x3F or
// the command byte is in the range of 0x70-0x7E.
func (c CsiSequence) IsPrivate() bool {
	if len(c) == 0 {
		return false
	}

	var i int
	for i = 0; i < len(c); i++ {
		if c[i] >= 0x30 && c[i] <= 0x3F {
			break
		}
	}

	return (c[i] >= 0x3C && c[i] <= 0x3F) ||
		(c[len(c)-1] >= 0x70 && c[len(c)-1] <= 0x7E)
}
