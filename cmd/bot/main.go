package main

import (
	"context"
	"database/sql"
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
	"whitelist/internal/repository/user/sqlite"
	"whitelist/internal/router"
	"whitelist/internal/router/matcher"

	"github.com/go-telegram/bot"

	_ "github.com/mattn/go-sqlite3"
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
	// repositoryService := memoryUserRepository.NewMemoryUserRepository()

	db, err := sql.Open("sqlite3", "data/whitelist.db")
	if err != nil {
		slog.Error("Failed to open database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	repositoryService := sqlite.NewSQLiteUserRepository(db)
	mainRouter := router.NewTelegramRouter(fsmService, lockerService, repositoryService)

	mainRouter.Use(router.RecoverMiddleware)
	mainRouter.Use(router.ContextMiddleware)
	mainRouter.Use(router.DurationMiddleware)

	h := handlers.New(repositoryService)

	mainRouter.AddRoute(
		matcher.Command("start"),
		h.Start,
	)

	mainRouter.AddRoute(
		matcher.And(matcher.Command("info"), matcher.State(fsm.StateIdle)),
		h.Info,
	)

	mainRouter.AddRoute(
		matcher.State(fsm.StateIdle),
		h.Echo,
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
