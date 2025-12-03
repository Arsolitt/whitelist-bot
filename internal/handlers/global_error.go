package handlers

import (
	"context"
	"errors"
	"log/slog"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/logger"

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

func GlobalErrorHandler() func(ctx context.Context, b *bot.Bot, update *models.Update, err error) {
	getCustomErrorMessage := func(target error) string {
		for err, message := range errorStatusMap {
			if errors.Is(target, err) {
				return message
			}
		}
		return ""
	}

	return func(ctx context.Context, b *bot.Bot, update *models.Update, err error) {
		slog.ErrorContext(ctx, "Failed to handle update", logger.ErrorField, err.Error())
		switch {
		case getCustomErrorMessage(err) != "" && update.Message != nil:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   getCustomErrorMessage(err),
			})
		case update.Message != nil:
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   ErrInternalErrorMessage,
			})
		default:
			return
		}
	}
}
