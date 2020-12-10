package util

import (
	"errors"
	"io"
	"sync"
)

var (
	CopyInterrupt = errors.New("stop with interrupt")
)

type CounterReader struct {
	r        io.Reader
	StopChan chan struct{}
	lock     sync.RWMutex // protects n and err
	n        int64
	err      error
}

// NewReader makes a new CounterReader that counts the bytes
// read through it.
func NewCounterReader(r io.Reader) *CounterReader {
	return &CounterReader{
		r:        r,
		StopChan: make(chan struct{}, 1),
	}
}

func (r *CounterReader) Read(p []byte) (n int, err error) {
	select {
	case <-r.StopChan:
		return 0, CopyInterrupt
	default:
		n, err = r.r.Read(p)
		r.lock.Lock()
		r.n += int64(n)
		r.err = err
		r.lock.Unlock()
		return
	}
}

// N gets the number of bytes that have been read
// so far.
func (r *CounterReader) N() int64 {
	var n int64
	r.lock.RLock()
	n = r.n
	r.lock.RUnlock()
	return n
}

// Err gets the last error from the CounterReader.
func (r *CounterReader) Err() error {
	var err error
	r.lock.RLock()
	err = r.err
	r.lock.RUnlock()
	return err
}
