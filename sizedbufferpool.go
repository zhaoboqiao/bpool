package bpool

import (
	"bytes"
)

// SizedBufferPool implements a pool of bytes.Buffers in the form of a bounded
// channel. Buffers are pre-allocated to the requested size.
type SizedBufferPool struct {
	c chan *bytes.Buffer
	a int
}

// SizedBufferPool creates a new BufferPool bounded to the given size.
// size defines the number of buffers to be retained in the pool and alloc sets
// the initial capacity of new buffers to minimize calls to make().
//
// The value of alloc should seek to provide a buffer that is representative of
// most data written to the the buffer (i.e. 95th percentile) without being
// overly large (which will increase static memory consumption). You may wish to
// track the capacity of your last N buffers (i.e. using an []int) prior to
// returning them to the pool as input into calculating a suitable alloc value.
func NewSizedBufferPool(size int, alloc int) (bp *SizedBufferPool) {
	return &SizedBufferPool{
		c: make(chan *bytes.Buffer, size),
		a: alloc,
	}
}

// Get gets a Buffer from the SizedBufferPool, or creates a new one if none are
// available in the pool. Buffers have a pre-allocated capacity.
func (bp *SizedBufferPool) Get() *bytes.Buffer {
	select {
	case b := <-bp.c:
		// reuse existing buffer
		return b
	default:
		// create new buffer
		return bytes.NewBuffer(make([]byte, 0, bp.a))
	}

}

// Put returns the given Buffer to the SizedBufferPool.
func (bp *SizedBufferPool) Put(b *bytes.Buffer) {
	// If the pool is full opportunistically throw the buffer away
	if len(bp.c) == cap(bp.c) {
		return
	}
	// Release buffers over our maximum capacity and re-create a pre-sized
	// buffer to replace it.
	// if b.Cap() > bp.a {
	// 	b = bytes.NewBuffer(make([]byte, 0, bp.a))
	// } else {
	// 	b.Reset()
	// }
	if b != nil {
		b.Reset()
	}

	select {
	case bp.c <- b:
		return
	default: // Discard the buffer if the pool is full.
	}
}
