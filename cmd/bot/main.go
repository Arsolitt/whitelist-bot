package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"whitelist/internal/core"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"
	memoryFSM "whitelist/internal/fsm/memory"
	"whitelist/internal/handlers"
	memoryLocker "whitelist/internal/locker/memory"
	sqliteUserRepository "whitelist/internal/repository/user/sqlite"
	sqliteWLRequestRepository "whitelist/internal/repository/wl_request/sqlite"
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

	lockerService := memoryLocker.NewLocker()
	fsmService := memoryFSM.NewFSM()

	// TODO: move db connection to core package
	db, err := sql.Open("sqlite3", "data/whitelist.db")
	if err != nil {
		slog.Error("Failed to open database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	userRepo := sqliteUserRepository.NewUserRepository(db)
	wlRequestRepo := sqliteWLRequestRepository.NewWLRequestRepository(db)
	mainRouter := router.NewTelegramRouter(fsmService, lockerService, userRepo)

	mainRouter.UseMsgMiddleware(router.RecoverMsgMiddleware)
	mainRouter.UseMsgMiddleware(router.DurationMsgMiddleware)
	mainRouter.UseCallbackMiddleware(router.RecoverCallbackMiddleware)
	mainRouter.UseCallbackMiddleware(router.DurationCallbackMiddleware)

	h := handlers.New(userRepo, wlRequestRepo)

	mainRouter.AddMsgRoute(
		matcher.Command("start"),
		h.Start,
	)

	mainRouter.AddMsgRoute(
		matcher.And(matcher.Command("info"), matcher.State(fsm.StateIdle)),
		h.Info,
	)

	mainRouter.AddMsgRoute(
		matcher.And(matcher.Text("Новая заявка"), matcher.State(fsm.StateIdle)),
		h.NewWLRequest,
	)

	mainRouter.AddMsgRoute(
		matcher.And(
			matcher.Text("Посмотреть заявки"),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
			matcher.State(fsm.StateIdle),
		),
		h.HandlePendingWLRequest,
	)

	mainRouter.AddMsgRoute(
		matcher.State(fsm.StateWaitingWLNickname),
		h.HandleWLRequestNickname,
	)

	mainRouter.AddMsgRoute(
		matcher.State(fsm.StateIdle),
		h.Echo,
	)

	mainRouter.AddCallbackRoute(
		matcher.CallbackPrefix("approve"),
		h.HandleApproveWLRequest,
	)

	opts := []bot.Option{
		bot.WithDefaultHandler(mainRouter.HandleMsg),
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
	b.RegisterHandlerRegexp(bot.HandlerTypeCallbackQueryData, regexp.MustCompile("."), mainRouter.HandleCallback)

	// Register callback query handlers
	// b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "approve", bot.MatchTypePrefix, h.HandleApproveWLRequest)
	// b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "decline", bot.MatchTypePrefix, h.HandleDeclineWLRequest)

	b.Start(ctx)
}
