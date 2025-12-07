package eventbus

import (
	"context"
	"errors"
)

var (
	ErrBusClosed     = errors.New("event bus is closed")
	ErrTopicNotFound = errors.New("topic not found")
)

type EventBus interface {
	IEventPublisher
	NewConsumer(topic string) (IEventConsumer, error)

	Close() error
}

type IEventPublisher interface {
	Publish(ctx context.Context, topic string, data any) error
}

type IEventConsumer interface {
	Consume(ctx context.Context) ([]byte, bool)
}
