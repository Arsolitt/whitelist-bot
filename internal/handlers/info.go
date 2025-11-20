package handlers

import (
	"context"
	"fmt"
	"whitelist/internal/core"
	"whitelist/internal/fsm"
	"whitelist/internal/msgs"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) Info(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error) {
	if currentState != fsm.StateIdle {
		return currentState, core.ErrInvalidUserState
	}

	user, err := h.useRepo.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		return fsm.StateIdle, fmt.Errorf("failed to get user: %w", err)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.UserInfo(user),
		ParseMode: "HTML",
	})
	return fsm.StateIdle, err
}
