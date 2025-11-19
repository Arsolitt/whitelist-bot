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

type MsgHandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, error)
type CallbackHandlerFunc func(ctx context.Context, b *bot.Bot, update *models.Update) error

type MsgMiddlewareFunc func(next MsgHandlerFunc) MsgHandlerFunc
type CallbackMiddlewareFunc func(next CallbackHandlerFunc) CallbackHandlerFunc

type MsgMatcherFunc func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool
type CallbackMatcherFunc func(ctx context.Context, b *bot.Bot, update *models.Update) bool

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	CreateUser(ctx context.Context, user domainUser.User) (domainUser.User, error)
}

type TelegramRouter struct {
	msgRoutes                 []msgRoute
	callbackRoutes            []callbackRoute
	globalMsgMiddlewares      []MsgMiddlewareFunc
	globalCallbackMiddlewares []CallbackMiddlewareFunc
	fsm                       fsm.IFSM
	userRepository            iUserRepository
	locker                    locker.ILocker
}

func NewTelegramRouter(fsm fsm.IFSM, locker locker.ILocker, repository iUserRepository) *TelegramRouter {
	return &TelegramRouter{
		msgRoutes:                 make([]msgRoute, 0),
		callbackRoutes:            make([]callbackRoute, 0),
		globalMsgMiddlewares:      make([]MsgMiddlewareFunc, 0),
		globalCallbackMiddlewares: make([]CallbackMiddlewareFunc, 0),
		fsm:                       fsm,
		locker:                    locker,
		userRepository:            repository,
	}
}

type msgRoute struct {
	matcher     MsgMatcherFunc
	handler     MsgHandlerFunc
	middlewares []MsgMiddlewareFunc
}

type callbackRoute struct {
	matcher     CallbackMatcherFunc
	handler     CallbackHandlerFunc
	middlewares []CallbackMiddlewareFunc
}

func (r *TelegramRouter) UseMsgMiddleware(middlewares ...MsgMiddlewareFunc) {
	r.globalMsgMiddlewares = append(r.globalMsgMiddlewares, middlewares...)
}

func (r *TelegramRouter) UseCallbackMiddleware(middlewares ...CallbackMiddlewareFunc) {
	r.globalCallbackMiddlewares = append(r.globalCallbackMiddlewares, middlewares...)
}

func (r *TelegramRouter) AddMsgRoute(matcher MsgMatcherFunc, handler MsgHandlerFunc, middlewares ...MsgMiddlewareFunc) {
	r.msgRoutes = append(r.msgRoutes, msgRoute{
		matcher:     matcher,
		handler:     handler,
		middlewares: middlewares,
	})
}

func (r *TelegramRouter) AddCallbackRoute(matcher CallbackMatcherFunc, handler CallbackHandlerFunc, middlewares ...CallbackMiddlewareFunc) {
	r.callbackRoutes = append(r.callbackRoutes, callbackRoute{
		matcher:     matcher,
		handler:     handler,
		middlewares: middlewares,
	})
}

func (r *TelegramRouter) HandleMsg(ctx context.Context, b *bot.Bot, update *models.Update) {

	err := func() error {
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

		messageHandler := func(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
			return r.messageRouting(ctx, b, update)
		}

		for i := len(r.globalMsgMiddlewares) - 1; i >= 0; i-- {
			messageHandler = r.globalMsgMiddlewares[i](messageHandler)
		}

		_, err := messageHandler(ctx, b, update, "")
		return err
	}()

	slog.InfoContext(ctx, fmt.Sprintf("Update handled: %d", update.ID))
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

func (r *TelegramRouter) HandleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := func() error {
		ctx = logger.WithLogValue(ctx, logger.ChatIDField, update.CallbackQuery.ChatInstance)
		ctx = logger.WithLogValue(ctx, logger.UserTelegramIDField, update.CallbackQuery.From.ID)
		ctx = logger.WithLogValue(ctx, logger.UserNameField, update.CallbackQuery.From.Username)
		ctx = logger.WithLogValue(ctx, logger.UserFirstNameField, update.CallbackQuery.From.FirstName)
		ctx = logger.WithLogValue(ctx, logger.UserLastNameField, update.CallbackQuery.From.LastName)
		ctx = logger.WithLogValue(ctx, logger.UpdateIDField, update.ID)
		ctx = logger.WithLogValue(ctx, logger.RequestIDField, utils.NewUniqueID().String())
		// ctx = logger.WithLogValue(ctx, logger.CorrelationIDField, utils.NewUniqueID().String())
		slog.InfoContext(ctx, fmt.Sprintf("Handling callback query: %d", update.ID))

		callbackHandler := func(ctx context.Context, b *bot.Bot, update *models.Update) error {
			return r.callbackRouting(ctx, b, update)
		}

		for i := len(r.globalCallbackMiddlewares) - 1; i >= 0; i-- {
			callbackHandler = r.globalCallbackMiddlewares[i](callbackHandler)
		}

		err := callbackHandler(ctx, b, update)
		return err
	}()

	slog.InfoContext(ctx, fmt.Sprintf("Callback query handled: %d", update.ID))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to handle update", logger.ErrorField, err.Error())
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: произошла ошибка при обработке запроса",
			ShowAlert:       true,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to send error message", logger.ErrorField, err)
		}
		return
	}
}

func (r *TelegramRouter) messageRouting(ctx context.Context, b *bot.Bot, update *models.Update) (fsm.State, error) {
	user, err := r.checkUser(ctx, update.Message.From.ID, update.Message.From.FirstName, update.Message.From.LastName, update.Message.From.Username)
	if err != nil {
		return "", fmt.Errorf("failed to check user: %w", err)
	}

	slog.DebugContext(ctx, "Trying to lock user")
	if err := r.locker.Lock(user.ID()); err != nil {
		return "", fmt.Errorf("failed to lock user: %w", err)
	}
	slog.DebugContext(ctx, "User locked")
	defer r.locker.Unlock(user.ID())

	slog.DebugContext(ctx, "Trying to get user state")
	currentState, err := r.fsm.GetState(user.ID())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user state", logger.ErrorField, err)
		return "", fmt.Errorf("failed to get user state: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.CurrentStateField, currentState)
	slog.DebugContext(ctx, "User state got")

	for _, route := range r.msgRoutes {
		if route.matcher(ctx, b, update, currentState) {
			handler := route.handler

			for i := len(route.middlewares) - 1; i >= 0; i-- {
				handler = route.middlewares[i](handler)
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

func (r *TelegramRouter) callbackRouting(ctx context.Context, b *bot.Bot, update *models.Update) error {
	// user, err := r.checkUser(ctx, update.CallbackQuery.From.ID, update.CallbackQuery.From.FirstName, update.CallbackQuery.From.LastName, update.CallbackQuery.From.Username)
	// if err != nil {
	// 	return fmt.Errorf("failed to check user: %w", err)
	// }

	// slog.DebugContext(ctx, "Trying to lock user")
	// if err := r.locker.Lock(user.ID()); err != nil {
	// 	return "", fmt.Errorf("failed to lock user: %w", err)
	// }
	// slog.DebugContext(ctx, "User locked")
	// defer r.locker.Unlock(user.ID())

	// slog.DebugContext(ctx, "Trying to get user state")
	// currentState, err := r.fsm.GetState(user.ID())
	// if err != nil {
	// 	slog.ErrorContext(ctx, "Failed to get user state", logger.ErrorField, err)
	// 	return "", fmt.Errorf("failed to get user state: %w", err)
	// }
	// ctx = logger.WithLogValue(ctx, logger.CurrentStateField, currentState)
	// slog.DebugContext(ctx, "User state got")

	for _, route := range r.callbackRoutes {
		if route.matcher(ctx, b, update) {
			handler := route.handler

			for i := len(route.middlewares) - 1; i >= 0; i-- {
				handler = route.middlewares[i](handler)
			}

			err := handler(ctx, b, update)

			if err != nil {
				return fmt.Errorf("failed to handle route: %w", err)
			}
			return nil
		}
	}
	return nil
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
