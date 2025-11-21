package handlers

import (
	"context"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, *bot.SendMessageParams, error) {
	return fsm.StateIdle, nil, core.ErrUnknownCommand
}
