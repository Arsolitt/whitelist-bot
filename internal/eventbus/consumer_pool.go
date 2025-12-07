package eventbus

import (
	"context"
	"log/slog"
	"sync"
	"whitelist-bot/internal/wp"
)

type ConsumerUnitHandler func(ctx context.Context, data []byte) error

type ConsumerUnit struct {
	Topic   string // TODO: add topic type with validation
	Handler ConsumerUnitHandler
}

type ConsumerPool struct {
	eBus        EventBus
	units       []ConsumerUnit
	sem         wp.ISemaphore
	wgConsumers sync.WaitGroup
	wgHandlers  sync.WaitGroup
}

func NewConsumerPool(eBus EventBus, units []ConsumerUnit, sem wp.ISemaphore) *ConsumerPool {
	return &ConsumerPool{
		eBus:  eBus,
		units: units,
		sem:   sem,
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
					if err := u.Handler(ctx, d); err != nil {
						slog.ErrorContext(ctx, "Failed to handle event", "error", err.Error())
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
}
