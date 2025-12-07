package handlers

import (
	"context"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Cancel() router.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, router.Response, error) {
		response := router.NewMessageResponse(
			&bot.SendMessageParams{
				Text: msgs.Cancel(),
			},
		)
		return fsm.StateIdle, response, nil
	}
}
