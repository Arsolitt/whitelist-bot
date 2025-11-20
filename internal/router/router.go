package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	"whitelist/internal/core/utils"
	domainUser "whitelist/internal/domain/user"
	"whitelist/internal/fsm"
	"whitelist/internal/locker"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error)
type ErrorHandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, err error)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	CreateUser(ctx context.Context, user domainUser.User) (domainUser.User, error)
}

type TelegramRouter struct {
	fsm            fsm.IFSM
	userRepository iUserRepository
	locker         locker.ILocker
	errorHandler   ErrorHandlerFunc
}

func NewTelegramRouter(fsm fsm.IFSM, locker locker.ILocker, repository iUserRepository, errorHandler ErrorHandlerFunc) *TelegramRouter {
	return &TelegramRouter{
		fsm:            fsm,
		locker:         locker,
		userRepository: repository,
		errorHandler:   errorHandler,
	}
}

func (r *TelegramRouter) WrapHandler(handler HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		ctx = logger.WithLogValue(ctx, logger.ChatIDField, update.Message.Chat.ID)
		ctx = logger.WithLogValue(ctx, logger.UserTelegramIDField, update.Message.From.ID)
		ctx = logger.WithLogValue(ctx, logger.UserNameField, update.Message.From.Username)
		ctx = logger.WithLogValue(ctx, logger.UserFirstNameField, update.Message.From.FirstName)
		ctx = logger.WithLogValue(ctx, logger.UserLastNameField, update.Message.From.LastName)
		ctx = logger.WithLogValue(ctx, logger.UpdateIDField, update.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageIDField, update.Message.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageChatIDField, update.Message.Chat.ID)
		ctx = logger.WithLogValue(ctx, logger.MessageChatTypeField, update.Message.Chat.Type)
		ctx = logger.WithLogValue(ctx, logger.RequestIDField, utils.NewUniqueID().String())
		ctx = logger.WithLogValue(ctx, logger.CorrelationIDField, utils.NewUniqueID().String())
		slog.InfoContext(ctx, fmt.Sprintf("Handling update: %d", update.ID))

		user, err := r.checkUser(ctx, update.Message.From.ID, update.Message.From.FirstName, update.Message.From.LastName, update.Message.From.Username)
		if err != nil {
			r.errorHandler(ctx, b, update, fmt.Errorf("failed to check user: %w", err))
			return
		}

		slog.DebugContext(ctx, "Trying to lock user")
		if err := r.locker.Lock(user.ID()); err != nil {
			r.errorHandler(ctx, b, update, fmt.Errorf("failed to lock user: %w", err))
			return
		}
		slog.DebugContext(ctx, "User locked")
		defer r.locker.Unlock(user.ID())

		slog.DebugContext(ctx, "Trying to get user state")
		currentState, err := r.fsm.GetState(user.ID())
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get user state", logger.ErrorField, err)
			r.errorHandler(ctx, b, update, fmt.Errorf("failed to get user state: %w", err))
			return
		}
		ctx = logger.WithLogValue(ctx, logger.CurrentStateField, currentState)
		slog.DebugContext(ctx, "User state got")
		nextState, err := handler(ctx, b, update, currentState)
		if err != nil {
			r.errorHandler(ctx, b, update, fmt.Errorf("failed to handle route: %w", err))
			return
		}
		if nextState != currentState {
			if r.fsm.SetState(user.ID(), nextState) != nil {
				r.errorHandler(ctx, b, update, fmt.Errorf("failed to set user state: %w", err))
				return
			}
			ctx = logger.WithLogValue(ctx, logger.NextStateField, nextState)
			slog.DebugContext(ctx, "User state updated")
		}
	}
}

func (r *TelegramRouter) checkUser(ctx context.Context, id int64, firstName string, lastName string, username string) (domainUser.User, error) {
	user, repoErr := r.userRepository.UserByTelegramID(ctx, id)
	// TODO: add cache for user

	if errors.Is(repoErr, core.ErrUserNotFound) {
		slog.WarnContext(ctx, "User not found, creating new user")

		newUser, err := domainUser.NewBuilder().
			NewID().
			TelegramIDFromInt(id).
			FirstNameFromString(firstName).
			LastNameFromString(lastName).
			UsernameFromString(username).
			Build()

		if err != nil {
			return domainUser.User{}, fmt.Errorf("failed to create new user model: %w", err)
		}

		newDBUser, err := r.userRepository.CreateUser(ctx, newUser)
		if err != nil {
			return domainUser.User{}, fmt.Errorf("failed to create new user in storage: %w", err)
		}

		user = newDBUser
	} else if repoErr != nil {
		return domainUser.User{}, fmt.Errorf("failed to get user by telegram ID: %w", repoErr)
	}
	logger.WithLogValue(ctx, logger.UserIDField, user.ID().String())
	return user, nil
}
