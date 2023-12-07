package threadsaferwc

import (
	"bytes"
	"io"
	"sync"
)

const DefaultBufCapacity = 1024

type ThreadsafeRWC struct {
	bbuf   *bytes.Buffer
	closed bool
	mu     *sync.Mutex
}

func (r *ThreadsafeRWC) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	return r.bbuf.Read(p) //nolint: wrapcheck
}

func (r *ThreadsafeRWC) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return 0, io.ErrClosedPipe
	}
	return r.bbuf.Write(p) //nolint: wrapcheck
}

func (r *ThreadsafeRWC) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	return nil
}

func (r *ThreadsafeRWC) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bbuf.String()
}

func NewThreadsafeRWC(b []byte) *ThreadsafeRWC {
	if cap(b) == 0 {
		b = make([]byte, 0, DefaultBufCapacity)
	}
	return &ThreadsafeRWC{
		bbuf: bytes.NewBuffer(b),
		mu:   &sync.Mutex{},
	}
}
