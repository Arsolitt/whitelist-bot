package handlers

import (
	"context"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DefaultHandler() router.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, router.Response, error) {
		return fsm.StateIdle, nil, core.ErrUnknownCommand
	}
}
