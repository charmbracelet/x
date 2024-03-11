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
	marker byte
	params [][]uint
	inter  byte
	ignore bool
	rune   byte
}

func (testCsiSequence) sequence() {}

type testOscSequence struct {
	params [][]byte
	bell   bool
}

func (testOscSequence) sequence() {}

type testSosPmApcSequence struct {
	k    byte
	data []byte
}

func (testSosPmApcSequence) sequence() {}

type testEscSequence struct {
	inter  byte
	ignore bool
	rune   byte
}

func (testEscSequence) sequence() {}

type testDcsSequence struct {
	marker byte
	params [][]uint
	inter  byte
	data   []byte
	rune   byte
	ignore bool
}

func (testDcsSequence) sequence() {}

type testRune rune

func (testRune) sequence() {}

type testDispatcher struct {
	dispatched []testSequence
}

// CsiDispatch implements Performer.
func (d *testDispatcher) CsiDispatch(marker byte, params [][]uint, inter byte, r byte, ignore bool) {
	d.dispatched = append(d.dispatched, testCsiSequence{
		params: params,
		inter:  inter,
		marker: marker,
		ignore: ignore,
		rune:   r,
	})
}

// EscDispatch implements Performer.
func (d *testDispatcher) EscDispatch(inter byte, r byte, ignore bool) {
	d.dispatched = append(d.dispatched, testEscSequence{
		inter:  inter,
		ignore: ignore,
		rune:   r,
	})
}

// Execute implements Performer.
func (*testDispatcher) Execute(b byte) {}

// DcsHook implements Performer.
func (d *testDispatcher) DcsDispatch(marker byte, params [][]uint, inter byte, r byte, data []byte, ignore bool) {
	d.dispatched = append(d.dispatched, testDcsSequence{
		params: params,
		inter:  inter,
		marker: marker,
		ignore: ignore,
		rune:   r,
		data:   data,
	})
}

// OscDispatch implements Performer.
func (d *testDispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	d.dispatched = append(d.dispatched, testOscSequence{
		params: params,
		bell:   bellTerminated,
	})
}

// SosPmApcDispatch implements Performer.
func (d *testDispatcher) SosPmApcDispatch(k byte, data []byte) {
	d.dispatched = append(d.dispatched, testSosPmApcSequence{
		k:    k,
		data: data,
	})
}

// Print implements Performer.
func (d *testDispatcher) Print(r rune) {
	d.dispatched = append(d.dispatched, testRune(r))
}

func testParser(d *testDispatcher) Parser {
	return Parser{
		Print:            d.Print,
		Execute:          d.Execute,
		CsiDispatch:      d.CsiDispatch,
		OscDispatch:      d.OscDispatch,
		EscDispatch:      d.EscDispatch,
		DcsDispatch:      d.DcsDispatch,
		SosPmApcDispatch: d.SosPmApcDispatch,
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

	parser := Parser{
		Print:            func(r rune) {},
		Execute:          func(b byte) {},
		CsiDispatch:      func(marker byte, params [][]uint, inter byte, r byte, ignore bool) {},
		OscDispatch:      func(params [][]byte, bellTerminated bool) {},
		EscDispatch:      func(inter byte, r byte, ignore bool) {},
		DcsDispatch:      func(marker byte, params [][]uint, inter byte, r byte, data []byte, ignore bool) {},
		SosPmApcDispatch: func(k byte, data []byte) {},
	}
	parser.Parse(bts)
}

func BenchmarkStateChanges(bm *testing.B) {
	input := "\x1b]2;X\x1b\\ \x1b[0m \x1bP0@\x1b\\"

	for i := 0; i < bm.N; i++ {
		var parser Parser
		for i := 0; i < 1000; i++ {
			parser.Parse([]byte(input))
		}
	}
}
