package ansi

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
	params        [][]uint
	intermediates [2]byte
	ignore        bool
	rune          byte
}

func (testCsiSequence) sequence() {}

type testOscSequence struct {
	params [][]byte
	bell   bool
}

func (testOscSequence) sequence() {}

type testEscSequence struct {
	intermediates [2]byte
	ignore        bool
	rune          byte
}

func (testEscSequence) sequence() {}

type testDcsSequence struct {
	params        [][]uint
	intermediates [2]byte
	data          []byte
	rune          byte
	ignore        bool
}

func (testDcsSequence) sequence() {}

type testDcsPutSequence byte

func (testDcsPutSequence) sequence() {}

type testDcsUnhookSequence struct{}

func (testDcsUnhookSequence) sequence() {}

type testRune rune

func (testRune) sequence() {}

type testDispatcher struct {
	dispatched []testSequence
}

// CsiDispatch implements Performer.
func (d *testDispatcher) CsiDispatch(params [][]uint, intermediates [2]byte, r byte, ignore bool) {
	d.dispatched = append(d.dispatched, testCsiSequence{
		params:        params,
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
	})
}

// EscDispatch implements Performer.
func (d *testDispatcher) EscDispatch(intermediates [2]byte, r byte, ignore bool) {
	d.dispatched = append(d.dispatched, testEscSequence{
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
	})
}

// Execute implements Performer.
func (*testDispatcher) Execute(b byte) {}

// DcsHook implements Performer.
func (d *testDispatcher) DcsDispatch(params [][]uint, intermediates [2]byte, r byte, data []byte, ignore bool) {
	d.dispatched = append(d.dispatched, testDcsSequence{
		params:        params,
		intermediates: intermediates,
		ignore:        ignore,
		rune:          r,
		data:          data,
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

func testHandler(d *testDispatcher) *Handler {
	return &Handler{
		Rune:       d.Print,
		Execute:    d.Execute,
		CsiHandler: d.CsiDispatch,
		OscHandler: d.OscDispatch,
		EscHandler: d.EscDispatch,
		DcsHandler: d.DcsDispatch,
	}
}

func TestSaddu16(t *testing.T) {
	assert.Equal(t, uint(math.MaxUint), saddu(math.MaxUint, 1))
	assert.Equal(t, uint(1), saddu(1, 0))
}

func TestSmulu16(t *testing.T) {
	assert.Equal(t, uint(math.MaxUint), smulu(math.MaxUint, 1))
	assert.Equal(t, uint(math.MaxUint), smulu(math.MaxUint, 2))
	assert.Equal(t, uint(0), smulu(math.MaxUint, 0))
	assert.Equal(t, uint(0), smulu(0, 0))
}

func BenchmarkNext(bm *testing.B) {
	bts, err := os.ReadFile("./fixtures/demo.vte")
	if err != nil {
		bm.Fatalf("Error: %v", err)
	}

	bm.ResetTimer()

	parser := New(nil)
	parser.Parse(bts)
}

func BenchmarkStateChanges(bm *testing.B) {
	input := "\x1b]2;X\x1b\\ \x1b[0m \x1bP0@\x1b\\"

	for i := 0; i < bm.N; i++ {
		parser := New(nil)

		for i := 0; i < 1000; i++ {
			parser.Parse([]byte(input))
		}
	}
}
