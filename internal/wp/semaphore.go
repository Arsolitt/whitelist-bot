package wp

import (
	"context"
)

type ISemaphore interface {
	Acquire(ctx context.Context) error
	Release()
}

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(maxConcurrent int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, maxConcurrent),
	}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.sem <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) Release() {
	<-s.sem
}
