package vt

import (
	"io"

	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"time"
)

// TestWriteDoesNotBlockOnReplies is the case from the issue: feeding the
// emulator a query with nobody draining its replies must not wedge the writer.
func TestWriteDoesNotBlockOnReplies(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)

	done := make(chan struct{})
	go func() {
		defer close(done)
		term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Write blocked waiting for its reply to be read")
	}
}

// TestRepliesSurviveUntilRead checks the replies are still there afterwards:
// not blocking must not mean discarding.
func TestRepliesSurviveUntilRead(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)
	term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail

	buf := make([]byte, 64)
	n, err := term.Read(buf)
	if err != nil {
		t.Fatalf("reading the reply: %v", err)
	}
	if got, want := string(buf[:n]), "\x1b[?0n"; got != want {
		t.Errorf("reply: got %q, want %q", got, want)
	}
}

// TestRepliesAreBounded checks an unread emulator cannot grow without limit. A
// screen model that never reads is the normal case, and traffic asking the
// terminal about itself is unbounded.
func TestRepliesAreBounded(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)
	for range 20000 {
		term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail
	}

	// Close the buffer so draining terminates, then count what survived.
	if err := term.replies.Close(); err != nil {
		t.Fatalf("closing the reply buffer: %v", err)
	}

	var drained []byte
	buf := make([]byte, 4096)
	for {
		n, err := term.replies.Read(buf)
		drained = append(drained, buf[:n]...)
		if err != nil {
			break
		}
	}

	if len(drained) > maxReplyBytes {
		t.Errorf("kept %d bytes of replies, want at most %d", len(drained), maxReplyBytes)
	}
	if len(drained) == 0 {
		t.Error("dropped every reply, want the most recent ones kept")
	}

	// What survives has to be whole replies in order, not a sequence sliced in
	// half. Compared against an exact repetition, since TrimLeft takes a cutset
	// and would accept any scramble of those bytes.
	const reply = "\x1b[?0n"
	if want := strings.Repeat(reply, len(drained)/len(reply)); string(drained) != want {
		t.Errorf("drained replies are not whole: got %q", drained)
	}
	if len(drained)%len(reply) != 0 {
		t.Errorf("drained %d bytes, not a whole number of %d-byte replies", len(drained), len(reply))
	}
}

// TestReadBlocksUntilAReplyArrives keeps the reader side of the contract: a
// consumer draining the emulator waits rather than spinning on empty reads.
func TestReadBlocksUntilAReplyArrives(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)

	got := make(chan string, 1)
	go func() {
		buf := make([]byte, 64)
		n, err := term.Read(buf)
		if err != nil {
			got <- "error: " + err.Error()
			return
		}
		got <- string(buf[:n])
	}()

	select {
	case reply := <-got:
		t.Fatalf("Read returned %q before anything was written", reply)
	case <-time.After(50 * time.Millisecond):
	}

	term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail

	select {
	case reply := <-got:
		if want := "\x1b[?0n"; reply != want {
			t.Errorf("reply: got %q, want %q", reply, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Read did not wake up when a reply arrived")
	}
}

// TestReadAfterCloseReturnsEOF preserves what closing meant with the pipe: a
// closed emulator reads as EOF. The buffer itself will drain what it holds, but
// [Emulator.Read] short-circuits on close, and that is existing behaviour this
// change deliberately leaves alone.
func TestReadAfterCloseReturnsEOF(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)
	term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail
	if err := term.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	buf := make([]byte, 64)
	if _, err := term.Read(buf); err != io.EOF { //nolint:errorlint // io.EOF is compared by identity
		t.Errorf("reading a closed emulator: got %v, want io.EOF", err)
	}
}

// TestReplyBufferDrainsAfterClose covers the buffer's own contract: a close
// wakes readers and hands over what is already queued before reporting EOF, so
// a consumer draining concurrently does not lose the last reply.
func TestReplyBufferDrainsAfterClose(t *testing.T) {
	t.Parallel()

	b := newReplyBuffer()
	if _, err := b.Write([]byte("\x1b[?0n")); err != nil {
		t.Fatalf("queueing a reply: %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	buf := make([]byte, 64)
	n, err := b.Read(buf)
	if err != nil {
		t.Fatalf("draining after close: %v", err)
	}
	if got, want := string(buf[:n]), "\x1b[?0n"; got != want {
		t.Errorf("drained reply: got %q, want %q", got, want)
	}

	if _, err := b.Read(buf); err != io.EOF { //nolint:errorlint // io.EOF is compared by identity
		t.Errorf("reading an empty closed buffer: got %v, want io.EOF", err)
	}

	if _, err := b.Write([]byte("late")); err != io.ErrClosedPipe { //nolint:errorlint // sentinel compared by identity
		t.Errorf("writing to a closed buffer: got %v, want io.ErrClosedPipe", err)
	}
}

// TestRepliesUnderConcurrentReaderAndWriter exercises the lock: a consumer
// draining while queries keep arriving is the shape this buffer exists for, and
// it is the shape a data race would show up in.
func TestRepliesUnderConcurrentReaderAndWriter(t *testing.T) {
	t.Parallel()

	b := newReplyBuffer()

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 16)
		for {
			if _, err := b.Read(buf); err != nil {
				return
			}
		}
	}()

	for range 5000 {
		if _, err := b.Write([]byte("\x1b[?0n")); err != nil {
			t.Errorf("queueing a reply: %v", err)
			break
		}
	}

	if err := b.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not wake on close")
	}
}

// TestPasteSurvivesReplyPressure checks that what the consumer sends is never
// what gets dropped. A paste larger than the reply bound is still consumer
// input: discarding it would lose the pasted text and leave the application
// with a bracketed-paste end and no start.
func TestPasteSurvivesReplyPressure(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)
	term.Write([]byte("\x1b[?2004h")) //nolint:errcheck // writing to an emulator cannot fail

	text := strings.Repeat("x", maxReplyBytes+4096)
	term.Paste(text)

	if err := term.replies.Close(); err != nil {
		t.Fatalf("closing the reply buffer: %v", err)
	}

	var drained []byte
	buf := make([]byte, 8192)
	for {
		n, err := term.replies.Read(buf)
		drained = append(drained, buf[:n]...)
		if err != nil {
			break
		}
	}

	want := ansi.BracketedPasteStart + text + ansi.BracketedPasteEnd
	if string(drained) != want {
		t.Errorf("paste of %d bytes drained as %d bytes; consumer input must not be dropped",
			len(text), len(drained))
	}
}

// TestPartiallyReadReplyIsNotEvicted covers the other way a stream can be
// spliced: dropping the remainder of a chunk the consumer is part-way through
// reading would hand it the front of one sequence and the back of another.
func TestPartiallyReadReplyIsNotEvicted(t *testing.T) {
	t.Parallel()

	b := newReplyBuffer()

	const reply = "\x1b[?0n"
	if _, err := b.writeReply([]byte(reply)); err != nil {
		t.Fatalf("queueing a reply: %v", err)
	}

	// Read two bytes, leaving the rest of that reply queued.
	head := make([]byte, 2)
	if _, err := b.Read(head); err != nil {
		t.Fatalf("reading the head of a reply: %v", err)
	}

	// Now push the buffer well past its bound.
	for range 20000 {
		if _, err := b.writeReply([]byte(reply)); err != nil {
			t.Fatalf("queueing a reply: %v", err)
		}
	}

	if err := b.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	rest := make([]byte, len(reply))
	n, err := b.Read(rest)
	if err != nil {
		t.Fatalf("reading the rest of the part-read reply: %v", err)
	}
	if got, want := string(rest[:n]), reply[2:]; got != want {
		t.Errorf("part-read reply resumed as %q, want %q", got, want)
	}
}

// TestCloseUnblocksAConcurrentReader is the shape that makes the closed flag
// worth synchronising: one goroutine parked in Read, another closing. Under
// -race this is where an unsynchronised flag shows up.
func TestCloseUnblocksAConcurrentReader(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)

	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		_, err := term.Read(buf)
		done <- err
	}()

	// Give the reader time to park.
	time.Sleep(50 * time.Millisecond)

	if err := term.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	select {
	case err := <-done:
		if err != io.EOF { //nolint:errorlint // io.EOF is compared by identity
			t.Errorf("parked reader woke with %v, want io.EOF", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not wake the parked reader")
	}
}

// drainReplies closes the emulator's reply buffer and returns everything it
// held, so a test can assert on replies without a concurrent reader.
func drainReplies(t *testing.T, term *Emulator) string {
	t.Helper()

	if err := term.replies.Close(); err != nil {
		t.Fatalf("closing the reply buffer: %v", err)
	}

	var drained []byte
	buf := make([]byte, 8192)
	for {
		n, err := term.replies.Read(buf)
		drained = append(drained, buf[:n]...)
		if err != nil {
			break
		}
	}

	return string(drained)
}

// TestQueriesAnsweredAfterABigPaste checks that consumer input, which is never
// dropped, cannot starve the replies. If the bound counted that input, a paste
// larger than the bound would leave the buffer permanently over it and answers
// arriving afterwards would be dropped — an application waiting on its
// device-attributes answer would hang.
func TestQueriesAnsweredAfterABigPaste(t *testing.T) {
	t.Parallel()

	// What the answers look like with nothing else going on.
	quiet := newTestTerminal(t, 80, 24)
	quiet.Write([]byte("\x1b[c"))  //nolint:errcheck // primary device attributes
	quiet.Write([]byte("\x1b[5n")) //nolint:errcheck // device status report
	answers := drainReplies(t, quiet)
	da, dsr, ok := strings.Cut(answers, "\x1b[?0n")
	if !ok || da == "" {
		t.Fatalf("could not tell the two answers apart in %q", answers)
	}
	dsr = "\x1b[?0n"

	term := newTestTerminal(t, 80, 24)
	term.Paste(strings.Repeat("x", maxReplyBytes+1))
	term.Write([]byte("\x1b[c"))  //nolint:errcheck // writing to an emulator cannot fail
	term.Write([]byte("\x1b[5n")) //nolint:errcheck // writing to an emulator cannot fail

	drained := drainReplies(t, term)
	if !strings.Contains(drained, da) {
		t.Errorf("device attributes answer %q was dropped after a large paste", da)
	}
	if !strings.Contains(drained, dsr) {
		t.Errorf("status answer %q was dropped after a large paste", dsr)
	}
}

// TestResizeNotificationSurvivesReplyPressure keeps the one reply that is not
// re-derivable. A stale status answer is worth dropping; a size the child never
// learns leaves it rendering to the wrong dimensions forever.
func TestResizeNotificationSurvivesReplyPressure(t *testing.T) {
	t.Parallel()

	term := newTestTerminal(t, 80, 24)
	term.Write([]byte("\x1b[?2048h")) //nolint:errcheck // ask for in-band resize notifications
	term.Resize(100, 30)

	for range 20000 {
		term.Write([]byte("\x1b[5n")) //nolint:errcheck // pile on droppable replies
	}

	drained := drainReplies(t, term)
	if want := ansi.InBandResize(30, 100, 0, 0); !strings.Contains(drained, want) {
		t.Errorf("resize notification %q was dropped under reply pressure", want)
	}
}
