package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"whitelist-bot/internal/eventbus"
)

type Bus struct {
	mu       sync.RWMutex
	topics   map[string]*Buffer
	capacity int
	closed   bool
}

// Not support multiple readers from 1 topic.
// Use different implementations if you need multiple readers from 1 topic.
func New(bufferCapacity int) *Bus {
	return &Bus{
		topics:   make(map[string]*Buffer),
		capacity: bufferCapacity,
	}
}

func (b *Bus) Publish(ctx context.Context, topic string, data any) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to json marshal data: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return eventbus.ErrBusClosed
	}

	_, exists := b.topics[topic]
	if !exists {
		buffer, err := NewBuffer(b.capacity)
		if err != nil {
			return fmt.Errorf("failed to create buffer: %w", err)
		}
		b.topics[topic] = buffer
	}
	b.topics[topic].Push(dataBytes)
	return nil
}

func (b *Bus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	for _, buffer := range b.topics {
		buffer.Close()
	}
	b.topics = nil
	return nil
}

func (b *Bus) NewConsumer(topic string) (eventbus.IEventConsumer, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, eventbus.ErrBusClosed
	}

	_, exists := b.topics[topic]
	if !exists {
		buffer, err := NewBuffer(b.capacity)
		if err != nil {
			return nil, fmt.Errorf("failed to create buffer: %w", err)
		}
		b.topics[topic] = buffer
	}
	return &Consumer{
		buffer: b.topics[topic],
	}, nil
}
