package router

import (
	"context"
	"log/slog"
	"time"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

func DurationMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error) {
		start := time.Now()

		slog.DebugContext(ctx, "Duration middleware started")

		nextState, err := next(ctx, b, update, currentState)

		duration := time.Since(start)

		ctx = logger.WithLogValue(ctx, logger.DurationField, duration.String())

		slog.DebugContext(ctx, "Duration middleware completed")

		return nextState, err
	}
}

func ContextMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error) {
		ctx = logger.WithLogValue(ctx, logger.ChatIDField, update.Message.Chat.ID)
		ctx = logger.WithLogValue(ctx, logger.UserTelegramIDField, update.Message.From.ID)
		ctx = logger.WithLogValue(ctx, logger.UserNameField, update.Message.From.Username)
		ctx = logger.WithLogValue(ctx, logger.UserFirstNameField, update.Message.From.FirstName)
		ctx = logger.WithLogValue(ctx, logger.UserLastNameField, update.Message.From.LastName)
		ctx = logger.WithLogValue(ctx, logger.UpdateIDField, update.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageIDField, update.Message.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageChatIDField, update.Message.Chat.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageChatTypeField, update.Message.Chat.Type)

		requestID, err := uuid.NewV7()
		if err != nil {
			requestID = uuid.New()
		}
		ctx = logger.WithLogValue(ctx, logger.RequestIDField, requestID.String())

		correlationID, err := uuid.NewV7()
		if err != nil {
			correlationID = uuid.New()
		}
		ctx = logger.WithLogValue(ctx, logger.CorrelationIDField, correlationID.String())

		return next(ctx, b, update, currentState)

	}
}

func RecoverMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "Panic recovered in handler",
					slog.Any("panic", r),
				)
			}
		}()

		return next(ctx, b, update, currentState)
	}
}
