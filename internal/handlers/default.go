package handlers

import (
	"context"
	"whitelist/internal/core"
	"whitelist/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
	return fsm.StateIdle, core.ErrUnknownCommand
}
