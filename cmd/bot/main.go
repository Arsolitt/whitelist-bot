package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"
	memoryFSM "whitelist/internal/fsm/memory"
	"whitelist/internal/handlers"
	memoryLocker "whitelist/internal/locker/memory"
	memoryUserRepository "whitelist/internal/repository/user/memory"
	"whitelist/internal/router"
	"whitelist/internal/router/matcher"

	"github.com/go-telegram/bot"
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
	repositoryService := memoryUserRepository.NewMemoryUserRepository()
	mainRouter := router.NewTelegramRouter(fsmService, lockerService, repositoryService)

	h := handlers.New(repositoryService)

	mainRouter.AddRoute(
		matcher.Command("start"),
		h.Start,
		router.DurationMiddleware,
	)

	mainRouter.AddRoute(
		matcher.And(matcher.Command("info"), matcher.State(fsm.StateIdle)),
		h.Info,
		router.DurationMiddleware,
	)

	mainRouter.AddRoute(
		matcher.State(fsm.StateIdle),
		h.Echo,
		router.DurationMiddleware,
	)

	opts := []bot.Option{
		bot.WithDefaultHandler(mainRouter.Handle),
		bot.WithErrorsHandler(func(err error) {
			if strings.Contains(err.Error(), "context canceled") {
				slog.Info("Bot stopped")
				return
			}
			slog.Error("Bot error", "error", err.Error())
		}),
	}

	b, err := bot.New(cfg.Telegram.Token, opts...)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	b.Start(ctx)
}
