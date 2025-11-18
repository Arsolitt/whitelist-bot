package handlers

import (
	"context"
	"errors"
	"fmt"
	"whitelist/internal/fsm"
	"whitelist/internal/msgs"

	domainUser "whitelist/internal/domain/user"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
}

type Handlers struct {
	repo iUserRepository
}

func New(repo iUserRepository) *Handlers {
	return &Handlers{repo: repo}
}

func (h *Handlers) Start(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Вы зарегистрированы в боте!",
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{Text: "/info"},
				},
			},
			ResizeKeyboard: true,
		},
	})
	return fsm.StateIdle, err
}

func (h *Handlers) Info(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
	user, err := h.repo.UserByTelegramID(ctx, update.Message.From.ID)
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
