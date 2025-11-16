package handlers

import (
	"context"
	"fmt"
	"whitelist/internal/fsm"
	"whitelist/internal/msgs"
	userRepo "whitelist/internal/repository/user"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handlers struct {
	repo userRepo.IUserRepository
}

func New(repo userRepo.IUserRepository) *Handlers {
	return &Handlers{repo: repo}
}

func (h *Handlers) Start(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
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

func (h *Handlers) Info(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
	user, err := h.repo.UserByTelegramID(update.Message.From.ID)
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

func (h *Handlers) Echo(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
	if update.Message.Text == "err" {
		return fsm.StateIdle, fmt.Errorf("test error")
	}

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	return fsm.StateIdle, err
}
