//revive:disable:add-constant
package ctrlioreader

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"
)

// const BufferCap = bufio.MaxScanTokenSize
const (
	BufferSize    = 1024
	PauseDuration = 100 * time.Millisecond
)

// controlled reader
type CtrlReader interface {
	Read(b []byte) (n int, err error)
	SwitchToAnother(r io.Reader)
	IsEOF() bool
	EOF()
}

type ctrlReader struct {
	buf          []byte
	nativeReader io.Reader
	controlChan  chan struct{}
	allowEOF     bool
	eof          bool
	mu           *sync.Mutex
}

func (in *ctrlReader) Read(p []byte) (n int, err error) {
	for {
		n, err := in.nativeReader.Read(in.buf)
		if n > 0 {
			copy(p, in.buf[:n])
		}
		switch {
		case n > 0 && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF): // No error of has error and data
			in.eof = false
			return n, err //nolint:wrapcheck
		case n > 0 && errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF): // EOF found
			in.eof = true
			return n, nil
		case err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF): // n == 0,
			in.eof = false
			return 0, err //nolint:wrapcheck
		case errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF): // EOF found
			in.eof = true
			if in.allowEOF {
				return 0, io.EOF
			}
		}
		// no data && EOF not allowed
		select { // block and wait next data portion
		case <-in.controlChan:
			return 0, io.EOF
		case <-time.After(PauseDuration):
			continue
		}
	}
}

func (in *ctrlReader) SwitchToAnother(r io.Reader) {
	in.mu.Lock()
	defer in.mu.Unlock()
	in.nativeReader = r
}

func (in *ctrlReader) EOF() {
	in.allowEOF = true
}

func (in *ctrlReader) IsEOF() bool {
	return in.eof
}

func (in *ctrlReader) finish() {
	in.allowEOF = true
	in.controlChan <- struct{}{}
}

func NewCtrlReader(ctx context.Context, r io.Reader, bufSize int) (CtrlReader, func()) {
	if bufSize == 0 {
		bufSize = BufferSize
	}

	newR := &ctrlReader{
		buf:          make([]byte, bufSize),
		nativeReader: r,
		controlChan:  make(chan struct{}, 16),
		mu:           &sync.Mutex{},
	}
	go func() {
		<-ctx.Done()
		newR.finish()
	}()
	return newR, func() {
		newR.finish()
	}
}
