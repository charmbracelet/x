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

	"github.com/charmbracelet/x/ansi"
)

var sequences = buildKeysTable(FlagTerminfo, "dumb")

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

func TestParseSequence(t *testing.T) {
	td := buildBaseSeqTests()
	td = append(td,
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Event{
				MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
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
				KeyDownEvent{Rune: 'a'},
			},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Alt},
			},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			[]Event{
				KeyDownEvent{Rune: 'a'},
				KeyDownEvent{Rune: 'a'},
				KeyDownEvent{Rune: 'a'},
			},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			[]Event{
				KeyDownEvent{Rune: '☃'},
			},
		},
		seqTest{
			[]byte("\x1b☃"),
			[]Event{
				KeyDownEvent{Rune: '☃', Mod: Alt},
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
			[]byte{ansi.SOH},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.SOH},
			[]Event{
				KeyDownEvent{Rune: 'a', Mod: Ctrl | Alt},
			},
		},
		seqTest{
			[]byte{ansi.NUL},
			[]Event{
				KeyDownEvent{Rune: ' ', Sym: KeySpace, Mod: Ctrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.NUL},
			[]Event{
				KeyDownEvent{Rune: ' ', Sym: KeySpace, Mod: Ctrl | Alt},
			},
		},
		// C1 control characters.
		seqTest{
			[]byte{'\x80'},
			[]Event{
				KeyDownEvent{Rune: 0x80 - '@', Mod: Ctrl | Alt},
			},
		},
	)

	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			[]Event{
				UnknownEvent(rune(0xfe)),
			},
		})
	}

	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Event
			buf := tc.seq
			for len(buf) > 0 {
				width, msg := ParseSequence(buf)
				events = append(events, msg)
				buf = buf[width:]
			}
			if !reflect.DeepEqual(tc.msgs, events) {
				t.Errorf("\nexpected event:\n    %#v\ngot:\n    %#v", tc.msgs, events)
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
	drv, err := NewDriver(strings.NewReader(input), "dumb", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	var msgs []Event
	for {
		events, err := drv.ReadEvents()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected input error: %v", err)
		}
		msgs = append(msgs, events...)
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
			[]Event{
				UnknownEvent(rune(0xfe)),
			},
		},
		{
			"a ?0xfe?   b",
			[]byte{'a', '\xfe', ' ', 'b'},
			[]Event{
				KeyDownEvent{Sym: KeyNone, Rune: 'a'},
				UnknownEvent(rune(0xfe)),
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
				t.Fatalf("unexpected message list length: got %d, expected %d\n  got: %#v\n  expected: %#v\n", len(msgs), len(td.out), msgs, td.out)
			}

			if !reflect.DeepEqual(td.out, msgs) {
				t.Fatalf("expected:\n%#v\ngot:\n%#v", td.out, msgs)
			}
		})
	}
}

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

	dr, err := NewDriver(input, "dumb", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	// The messages we're consuming.
	msgsC := make(chan Event)

	// Start the reader in the background.
	wg.Add(1)
	go func() {
		defer wg.Done()
		var events []Event
		events, inputErr = dr.ReadEvents()
	out:
		for _, ev := range events {
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

func FuzzParseSequence(f *testing.F) {
	for seq := range sequences {
		f.Add(seq)
	}
	f.Add("\x1b]52;?\x07")                      // OSC 52
	f.Add("\x1b]11;rgb:0000/0000/0000\x1b\\")   // OSC 11
	f.Add("\x1bP>|charm terminal(0.1.2)\x1b\\") // DCS (XTVERSION)
	f.Add("\x1b_Gi=123\x1b\\")                  // APC
	f.Fuzz(func(t *testing.T, seq string) {
		n, _ := ParseSequence([]byte(seq))
		if n == 0 && seq != "" {
			t.Errorf("expected a non-zero width for %q", seq)
		}
	})
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	td := genRandomDataWithSeed(123, 10000)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			w, _ = ParseSequence(td.data[j:])
		}
	}
}
