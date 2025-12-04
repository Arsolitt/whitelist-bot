package handlers

import (
	"context"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TODO: rewrite routing for callback queries.
func NewWLRequest() router.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, router.Response, error) {
		response := router.NewMessageResponse(
			&bot.SendMessageParams{
				Text: msgs.WaitingForNickname(),
			},
		)
		return fsm.StateWaitingWLNickname, response, nil
	}
}
