package metastore

import (
	"context"
	"encoding/json"
	"time"
)

type Metastore interface {
	Get(ctx context.Context, uniqueID string, key string) ([]byte, error)
	Set(ctx context.Context, uniqueID string, key string, value any) error
	SetWithTTL(ctx context.Context, uniqueID string, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, uniqueID string, key string) error
	Exists(ctx context.Context, uniqueID string, key string) (bool, error)
}

func TypedJSONMeta[T any](ctx context.Context, metastore Metastore, uniqueID string, key string) (T, error) {
	var zero T
	data, err := metastore.Get(ctx, uniqueID, key)
	if err != nil {
		return zero, err
	}

	var typedData T
	err = json.Unmarshal(data, &typedData)
	if err != nil {
		return zero, err
	}

	return typedData, nil
}
