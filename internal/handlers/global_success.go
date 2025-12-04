package handlers

import (
	"context"
	"log/slog"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func GlobalSuccessHandler(
	cfg core.Config,
) func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State, response router.Response) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State, response router.Response) {
		if response == nil {
			return
		}
		err := response.Answer(ctx, b, update, state, cfg)
		slog.DebugContext(ctx, "Success handler called")
		if err != nil {
			slog.ErrorContext(ctx, "Failed to answer response", logger.ErrorField, err.Error())
			return
		}
	}
}
