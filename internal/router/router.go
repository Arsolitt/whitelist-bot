package router

import (
	"context"
	"errors"
	"log/slog"
	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"
	"whitelist/internal/locker"
	"whitelist/internal/model"
	"whitelist/internal/repository"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

type HandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error)

type MatcherFunc func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool

type TelegramRouter struct {
	routes     []TelegramRoute
	fsm        fsm.IFSM
	repository repository.IRepository
	locker     locker.ILocker
}

func NewTelegramRouter(fsm fsm.IFSM, locker locker.ILocker, repository repository.IRepository) *TelegramRouter {
	return &TelegramRouter{
		routes:     make([]TelegramRoute, 0),
		fsm:        fsm,
		locker:     locker,
		repository: repository,
	}
}

type TelegramRoute struct {
	Matcher MatcherFunc
	Handler HandlerFunc
}

func (r *TelegramRouter) AddRoute(matcher MatcherFunc, handler HandlerFunc) {
	r.routes = append(r.routes, TelegramRoute{
		Matcher: matcher,
		Handler: handler,
	})
}

func (r *TelegramRouter) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
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
		slog.ErrorContext(ctx, "Failed to generate request ID", logger.ErrorField, err)
		return
	}
	ctx = logger.WithLogValue(ctx, logger.RequestIDField, requestID.String())

	correlationID, err := uuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate correlation ID", logger.ErrorField, err)
		return
	}
	ctx = logger.WithLogValue(ctx, logger.CorrelationIDField, correlationID.String())

	slog.InfoContext(ctx, "Handling update")

	user, err := r.repository.UserByTelegramID(update.Message.From.ID)
	if errors.Is(err, core.ErrUserNotFound) {
		slog.WarnContext(ctx, "User not found, creating new user")
		newUser, err := model.NewUser(
			update.Message.From.ID,
			"",
		)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create new user moder", logger.ErrorField, err)
			return
		}

		err = r.repository.CreateUser(newUser)
		if err != nil {
			slog.Error("Failed to create new user in repository", logger.ErrorField, err)
			return
		}

		user = newUser
	} else if err != nil {
		slog.Error("Failed to get user by telegram ID", logger.ErrorField, err)
		return
	}
	ctx = logger.WithLogValue(ctx, logger.UserIDField, user.ID.String())

	slog.DebugContext(ctx, "Trying to lock user")
	if err := r.locker.Lock(user.ID); err != nil {
		slog.ErrorContext(ctx, "Failed to lock user", logger.ErrorField, err)
		return
	}
	slog.DebugContext(ctx, "User locked")
	defer func() {
		if err := r.locker.Unlock(user.ID); err != nil {
			slog.ErrorContext(ctx, "Failed to unlock user", logger.ErrorField, err)
		}
	}()

	slog.DebugContext(ctx, "Trying to get user state")
	currentState, err := r.fsm.GetState(user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user state", logger.ErrorField, err)
		return
	}
	ctx = logger.WithLogValue(ctx, logger.CurrentStateField, currentState)
	slog.DebugContext(ctx, "User state got")

	for _, route := range r.routes {
		if route.Matcher(ctx, b, update, currentState) {
			nextState, err := route.Handler(ctx, b, update, currentState)
			ctx = logger.WithLogValue(ctx, logger.NextStateField, nextState)
			if err != nil {
				slog.Error("Failed to handle route", logger.ErrorField, err)
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Произошла ошибка при обработке команды",
				})
			}
			if nextState != currentState {
				if r.fsm.SetState(user.ID, nextState) != nil {
					slog.ErrorContext(ctx, "Failed to set user state", logger.ErrorField, err)
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   "Произошла ошибка при обработке команды",
					})
					return
				}
				ctx = logger.WithLogValue(ctx, logger.NextStateField, nextState)
				slog.DebugContext(ctx, "User state updated")
			}
			return
		}
	}
}
