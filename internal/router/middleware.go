package router

import (
	"context"
	"log/slog"
	"time"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DurationMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error) {
		start := time.Now()

		slog.InfoContext(ctx, "Duration middleware started")

		nextState, err := next(ctx, b, update, currentState)

		duration := time.Since(start)

		ctx = logger.WithLogValue(ctx, logger.DurationField, duration.String())

		slog.InfoContext(ctx, "Duration middleware completed")

		return nextState, err
	}
}
