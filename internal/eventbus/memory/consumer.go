package memory

import (
	"context"
)

type Consumer struct {
	buffer *Buffer
}

func (c *Consumer) Consume(ctx context.Context) ([]byte, bool) {
	for {

		if data, ok := c.buffer.Pop(); ok {
			return data, true
		}

		select {
		case <-ctx.Done():
			if data, ok := c.buffer.Pop(); ok {
				return data, true
			}
			return nil, false
		case _, ok := <-c.buffer.Notifier():
			if !ok {
				if data, ok := c.buffer.Pop(); ok {
					return data, true
				}
				return nil, false
			}
		}
	}
}
