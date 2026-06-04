package parser

import "testing"

// TestStringState_C1ST_NotDispatched verifies that 0x9C (C1 ST) inside any of
// the string-typed states (OSC / DCS / SOS / PM / APC) does NOT terminate the
// sequence and instead remains part of the string payload.
//
// 0x9C also happens to be a valid UTF-8 continuation byte — for example U+2733
// (✳) encodes as 0xE2 0x9C 0xB3, and U+672B (末) encodes as 0xE6 0x9C 0xAB —
// so treating it as a terminator splits UTF-8 string payloads (window titles,
// hyperlink IDs, DCS data, …) in the middle of a multi-byte character. The
// fix re-registers 0x9C as PutAction inside every string state via the
// AddRange(0x20..0xFF, …, PutAction, …) entry; callers terminate with the
// 7-bit ST (ESC \\) or BEL instead.
func TestStringState_C1ST_NotDispatched(t *testing.T) {
	cases := []struct {
		name  string
		state State
	}{
		{"Osc", OscStringState},
		{"Dcs", DcsStringState},
		{"Sos", SosStringState},
		{"Pm", PmStringState},
		{"Apc", ApcStringState},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next, act := Table.Transition(tc.state, 0x9C)
			if act != PutAction {
				t.Errorf("0x9C action in %s state: got %s, want PutAction",
					tc.name, ActionNames[act])
			}
			if next != tc.state {
				t.Errorf("0x9C next state from %s: got %s, want %s",
					tc.name, StateNames[next], StateNames[tc.state])
			}
		})
	}
}

// TestStringState_Utf8LeadBytes_StayInState verifies that UTF-8 lead bytes
// (0xC2..0xF4) inside string-typed states remain in the same state (PutAction)
// rather than transitioning into Utf8State.
//
// The Anywhere block of GenerateTransitionTable installs
// "0xC2..0xF4 -> CollectAction + Utf8State" for every state including the
// string states. Utf8State completes the rune via PrintAction and returns to
// GroundState — abandoning the string mid-payload and drawing the rest to the
// grid. String states override this with AddRange(0x20..0xFF, …, PutAction,
// …) (or 0x80..0xFF for DCS / SOS / PM / APC, after this commit) so that the
// entire UTF-8 byte sequence is collected as opaque payload.
func TestStringState_Utf8LeadBytes_StayInState(t *testing.T) {
	leadBytes := []byte{0xC2, 0xDF, 0xE0, 0xEF, 0xF0, 0xF4}
	states := []struct {
		name  string
		state State
	}{
		{"Osc", OscStringState},
		{"Dcs", DcsStringState},
		{"Sos", SosStringState},
		{"Pm", PmStringState},
		{"Apc", ApcStringState},
	}
	for _, tc := range states {
		for _, b := range leadBytes {
			b := b
			t.Run(tc.name+"_0x"+hex(b), func(t *testing.T) {
				next, act := Table.Transition(tc.state, b)
				if act != PutAction {
					t.Errorf("0x%02X action in %s state: got %s, want PutAction",
						b, tc.name, ActionNames[act])
				}
				if next != tc.state {
					t.Errorf("0x%02X next state from %s: got %s, want %s",
						b, tc.name, StateNames[next], StateNames[tc.state])
				}
			})
		}
	}
}

// TestStringState_7BitTerminators verifies the supported string terminators
// still fire: ESC (\x1B) for the 7-bit ST prefix and BEL (\x07) for OSC.
// These are the codepoints UTF-8 terminals should rely on after the 8-bit C1
// ST drop.
func TestStringState_7BitTerminators(t *testing.T) {
	// ESC transitions every string state into EscapeState; the following \\
	// then completes the 7-bit ST in EscapeState. DcsStringState behaves the
	// same way (it's DcsEntryState that accepts ESC as PutAction for
	// passthrough — once we're already in DcsStringState the ESC terminates
	// the string just like in OSC/SOS/PM/APC).
	escStates := []struct {
		name  string
		state State
	}{
		{"Osc", OscStringState},
		{"Dcs", DcsStringState},
		{"Sos", SosStringState},
		{"Pm", PmStringState},
		{"Apc", ApcStringState},
	}
	for _, tc := range escStates {
		t.Run("ESC_"+tc.name, func(t *testing.T) {
			next, act := Table.Transition(tc.state, 0x1B)
			if act != DispatchAction {
				t.Errorf("ESC action in %s state: got %s, want DispatchAction",
					tc.name, ActionNames[act])
			}
			if next != EscapeState {
				t.Errorf("ESC next state from %s: got %s, want EscapeState",
					tc.name, StateNames[next])
			}
		})
	}
	// BEL terminates OSC (only OSC, not DCS / SOS / PM / APC).
	t.Run("BEL_Osc", func(t *testing.T) {
		next, act := Table.Transition(OscStringState, 0x07)
		if act != DispatchAction {
			t.Errorf("BEL action in OscStringState: got %s, want DispatchAction",
				ActionNames[act])
		}
		if next != GroundState {
			t.Errorf("BEL next state from OscStringState: got %s, want GroundState",
				StateNames[next])
		}
	})
}

func hex(b byte) string {
	const hexdigits = "0123456789ABCDEF"
	return string([]byte{hexdigits[b>>4], hexdigits[b&0xF]})
}
