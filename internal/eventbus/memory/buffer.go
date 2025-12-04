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
}

func NewBuffer(capacity int) (*Buffer, error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be greater than 0")
	}
	return &Buffer{
		capacity: capacity,
		buffer:   make([][]byte, capacity),
	}, nil
}

func (b *Buffer) Push(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer[b.head] = data
	b.head = (b.head + 1) % b.capacity

	if b.size < b.capacity {
		b.size++
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
