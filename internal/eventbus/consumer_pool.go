package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"whitelist-bot/internal/wp"
)

type ConsumerUnitHandler func(ctx context.Context, data []byte) error

type ConsumerUnit struct {
	Topic   string
	Handler ConsumerUnitHandler
}

type ConsumerPool struct {
	eBus          EventBus
	units         []ConsumerUnit
	sem           wp.ISemaphore
	wgConsumers   sync.WaitGroup
	wgHandlers    sync.WaitGroup
	handlerCtx    context.Context
	handlerCancel context.CancelFunc
}

func NewConsumerPool(eBus EventBus, units []ConsumerUnit, sem wp.ISemaphore) *ConsumerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConsumerPool{
		eBus:          eBus,
		units:         units,
		sem:           sem,
		handlerCtx:    ctx,
		handlerCancel: cancel,
	}
}

func (p *ConsumerPool) Start(ctx context.Context) error {
	for _, unit := range p.units {
		consumer, err := p.eBus.NewConsumer(unit.Topic)
		if err != nil {
			slog.Error("Failed to get consumer", "error", err.Error())
			return err
		}

		p.wgConsumers.Add(1)
		go func(u ConsumerUnit) {
			defer p.wgConsumers.Done()
			for {
				data, closed := consumer.Consume(ctx)
				if !closed {
					slog.DebugContext(ctx, "Event bus consumer closed")
					return
				}
				if err := p.sem.Acquire(ctx); err != nil {
					slog.DebugContext(ctx, "Semaphore closed", "error", err.Error())
					return
				}

				p.wgHandlers.Add(1)
				go func(d []byte) {
					defer p.wgHandlers.Done()
					defer p.sem.Release()
					if err := u.Handler(p.handlerCtx, d); err != nil {
						slog.ErrorContext(p.handlerCtx, "Failed to handle event", "error", err.Error())
						return
					}
				}(data)
			}
		}(unit)
	}
	return nil
}

func (p *ConsumerPool) Wait() {
	p.wgConsumers.Wait()
	p.wgHandlers.Wait()
	p.handlerCancel()
}
