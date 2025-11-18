package handlers

import (
	"context"
	"errors"
	"whitelist/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) Echo(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
	if update.Message.Text == "err" {
		return fsm.StateIdle, errors.New("test error")
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	return fsm.StateIdle, err
}
