package ansi

import (
	"strings"
	"unicode"
)

// OscSequence represents an OSC sequence.
//
// The sequence starts with a OSC sequence, OSC (0x9D) in a 8-bit environment
// or ESC ] (0x1B 0x5D) in a 7-bit environment, followed by positive integer identifier,
// then by arbitrary data terminated by a ST (0x9C) in a 8-bit environment,
// ESC \ (0x1B 0x5C) in a 7-bit environment, or BEL (0x07) for backwards compatibility.
//
//	OSC Ps ; Pt ST
//	OSC Ps ; Pt BEL
//
// See ECMA-48 § 5.7.
type OscSequence string

// IsValid reports whether the control sequence is valid.
// We allow UTF-8 in the data.
func (o OscSequence) IsValid() bool {
	if len(o) == 0 {
		return false
	}

	var i int
	if o[0] == OSC {
		i++
	} else if len(o) > 1 && o[0] == ESC && o[1] == ']' {
		i += 2
	} else {
		return false
	}

	// Osc data
	start := i
	end := -1
	for ; i < len(o) && o[i] >= 0x20 && o[i] <= 0xFF && o[i] != ST && o[i] != BEL && o[i] != ESC; i++ { // nolint: revive
		if end == -1 && o[i] == ';' {
			end = i
		}
	}
	if end == -1 {
		end = i
	}

	// Identifier must be all digits.
	for j := start; j < end; j++ {
		if !unicode.IsDigit(rune(o[j])) {
			return false
		}
	}

	// Terminator is one of the following:
	//  - ST (0x9C)
	//  - ESC \ (0x1B 0x5C)
	//  - BEL (0x07)
	return i < len(o) &&
		(o[i] == ST || o[i] == BEL || (i+1 < len(o) && o[i] == ESC && o[i+1] == '\\'))
}

// Identifier returns the identifier of the control sequence.
func (o OscSequence) Identifier() string {
	if len(o) == 0 {
		return ""
	}

	start := strings.IndexFunc(string(o), func(r rune) bool {
		return r >= '0' && r <= '9'
	})
	if start == -1 {
		return ""
	}
	end := strings.Index(string(o), ";")
	if end == -1 {
		for i := len(o) - 1; i > start; i-- {
			if o[i] == ST || o[i] == BEL || o[i] == ESC {
				end = i
				break
			}
		}
	}
	if end == -1 || start >= end {
		return ""
	}

	id := string(o[start:end])
	for _, r := range id {
		if !unicode.IsDigit(r) {
			return ""
		}
	}

	return id
}

// Data returns the data of the control sequence.
func (o OscSequence) Data() string {
	if len(o) == 0 {
		return ""
	}

	start := strings.Index(string(o), ";")
	if start == -1 {
		return ""
	}

	end := -1
	for i := len(o) - 1; i > start; i-- {
		if o[i] == ST || o[i] == BEL || o[i] == ESC {
			end = i
			break
		}
	}
	if end == -1 || start >= end {
		return ""
	}

	return string(o[start+1 : end])
}

// Terminator returns the terminator of the control sequence.
func (o OscSequence) Terminator() string {
	if len(o) == 0 {
		return ""
	}

	i := len(o) - 1
	for ; i > 0; i-- {
		if o[i] == ST || o[i] == BEL || o[i] == ESC {
			break
		}
	}
	if i == -1 {
		return ""
	}

	return string(o[i:])
}
