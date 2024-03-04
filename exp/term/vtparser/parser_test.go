package parser

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	input    string
	expected []testSequence
}

type testSequence interface {
	sequence()
}

type testCsiSequence struct {
	prefix        string
	params        [][]uint16
	intermediates []byte
	ignore        bool
	rune          rune
}

func (testCsiSequence) sequence() {}

type testOscSequence struct {
	params [][]byte
	bell   bool
}

func (testOscSequence) sequence() {}

type testEscSequence struct {
	intermediates []byte
	ignore        bool
	rune          rune
}

func (testEscSequence) sequence() {}

type testDcsHookSequence struct {
	prefix        string
	params        [][]uint16
	intermediates []byte
	ignore        bool
	rune          rune
}

func (testDcsHookSequence) sequence() {}

type testDcsPutSequence byte

func (testDcsPutSequence) sequence() {}

type testDcsUnhookSequence struct{}

func (testDcsUnhookSequence) sequence() {}

type testRune rune

func (testRune) sequence() {}

type testDispatcher struct {
	dispatched []testSequence
}

var _ Handler = &testDispatcher{}

// CsiDispatch implements Performer.
func (d *testDispatcher) CsiDispatch(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
	d.dispatched = append(d.dispatched, testCsiSequence{
		prefix:        prefix,
		params:        params,
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
	})
}

// EscDispatch implements Performer.
func (d *testDispatcher) EscDispatch(intermediates []byte, r rune, ignore bool) {
	d.dispatched = append(d.dispatched, testEscSequence{
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
	})
}

// Execute implements Performer.
func (*testDispatcher) Execute(b byte) {}

// DcsHook implements Performer.
func (d *testDispatcher) DcsHook(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
	d.dispatched = append(d.dispatched, testDcsHookSequence{
		prefix:        prefix,
		params:        params,
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
	})
}

// OscDispatch implements Performer.
func (d *testDispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	d.dispatched = append(d.dispatched, testOscSequence{
		params: params,
		bell:   bellTerminated,
	})
}

// Print implements Performer.
func (d *testDispatcher) Print(r rune) {
	d.dispatched = append(d.dispatched, testRune(r))
}

// DcsPut implements Performer.
func (d *testDispatcher) DcsPut(b byte) {
	d.dispatched = append(d.dispatched, testDcsPutSequence(b))
}

// DcsUnhook implements Performer.
func (d *testDispatcher) DcsUnhook() {
	d.dispatched = append(d.dispatched, testDcsUnhookSequence{})
}

func TestSaddu16(t *testing.T) {
	assert.Equal(t, uint16(math.MaxUint16), saddu16(math.MaxUint16, 1))
	assert.Equal(t, uint16(1), saddu16(1, 0))
}

func TestSmulu16(t *testing.T) {
	assert.Equal(t, uint16(math.MaxUint16), smulu16(math.MaxUint16, 1))
	assert.Equal(t, uint16(math.MaxUint16), smulu16(math.MaxUint16, 2))
	assert.Equal(t, uint16(0), smulu16(math.MaxUint16, 0))
	assert.Equal(t, uint16(0), smulu16(0, 0))
}

type benchDispatcher struct{}

func (p *benchDispatcher) Print(r rune) {}

func (p *benchDispatcher) Execute(code byte) {}

func (p *benchDispatcher) DcsPut(code byte) {}

func (p *benchDispatcher) DcsUnhook() {}

func (p *benchDispatcher) DcsHook(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
}

func (p *benchDispatcher) OscDispatch(params [][]byte, bellTerminated bool) {}

func (p *benchDispatcher) CsiDispatch(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
}

func (p *benchDispatcher) EscDispatch(intermediates []byte, final rune, ignore bool) {}

func BenchmarkNext(bm *testing.B) {
	f, err := os.Open("./fixtures/demo.vte")
	if err != nil {
		bm.Fatalf("Error: %v", err)
	}

	defer f.Close()
	bm.ResetTimer()
	dispatcher := &benchDispatcher{}
	parser := New(dispatcher)
	if err := parser.Parse(f); err != nil {
		bm.Fatal(err)
	}
}

func BenchmarkStateChanges(bm *testing.B) {
	input := "\x1b]2;X\x1b\\ \x1b[0m \x1bP0@\x1b\\"

	for i := 0; i < bm.N; i++ {
		dispatcher := &benchDispatcher{}
		parser := New(dispatcher)

		for i := 0; i < 1000; i++ {
			for _, b := range []byte(input) {
				parser.advance(b)
			}
		}
	}
}
