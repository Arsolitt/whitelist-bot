package memory

import (
	"errors"
	"sync"
)

// TODO: add generics for data type
// TODO: add sync.Cond for waiting for data
type Buffer struct {
	mu       sync.RWMutex
	buffer   [][]byte
	capacity int
	head     int
	size     int
	dropped  int
	notifier chan struct{}
	closed   bool
}

func NewBuffer(capacity int) (*Buffer, error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be greater than 0")
	}
	return &Buffer{
		capacity: capacity,
		buffer:   make([][]byte, capacity),
		notifier: make(chan struct{}, 1),
	}, nil
}

func (b *Buffer) Push(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.buffer[b.head] = data
	b.head = (b.head + 1) % b.capacity

	if b.size < b.capacity {
		b.size++
	} else {
		b.dropped++
	}

	select {
	case b.notifier <- struct{}{}:
	default:
	}
}

func (b *Buffer) Pop() ([]byte, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.size == 0 {
		return nil, false
	}

	tailIndex := (b.head - b.size + b.capacity) % b.capacity
	data := b.buffer[tailIndex]
	b.buffer[tailIndex] = nil
	b.size--

	return data, true
}

func (b *Buffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		b.closed = true
		close(b.notifier)
	}

}

func (b *Buffer) IsClosed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.closed
}

func (b *Buffer) Notifier() <-chan struct{} {
	return b.notifier
}

func (b *Buffer) Dropped() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.dropped
}
