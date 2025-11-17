package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	domainUser "whitelist/internal/domain/user"
	"whitelist/internal/fsm"
	"whitelist/internal/locker"
	userRepo "whitelist/internal/repository/user"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error)

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type MatcherFunc func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool

type TelegramRouter struct {
	routes            []TelegramRoute
	globalMiddlewares []MiddlewareFunc
	fsm               fsm.IFSM
	userRepository    userRepo.IUserRepository
	locker            locker.ILocker
}

func NewTelegramRouter(fsm fsm.IFSM, locker locker.ILocker, repository userRepo.IUserRepository) *TelegramRouter {
	return &TelegramRouter{
		routes:            make([]TelegramRoute, 0),
		globalMiddlewares: make([]MiddlewareFunc, 0),
		fsm:               fsm,
		locker:            locker,
		userRepository:    repository,
	}
}

type TelegramRoute struct {
	Matcher     MatcherFunc
	Handler     HandlerFunc
	Middlewares []MiddlewareFunc
}

func (r *TelegramRouter) Use(middlewares ...MiddlewareFunc) {
	r.globalMiddlewares = append(r.globalMiddlewares, middlewares...)
}

func (r *TelegramRouter) AddRoute(matcher MatcherFunc, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	r.routes = append(r.routes, TelegramRoute{
		Matcher:     matcher,
		Handler:     handler,
		Middlewares: middlewares,
	})
}

func (r *TelegramRouter) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := func() error {
		slog.InfoContext(ctx, "Handling update")

		rootHandler := func(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
			return r.executeRouting(ctx, b, update)
		}

		for i := len(r.globalMiddlewares) - 1; i >= 0; i-- {
			rootHandler = r.globalMiddlewares[i](rootHandler)
		}

		_, err := rootHandler(ctx, b, update, "")
		return err
	}()

	if err != nil {
		slog.ErrorContext(ctx, "Failed to handle update", logger.ErrorField, err.Error())
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Произошла ошибка при обработке команды",
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to send error message", logger.ErrorField, err)
		}
		return
	}
}

func (r *TelegramRouter) executeRouting(ctx context.Context, b *bot.Bot, update *models.Update) (fsm.State, error) {
	user, repoErr := r.userRepository.UserByTelegramID(ctx, update.Message.From.ID)
	// TODO: add in-memory LRU cache for user uuid

	if errors.Is(repoErr, core.ErrUserNotFound) {
		slog.WarnContext(ctx, "User not found, creating new user")

		newUser, err := domainUser.NewBuilder().
			NewID().
			TelegramID(update.Message.From.ID).
			FirstName(update.Message.From.FirstName).
			LastName(update.Message.From.LastName).
			Username(update.Message.From.Username).
			CreatedAt(time.Time{}).
			UpdatedAt(time.Time{}).
			Build()

		if err != nil {
			return "", fmt.Errorf("failed to create new user model: %w", err)
		}

		err = r.userRepository.CreateUser(ctx, newUser)
		if err != nil {
			return "", fmt.Errorf("failed to create new user in storage: %w", err)
		}

		user = newUser
	} else if repoErr != nil {
		return "", fmt.Errorf("failed to get user by telegram ID: %w", repoErr)
	}
	ctx = logger.WithLogValue(ctx, logger.UserIDField, user.ID().String())

	slog.DebugContext(ctx, "Trying to lock user")
	if err := r.locker.Lock(user.ID()); err != nil {
		return "", fmt.Errorf("failed to lock user: %w", err)
	}
	slog.DebugContext(ctx, "User locked")
	defer func() {
		if err := r.locker.Unlock(user.ID()); err != nil {
			slog.ErrorContext(ctx, "Failed to unlock user", logger.ErrorField, err)
		}
		slog.DebugContext(ctx, "User unlocked")
	}()

	slog.DebugContext(ctx, "Trying to get user state")
	currentState, err := r.fsm.GetState(user.ID())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user state", logger.ErrorField, err)
		return "", fmt.Errorf("failed to get user state: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.CurrentStateField, currentState)
	slog.DebugContext(ctx, "User state got")

	for _, route := range r.routes {
		if route.Matcher(ctx, b, update, currentState) {
			handler := route.Handler

			for i := len(route.Middlewares) - 1; i >= 0; i-- {
				handler = route.Middlewares[i](handler)
			}

			nextState, err := handler(ctx, b, update, currentState)
			//nolint:fatcontext // loop stop after first matched route
			ctx = logger.WithLogValue(ctx, logger.NextStateField, nextState)

			if nextState != currentState {
				if r.fsm.SetState(user.ID(), nextState) != nil {
					return "", fmt.Errorf("failed to set user state: %w", err)
				}
				ctx = logger.WithLogValue(ctx, logger.NextStateField, nextState)
				slog.DebugContext(ctx, "User state updated")
			}
			if err != nil {
				return "", fmt.Errorf("failed to handle route: %w", err)
			}
			return nextState, nil
		}
	}
	return currentState, nil
}
