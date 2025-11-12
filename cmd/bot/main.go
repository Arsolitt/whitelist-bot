package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"
	memoryFSM "whitelist/internal/fsm/memory"
	memoryLocker "whitelist/internal/locker/memory"
	memoryRepository "whitelist/internal/repository/memory"
	"whitelist/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := core.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("Config loaded successfully")

	logger.InitLogger(cfg.Logs)
	slog.Info("Logger initialized successfully")

	lockerService := memoryLocker.NewMemoryLocker()
	fsmService := memoryFSM.NewMemoryFSM()
	repositoryService := memoryRepository.NewMemoryRepository()
	mainRouter := router.NewTelegramRouter(fsmService, lockerService, repositoryService)

	mainRouter.AddRoute(func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		return update.Message.Text == "/start"
	}, func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Пожалуйста, введите ваш никнейм",
		})
		return fsm.StateWaitingNickname, nil
	})

	mainRouter.AddRoute(func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		return state == fsm.StateWaitingNickname
	}, func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
		user, err := repositoryService.UserByTelegramID(update.Message.From.ID)
		if err != nil {
			return fsm.StateIdle, fmt.Errorf("failed to get user: %w", err)
		}

		user.CustomName = update.Message.Text
		err = repositoryService.UpdateUser(user)
		if err != nil {
			return fsm.StateIdle, fmt.Errorf("failed to update user: %w", err)
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Никнейм успешно обновлен",
		})

		return fsm.StateIdle, nil
	})

	mainRouter.AddRoute(func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		return state == fsm.StateIdle
	}, func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   update.Message.Text,
		})
		return fsm.StateIdle, nil
	})

	opts := []bot.Option{
		bot.WithDefaultHandler(mainRouter.Handle),
		bot.WithErrorsHandler(func(err error) {
			slog.Error("Bot error", "error", err)
		}),
	}

	b, err := bot.New(cfg.Telegram.Token, opts...)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	b.Start(ctx)
}
