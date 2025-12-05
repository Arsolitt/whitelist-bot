package memory

import (
	"context"
	"log/slog"
	"whitelist-bot/internal/eventbus"
	"whitelist-bot/internal/wp"
)

type ConsumerUnitHandler func(ctx context.Context, data []byte) error

type ConsumerUnit struct {
	Topic   string
	Handler ConsumerUnitHandler
}

type ConsumerPool struct {
	eventBus eventbus.EventBus
	units    []ConsumerUnit
	sem      *wp.Semaphore
}

func NewConsumerPool(eventBus eventbus.EventBus, units []ConsumerUnit, sem *wp.Semaphore) *ConsumerPool {
	return &ConsumerPool{
		eventBus: eventBus,
		units:    units,
		sem:      sem,
	}
}

func (p *ConsumerPool) Start(ctx context.Context) error {
	for _, unit := range p.units {
		consumer, err := p.eventBus.NewConsumer(unit.Topic)
		if err != nil {
			slog.Error("Failed to get consumer", "error", err.Error())
			return err
		}
		go func(unit ConsumerUnit) {
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
				go func(d []byte) {
					defer p.sem.Release()
					if err := unit.Handler(ctx, d); err != nil {
						slog.ErrorContext(ctx, "Failed to handle event", "error", err.Error())
						return
					}
				}(data)
			}
		}(unit)
	}
	return nil
}
