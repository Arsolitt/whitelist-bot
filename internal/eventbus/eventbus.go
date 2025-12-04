package eventbus

import (
	"context"
	"errors"
)

const (
	TopicUserInfo = "user.info"
)

var (
	ErrBusClosed     = errors.New("event bus is closed")
	ErrTopicNotFound = errors.New("topic not found")
)

type EventBus interface {
	Publish(ctx context.Context, topic string, data any) error

	Close() error
}

type IEventPublisher interface {
	Publish(ctx context.Context, topic string, data any) error
}

type IEventConsumer interface {
	Consume(ctx context.Context) ([]byte, bool)
}

// func TypedJSONData[T any](ctx context.Context, metastore Metastore, uniqueID string, key string) (T, error) {
// 	var zero T
// 	data, err := metastore.Get(ctx, uniqueID, key)
// 	if err != nil {
// 		return zero, err
// 	}

// 	var typedData T
// 	err = json.Unmarshal(data, &typedData)
// 	if err != nil {
// 		return zero, err
// 	}

// 	return typedData, nil
// }
