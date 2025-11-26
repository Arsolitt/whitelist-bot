package handlers

import (
	"context"
	"fmt"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) Info(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	currentState fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	if currentState != fsm.StateIdle {
		return currentState, nil, core.ErrInvalidUserState
	}

	user, err := h.useRepo.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		return fsm.StateIdle, nil, fmt.Errorf("failed to get user: %w", err)
	}

	msgParams := &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.UserInfo(user),
		ParseMode: "HTML",
	}
	return fsm.StateIdle, msgParams, nil
}
