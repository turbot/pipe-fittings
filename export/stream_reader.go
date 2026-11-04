package export

import (
	"bytes"
	"io"
)

// StreamReader is an io.Reader whose content is produced by a write function.
//
// io.Copy uses WriteTo, which calls the write function with the destination writer,
// so the content is never held in memory. Read is a fallback for other consumers:
// it writes all the content to a buffer on the first call.
type StreamReader struct {
	write func(io.Writer) error
	buf   *bytes.Buffer
	err   error
}

func NewStreamReader(write func(io.Writer) error) *StreamReader {
	return &StreamReader{write: write}
}

// WriteTo implements io.WriterTo
func (r *StreamReader) WriteTo(w io.Writer) (int64, error) {
	if r.buf != nil || r.err != nil {
		// Read has already been called - copy what is left
		if r.err != nil {
			return 0, r.err
		}
		return r.buf.WriteTo(w)
	}
	cw := &countingWriter{w: w}
	err := r.write(cw)
	return cw.n, err
}

// Read implements io.Reader
func (r *StreamReader) Read(p []byte) (int, error) {
	if r.buf == nil && r.err == nil {
		r.buf = &bytes.Buffer{}
		r.err = r.write(r.buf)
	}
	if r.err != nil {
		return 0, r.err
	}
	return r.buf.Read(p)
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}
