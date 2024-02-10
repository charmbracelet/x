package input

import "io"

// Reader is an interface for reading input from a terminal.
// It supports peeking bytes.
type Reader struct {
	r        io.Reader  // underlying reader
	leftover []byte     // leftover bytes from last read
	buf      [1024]byte // buffer for reading
	n        int        // last number of bytes read
}

// NewReader returns a new Reader.
func NewReader(r io.Reader) *Reader {
	rd := new(Reader)
	rd.r = r
	return rd
}

// Read reads input from the terminal.
func (r *Reader) Read(p []byte) (n int, err error) {
	// If there are leftover bytes from the last read, use them first.
	if len(r.leftover) > 0 {
		n = copy(p, r.leftover)
		r.leftover = r.leftover[n:]
		p = p[n:]
	}

	// If there are still bytes to read, read them.
	if len(p) > 0 {
		r.n, err = r.r.Read(r.buf[:])
		if r.n > 0 {
			n += copy(p, r.buf[:r.n])
			r.leftover = r.buf[n:r.n]
		}
	}

	return
}

// Peek returns the next n bytes without advancing the reader.
func (r *Reader) Peek(n int) (p []byte, err error) {
	// If there are leftover bytes from the last read, use them first.
	if len(r.leftover) > 0 {
		p = append(p, r.leftover...)
		n -= len(r.leftover)
	}

	// If there are still bytes to read, read them.
	if n > 0 {
		r.n, err = r.r.Read(r.buf[:])
		if r.n > 0 {
			p = append(p, r.buf[:r.n]...)
		}
	}

	return
}
