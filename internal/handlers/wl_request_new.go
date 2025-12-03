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
	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
		msgParams := bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      msgs.WaitingForNickname(),
			ParseMode: "HTML",
		}
		return fsm.StateWaitingWLNickname, &msgParams, nil
	}
}
