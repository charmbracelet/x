package input

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/term/ansi"
)

var sequences = registerKeys(FlagNoTerminfo, "dumb")

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		k := KeyDownEvent{Sym: KeySpace, Rune: ' ', Mod: Alt}
		if got := k.String(); got != "alt+space" {
			t.Fatalf(`expected a "alt+space ", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		k := KeyDownEvent{Rune: 'a'}
		if got := k.String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		k := KeyDownEvent{Sym: 99999}
		if got := k.String(); got != "unknown" {
			t.Fatalf(`expected a "unknown", got %q`, got)
		}
	})
}

func TestKeyTypeString(t *testing.T) {
	t.Run("space", func(t *testing.T) {
		if got := KeySpace.String(); got != "space" {
			t.Fatalf(`expected a "space", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeySym(99999).String(); got != "unknown" {
			t.Fatalf(`expected a "unknown", got %q`, got)
		}
	})
}

type seqTest struct {
	seq  []byte
	msgs []Event
}

// buildBaseSeqTests returns sequence tests that are valid for the
// detectSequence() function.
func buildBaseSeqTests() []seqTest {
	td := []seqTest{}
	for seq, key := range sequences {
		td = append(td, seqTest{[]byte(seq), []Event{KeyDownEvent(key)}})
	}

	// Additional special cases.
	td = append(td,
		// Unrecognized CSI sequence.
		seqTest{
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Event{
				UnknownCsiEvent([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'}),
			},
		},
		// A lone space character.
		seqTest{
			[]byte{' '},
			[]Event{
				KeyDownEvent{Sym: KeySpace, Rune: ' '},
			},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			[]Event{
				KeyDownEvent{Sym: KeySpace, Rune: ' ', Mod: Alt},
			},
		},
	)
	return td
}

func TestDetectSequence(t *testing.T) {
	td := buildBaseSeqTests()
	parser := NewEventParser("dumb", FlagNoTerminfo)
	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Event
			buf := tc.seq
			for len(buf) > 0 {
				width, msg := parser.ParseSequence(buf)
				events = append(events, msg)
				buf = buf[width:]
			}
			if !reflect.DeepEqual(tc.msgs, events) {
				t.Errorf("\nexpected event:\n    %#v\ngot:\n    %#v", tc.msgs, events)
			}
		})
	}
}

func TestDetectOneEvent(t *testing.T) {
	t.Skip("WIP")
	td := buildBaseSeqTests()
	// Add tests for the inputs that detectOneEvent() can parse, but
	// detectSequence() cannot.
	td = append(td,
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Event{
				MouseDownEvent{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		// SGR Mouse event.
		seqTest{
			[]byte("\x1b[<0;33;17M"),
			[]Event{
				MouseDownEvent{X: 32, Y: 16, Button: MouseLeft},
			},
		},
		// Runes.
		seqTest{
			[]byte{'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
			},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a', Mod: Alt},
			},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
			},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: '☃'},
			},
		},
		seqTest{
			[]byte("\x1b☃"),
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: '☃', Mod: Alt},
			},
		},
		// Standalone control chacters.
		seqTest{
			[]byte{'\x1b'},
			[]Event{
				KeyDownEvent{Sym: KeyEscape},
			},
		},
		seqTest{
			[]byte{byte(ansi.SOH)},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl},
			},
		},
		seqTest{
			[]byte{'\x1b', byte(ansi.SOH)},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl | Alt},
			},
		},
		seqTest{
			[]byte{byte(ansi.NUL)},
			[]Event{
				KeyDownEvent{Rune: '@', Mod: Ctrl},
			},
		},
		seqTest{
			[]byte{'\x1b', byte(ansi.NUL)},
			[]Event{
				KeyDownEvent{Rune: '@', Mod: Ctrl | Alt},
			},
		},
		// Invalid characters.
		seqTest{
			[]byte{'\x80'},
			[]Event{
				UnknownEvent(byte(0x80)),
			},
		},
	)

	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			[]Event{
				UnknownEvent(byte(0xfe)),
			},
		})
	}

	parser := NewEventParser("dumb", 0)
	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Event
			buf := tc.seq
			for len(buf) > 0 {
				width, msg := parser.ParseSequence(buf)
				events = append(events, msg)
				buf = buf[width:]
			}
			if !reflect.DeepEqual(tc.msgs, events) {
				t.Errorf("expected event %#v (%T), got %#v (%T)", tc.msgs, tc.msgs, events, events)
			}
		})
	}
}

func TestReadLongInput(t *testing.T) {
	expect := make([]Event, 1000)
	for i := 0; i < 1000; i++ {
		expect[i] = KeyDownEvent{Rune: 'a'}
	}
	input := strings.Repeat("a", 1000)
	drv, err := NewDriver(strings.NewReader(input), "", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	var buf [maxBufferSize]Event
	var msgs []Event
	for {
		n, err := drv.ReadInput(buf[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected input error: %v", err)
		}
		msgs = append(msgs, buf[:n]...)
	}

	if !reflect.DeepEqual(expect, msgs) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, msgs)
	}
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname string
		in      []byte
		out     []Event
	}
	testData := []test{
		{
			"a",
			[]byte{'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
			},
		},
		{
			"space",
			[]byte{' '},
			[]Event{
				KeyDownEvent{Sym: KeySpace, Rune: ' '},
			},
		},
		{
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				KeyDownEvent{Sym: KeyNone, Rune: 'a', Mod: Alt},
			},
		},
		{
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				KeyDownEvent{Sym: KeyNone, Rune: 'a', Mod: Alt},
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
			},
		},
		{
			"ctrl+a",
			[]byte{byte(ansi.SOH)},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl},
			},
		},
		{
			"ctrl+a ctrl+b",
			[]byte{byte(ansi.SOH), byte(ansi.STX)},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl},
				KeyDownEvent{Rune: 'b', Mod: Ctrl},
			},
		},
		{
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Mod: Alt, Rune: 'a'},
			},
		},
		{
			"a b c d",
			[]byte{'a', 'b', 'c', 'd'},
			[]Event{
				KeyDownEvent{Rune: 'a'},
				KeyDownEvent{Rune: 'b'},
				KeyDownEvent{Rune: 'c'},
				KeyDownEvent{Rune: 'd'},
			},
		},
		{
			"up",
			[]byte("\x1b[A"),
			[]Event{
				KeyDownEvent{Sym: KeyUp},
			},
		},
		{
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Event{
				MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		{
			"left motion release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Event{
				MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
				MouseUpEvent{X: 64, Y: 32, Button: MouseNone},
			},
		},
		{
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Event{
				KeyDownEvent{Sym: KeyTab, Mod: Shift},
			},
		},
		{
			"enter",
			[]byte{'\r'},
			[]Event{KeyDownEvent{Sym: KeyEnter}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Event{
				KeyDownEvent{Sym: KeyEnter, Mod: Alt},
			},
		},
		{
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Event{
				KeyDownEvent{Sym: KeyInsert},
			},
		},
		{
			"ctrl+alt+a",
			[]byte{'\x1b', byte(ansi.SOH)},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl | Alt},
			},
		},
		{
			"CSI?----X?",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Event{UnknownCsiEvent([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'})},
		},
		// Powershell sequences.
		{
			"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Event{KeyDownEvent{Sym: KeyUp}},
		},
		{
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Event{KeyDownEvent{Sym: KeyDown}},
		},
		{
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Event{KeyDownEvent{Sym: KeyRight}},
		},
		{
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Event{KeyDownEvent{Sym: KeyLeft}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Event{KeyDownEvent{Sym: KeyEnter, Mod: Alt}},
		},
		{
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Event{KeyDownEvent{Sym: KeyBackspace, Mod: Alt}},
		},
		{
			"ctrl+space",
			[]byte{'\x00'},
			[]Event{KeyDownEvent{Sym: KeySpace, Rune: ' ', Mod: Ctrl}},
		},
		{
			"ctrl+alt+space",
			[]byte{'\x1b', '\x00'},
			[]Event{KeyDownEvent{Sym: KeySpace, Rune: ' ', Mod: Ctrl | Alt}},
		},
		{
			"esc",
			[]byte{'\x1b'},
			[]Event{KeyDownEvent{Sym: KeyEscape}},
		},
		{
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Event{KeyDownEvent{Sym: KeyEscape, Mod: Alt}},
		},
		{
			"a b o",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', ' ', 'b',
				'\x1b', '[', '2', '0', '1', '~',
				'o',
			},
			[]Event{
				PasteStartEvent{},
				PasteEvent("a b"),
				PasteEndEvent{},
				KeyDownEvent{Sym: KeyNone, Rune: 'o'},
			},
		},
		{
			"a\x03\nb",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', '\x03', '\n', 'b',
				'\x1b', '[', '2', '0', '1', '~',
			},
			[]Event{
				PasteStartEvent{},
				PasteEvent("a\x03\nb"),
				PasteEndEvent{},
			},
		},
		{
			"?0xfe?",
			[]byte{'\xfe'},
			nil,
		},
		{
			"a ?0xfe?   b",
			[]byte{'a', '\xfe', ' ', 'b'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				KeyDownEvent{Sym: KeySpace, Rune: ' '},
				KeyDownEvent{Sym: KeyNone, Rune: 'b'},
			},
		},
	}

	for i, td := range testData {
		t.Run(fmt.Sprintf("%d: %s", i, td.keyname), func(t *testing.T) {
			msgs := testReadInputs(t, bytes.NewReader(td.in))
			var buf strings.Builder
			for i, msg := range msgs {
				if i > 0 {
					buf.WriteByte(' ')
				}
				if s, ok := msg.(fmt.Stringer); ok {
					buf.WriteString(s.String())
				} else {
					fmt.Fprintf(&buf, "%#v:%T", msg, msg)
				}
			}

			if len(msgs) != len(td.out) {
				t.Fatalf("unexpected message list length: got %d, expected %d\n%#v", len(msgs), len(td.out), msgs)
			}

			if !reflect.DeepEqual(td.out, msgs) {
				t.Fatalf("expected:\n%#v\ngot:\n%#v", td.out, msgs)
			}
		})
	}
}

const maxBufferSize = 256

func testReadInputs(t *testing.T, input io.Reader) []Event {
	// We'll check that the input reader finishes at the end
	// without error.
	var wg sync.WaitGroup
	var inputErr error
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
		if inputErr != nil && !errors.Is(inputErr, io.EOF) {
			t.Fatalf("unexpected input error: %v", inputErr)
		}
	}()

	dr, err := NewDriver(input, "", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	// The messages we're consuming.
	msgsC := make(chan Event)

	// Start the reader in the background.
	wg.Add(1)
	go func() {
		defer wg.Done()
		var n int
		events := make([]Event, maxBufferSize)
		n, inputErr = dr.ReadInput(events)
	out:
		for _, ev := range events[:n] {
			select {
			case msgsC <- ev:
			case <-ctx.Done():
				break out
			}
		}
		msgsC <- nil
	}()

	var msgs []Event
loop:
	for {
		select {
		case msg := <-msgsC:
			if msg == nil {
				// end of input marker for the test.
				break loop
			}
			msgs = append(msgs, msg)
		case <-time.After(2 * time.Second):
			t.Errorf("timeout waiting for input event")
			break loop
		}
	}
	return msgs
}

// randTest defines the test input and expected output for a sequence
// of interleaved control sequences and control characters.
type randTest struct {
	data    []byte
	lengths []int
	names   []string
}

// seed is the random seed to randomize the input. This helps check
// that all the sequences get ultimately exercised.
var seed = flag.Int64("seed", 0, "random seed (0 to autoselect)")

// genRandomData generates a randomized test, with a random seed unless
// the seed flag was set.
func genRandomData(logfn func(int64), length int) randTest {
	// We'll use a random source. However, we give the user the option
	// to override it to a specific value for reproduceability.
	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	// Inform the user so they know what to reuse to get the same data.
	logfn(s)
	return genRandomDataWithSeed(s, length)
}

// genRandomDataWithSeed generates a randomized test with a fixed seed.
func genRandomDataWithSeed(s int64, length int) randTest {
	src := rand.NewSource(s)
	r := rand.New(src)

	// allseqs contains all the sequences, in sorted order. We sort
	// to make the test deterministic (when the seed is also fixed).
	type seqpair struct {
		seq  string
		name string
	}
	var allseqs []seqpair
	for seq, key := range sequences {
		allseqs = append(allseqs, seqpair{seq, key.String()})
	}
	sort.Slice(allseqs, func(i, j int) bool { return allseqs[i].seq < allseqs[j].seq })

	// res contains the computed test.
	var res randTest

	for len(res.data) < length {
		alt := r.Intn(2)
		prefix := ""
		esclen := 0
		if alt == 1 {
			prefix = "alt+"
			esclen = 1
		}
		kind := r.Intn(3)
		switch kind {
		case 0:
			// A control character.
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res.data = append(res.data, 1)
			res.names = append(res.names, "ctrl+"+prefix+"a")
			res.lengths = append(res.lengths, 1+esclen)

		case 1, 2:
			// A sequence.
			seqi := r.Intn(len(allseqs))
			s := allseqs[seqi]
			if strings.Contains(s.name, "alt+") || strings.Contains(s.name, "meta+") {
				esclen = 0
				prefix = ""
				alt = 0
			}
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res.data = append(res.data, s.seq...)
			if strings.HasPrefix(s.name, "ctrl+") {
				prefix = "ctrl+" + prefix
			}
			name := prefix + strings.TrimPrefix(s.name, "ctrl+")
			res.names = append(res.names, name)
			res.lengths = append(res.lengths, len(s.seq)+esclen)
		}
	}
	return res
}

// TestDetectRandomSequencesLex checks that the lex-generated sequence
// detector works over concatenations of random sequences.
func TestDetectRandomSequencesLex(t *testing.T) {
	t.Skip("WIP")
	runTestDetectSequence(t, NewEventParser("dumb", 0).ParseSequence)
}

func runTestDetectSequence(
	t *testing.T, detectSequence func(input []byte) (width int, msg Event),
) {
	for i := 0; i < 10; i++ {
		t.Run("", func(t *testing.T) {
			td := genRandomData(func(s int64) { t.Logf("using random seed: %d", s) }, 1000)

			t.Logf("%#v", td)

			// tn is the event number in td.
			// i is the cursor in the input data.
			// w is the length of the last sequence detected.
			for tn, i, w := 0, 0, 0; i < len(td.data); tn, i = tn+1, i+w {
				width, msg := detectSequence(td.data[i:])
				if width != td.lengths[tn] {
					t.Errorf("at %d (ev %d): expected width %d, got %d", i, tn, td.lengths[tn], width)
				}
				w = width

				s, ok := msg.(fmt.Stringer)
				if !ok {
					t.Errorf("at %d (ev %d): expected stringer event, got %T", i, tn, msg)
				} else {
					if td.names[tn] != s.String() {
						t.Errorf("at %d (ev %d): expected event %q, %q, got %q", i, tn, td.names[tn], td.data[i:i+w], s.String())
					}
				}
			}
		})
	}
}

// TestDetectRandomSequences checks that the map-based sequence
// detector works over concatenations of random sequences.
func TestDetectRandomSequences(t *testing.T) {
	t.Skip("WIP")
	parser := NewEventParser("dumb", FlagNoTerminfo)
	runTestDetectSequence(t, parser.ParseSequence)
}

func FuzzParseSequence(f *testing.F) {
	parser := NewEventParser("dumb", FlagNoTerminfo)
	for seq := range sequences {
		f.Add(seq)
	}
	f.Fuzz(func(t *testing.T, a string) {
		_, k := parser.ParseSequence([]byte(a))
		v, ok := sequences[a]
		if !ok {
			t.Fatalf("unknown sequence: %q", a)
		}
		if reflect.DeepEqual(k, v) {
			t.Errorf("expected %v, got %v", v, k)
		}
	})
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	td := genRandomDataWithSeed(123, 10000)
	parser := NewEventParser("dumb", FlagNoTerminfo)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			w, _ = parser.ParseSequence(td.data[j:])
		}
	}
}
