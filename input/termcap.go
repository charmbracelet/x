package input

import (
	"bytes"
	"encoding/hex"
)

// TermcapEvent represents a Termcap response event. Termcap responses are
// generated by the terminal in response to RequestTermcap (XTGETTCAP)
// requests.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
type TermcapEvent struct {
	Values  map[string]string
	IsValid bool
}

func parseTermcap(data []byte) TermcapEvent {
	// XTGETTCAP
	if len(data) == 0 {
		return TermcapEvent{}
	}

	tc := TermcapEvent{Values: make(map[string]string)}
	split := bytes.Split(data, []byte{';'})
	for _, s := range split {
		parts := bytes.SplitN(s, []byte{'='}, 2)
		if len(parts) == 0 {
			return TermcapEvent{}
		}

		name, err := hex.DecodeString(string(parts[0]))
		if err != nil || len(name) == 0 {
			continue
		}

		var value []byte
		if len(parts) > 1 {
			value, err = hex.DecodeString(string(parts[1]))
			if err != nil {
				continue
			}
		}

		tc.Values[string(name)] = string(value)
	}

	return tc
}
