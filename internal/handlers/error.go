package handlers

import (
	"context"
	"errors"
	"log/slog"
	"whitelist/internal/core"
	"whitelist/internal/core/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	ErrUnknownCommandMessage   = "Неизвестная команда"
	ErrInternalErrorMessage    = "Произошла ошибка при обработке команды"
	ErrInvalidUserStateMessage = "Неверное состояние пользователя"
)

var errorStatusMap = map[error]string{
	core.ErrUnknownCommand:   ErrUnknownCommandMessage,
	core.ErrInvalidUserState: ErrInvalidUserStateMessage,
}

func (h *Handlers) GlobalErrorHandler(ctx context.Context, b *bot.Bot, update *models.Update, err error) {
	slog.ErrorContext(ctx, "Failed to handle update", logger.ErrorField, err.Error())
	switch {
	case h.getCustomErrorMessage(err) != "":
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   h.getCustomErrorMessage(err),
		})
	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   ErrInternalErrorMessage,
		})
	}

}

func (h *Handlers) getCustomErrorMessage(target error) string {
	for err, message := range errorStatusMap {
		if errors.Is(target, err) {
			return message
		}
	}
	return ""
}
