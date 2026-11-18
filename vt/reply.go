package vt

import (
	"io"
	"sync"
)

// maxReplyBytes bounds how much unread *reply* traffic the emulator keeps. A
// consumer that never reads is the normal case for a passive screen model, and
// the traffic that provokes replies — status, device attribute and colour
// queries — is unbounded, so something has to give. Input the consumer sends
// itself is never dropped; see [replyBuffer.Write].
const maxReplyBytes = 64 << 10

// outgoing is one queued write: a complete escape sequence or a run of text.
type outgoing struct {
	data []byte
	// reply marks traffic the emulator generated in answer to a query, which
	// may be dropped under pressure. Everything the consumer sent — keys,
	// mouse reports, pastes — is kept whatever happens.
	reply bool
	// read marks a chunk the consumer has already started reading, which must
	// not be dropped: doing so would splice the middle out of a sequence the
	// consumer is part-way through parsing.
	read bool
}

// replyBuffer carries what the emulator sends to whatever drives it: answers to
// queries such as DSR and DA, and the input a consumer injects through
// [Emulator.InputPipe], [Emulator.Paste] and the key and mouse helpers.
//
// It holds that traffic rather than handing it to an [io.Pipe], because a pipe
// write parks until somebody reads. Ordinary terminal startup traffic asks the
// terminal about itself, so [Emulator.Write] would block forever unless the
// consumer happened to be draining replies concurrently — a requirement that
// was easy to miss and turned normal output into a deadlock.
//
// Writes therefore never block. When the buffer is full, queued replies are
// dropped oldest-first: a stale answer is worth less than a fresh one, and a
// consumer that starts reading late wants the terminal's current state. Queued
// writes are whole, so dropping loses complete sequences rather than slicing
// one in half.
type replyBuffer struct {
	more  *sync.Cond
	queue []outgoing
	mu    sync.Mutex
	// replyBytes counts only the droppable chunks, since those are what the
	// bound governs. Counting consumer input here too would mean a paste
	// larger than the bound left the buffer permanently over it, and every
	// reply after that would be dropped on arrival.
	replyBytes int
	closed     bool
}

func newReplyBuffer() *replyBuffer {
	b := &replyBuffer{}
	b.more = sync.NewCond(&b.mu)
	return b
}

// Write queues input from the consumer. It never blocks and is never dropped,
// since the consumer chose to send it and knows how much it sent.
func (b *replyBuffer) Write(p []byte) (int, error) {
	return b.write(p, false)
}

// writeReply queues traffic the emulator generated in answer to a query. It
// never blocks, and may be dropped if the consumer is not reading.
func (b *replyBuffer) writeReply(p []byte) (int, error) {
	return b.write(p, true)
}

func (b *replyBuffer) write(p []byte, reply bool) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, io.ErrClosedPipe
	}

	// The caller may reuse p.
	chunk := make([]byte, len(p))
	copy(chunk, p)

	b.queue = append(b.queue, outgoing{data: chunk, reply: reply})
	if reply {
		b.replyBytes += len(chunk)
		b.evict()
	}
	b.more.Signal()

	return len(p), nil
}

// evict drops queued replies, oldest first, until the buffer is back within its
// bound. It skips consumer input and any chunk already part-read, so what
// survives is always a sequence of whole, in-order writes. The newest chunk is
// kept even when it alone exceeds the bound, since dropping it would mean a
// large but perfectly good answer never arrives.
func (b *replyBuffer) evict() {
	for i := 0; b.replyBytes > maxReplyBytes && i < len(b.queue)-1; {
		if !b.queue[i].reply || b.queue[i].read {
			i++
			continue
		}

		b.replyBytes -= len(b.queue[i].data)
		if i == 0 {
			// The common case: reslicing beats memmoving the whole queue on
			// every write once the buffer is full.
			b.queue = b.queue[1:]
			continue
		}
		b.queue = append(b.queue[:i], b.queue[i+1:]...)
	}
}

// Read returns queued traffic, blocking until some arrives or the buffer is
// closed. What is already queued is still delivered after a close; [io.EOF]
// follows once it runs out.
func (b *replyBuffer) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for len(b.queue) == 0 && !b.closed {
		b.more.Wait()
	}

	if len(b.queue) == 0 {
		return 0, io.EOF
	}

	head := &b.queue[0]
	n := copy(p, head.data)
	if head.reply {
		b.replyBytes -= n
	}

	if n == len(head.data) {
		b.queue = b.queue[1:]
	} else {
		head.data = head.data[n:]
		head.read = true
	}

	return n, nil
}

// Close stops further writes and wakes every waiting reader.
func (b *replyBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.more.Broadcast()

	return nil
}

// writeReplyString is [io.WriteString] for the droppable reply path.
func writeReplyString(b *replyBuffer, s string) (int, error) {
	return b.writeReply([]byte(s))
}
