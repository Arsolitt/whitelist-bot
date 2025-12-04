package memory

import (
	"context"
	"time"
)

type Consumer struct {
	buffer *Buffer
}

func (c *Consumer) Consume(ctx context.Context) ([]byte, bool) {
	for {
		select {
		case <-ctx.Done():
			return nil, false
		default:
			if data, ok := c.buffer.Pop(); ok {
				return data, true
			}
			time.Sleep(time.Second)
		}
	}
}
