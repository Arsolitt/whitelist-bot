package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/core/utils"
	domainUser "whitelist-bot/internal/domain/user"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/locker"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, *bot.SendMessageParams, error)
type ErrorHandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, err error)

type SuccessHandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State, msgParams *bot.SendMessageParams)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	CreateUser(
		ctx context.Context,
		telegramId domainUser.TelegramID,
		firstName domainUser.FirstName,
		lastName domainUser.LastName,
		username domainUser.Username,
	) (domainUser.User, error)
}

type TelegramRouter struct {
	bot            *bot.Bot
	fsm            fsm.IFSM
	userRepository iUserRepository
	locker         locker.ILocker
	errorHandler   ErrorHandlerFunc
	successHandler SuccessHandlerFunc
}

func NewTelegramRouter(
	fsm fsm.IFSM,
	locker locker.ILocker,
	repository iUserRepository,
	errorHandler ErrorHandlerFunc,
	successHandler SuccessHandlerFunc,
) *TelegramRouter {
	return &TelegramRouter{
		fsm:            fsm,
		locker:         locker,
		userRepository: repository,
		errorHandler:   errorHandler,
		successHandler: successHandler,
	}
}

func (r *TelegramRouter) SetBot(b *bot.Bot) {
	r.bot = b
}

func (r *TelegramRouter) WrapHandler(handler HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var userID int64
		var userName, firstName, lastName string
		var chatID int64

		if update.Message != nil {
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
			userName = update.Message.From.Username
			firstName = update.Message.From.FirstName
			lastName = update.Message.From.LastName
			ctx = logger.WithLogValue(ctx, logger.MessageIDField, update.Message.ID)
			ctx = logger.WithLogValue(ctx, logger.MessageChatIDField, update.Message.Chat.ID)
			ctx = logger.WithLogValue(ctx, logger.MessageChatTypeField, update.Message.Chat.Type)
		} else if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Message.Chat.ID
			userID = update.CallbackQuery.From.ID
			userName = update.CallbackQuery.From.Username
			firstName = update.CallbackQuery.From.FirstName
			lastName = update.CallbackQuery.From.LastName
		}

		ctx = logger.WithLogValue(ctx, logger.ChatIDField, chatID)
		ctx = logger.WithLogValue(ctx, logger.UserTelegramIDField, userID)
		ctx = logger.WithLogValue(ctx, logger.UserNameField, userName)
		ctx = logger.WithLogValue(ctx, logger.UserFirstNameField, firstName)
		ctx = logger.WithLogValue(ctx, logger.UserLastNameField, lastName)
		ctx = logger.WithLogValue(ctx, logger.UpdateIDField, update.ID)
		ctx = logger.WithLogValue(ctx, logger.RequestIDField, utils.NewUniqueID().String())
		ctx = logger.WithLogValue(ctx, logger.CorrelationIDField, utils.NewUniqueID().String())
		slog.InfoContext(ctx, fmt.Sprintf("Handling update: %d", update.ID))

		user, err := r.checkUser(ctx, userID, firstName, lastName, userName)
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
		nextState, msgParams, err := handler(ctx, b, update, currentState)
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

		r.successHandler(ctx, b, update, nextState, msgParams)
	}
}

func (r *TelegramRouter) checkUser(
	ctx context.Context,
	id int64,
	firstName string,
	lastName string,
	username string,
) (domainUser.User, error) {
	user, repoErr := r.userRepository.UserByTelegramID(ctx, id)
	// TODO: add cache for user

	if errors.Is(repoErr, core.ErrUserNotFound) {
		slog.WarnContext(ctx, "User not found, creating new user")

		newDBUser, err := r.userRepository.CreateUser(
			ctx,
			domainUser.TelegramID(id),
			domainUser.FirstName(firstName),
			domainUser.LastName(lastName),
			domainUser.Username(username),
		)
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

func (r *TelegramRouter) StateMatchFunc(expectedState fsm.State) bot.MatchFunc {
	return func(update *models.Update) bool {
		if update.Message == nil || update.Message.From == nil {
			return false
		}

		user, err := r.userRepository.UserByTelegramID(context.Background(), update.Message.From.ID)
		if err != nil {
			return false
		}

		currentState, err := r.fsm.GetState(user.ID())
		if err != nil {
			return false
		}

		return currentState == expectedState
	}
}

func (r *TelegramRouter) RegisterHandlerMatchFunc(matcher bot.MatchFunc, handler HandlerFunc) {
	r.bot.RegisterHandlerMatchFunc(matcher, r.WrapHandler(handler))
}
