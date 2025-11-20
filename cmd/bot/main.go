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
	sqliteUserRepository "whitelist/internal/repository/user/sqlite"
	sqliteWLRequestRepository "whitelist/internal/repository/wl_request/sqlite"
	"whitelist/internal/router"
	"whitelist/internal/router/matcher"

	"github.com/go-telegram/bot"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: refactor FSM to store json metadata for each state.
// TODO: add scheduler for checking pending wl requests and sending notifications to admins.
// TODO: add notification to user when their wl request is approved or declined.
// TODO: add validation for nickname. Length, special characters, etc.
// TODO: refactor wl_request.go handlers for better readability.
// TODO: add emojis to messages.
// TODO: add middleware for checking permissions.
// TODO: add middleware for recovering panics.
// TODO: add requests limit for users.

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
	h := handlers.New(userRepo, wlRequestRepo, cfg)
	r := router.NewTelegramRouter(fsmService, lockerService, userRepo, h.GlobalErrorHandler, h.GlobalSuccessHandler)

	opts := []bot.Option{
		bot.WithDefaultHandler(r.WrapHandler(h.DefaultHandler)),
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

	r.SetBot(b)

	r.RegisterHandlerMatchFunc(matcher.And(matcher.MsgText(core.CommandInfo), r.StateMatchFunc(fsm.StateIdle)), h.Info)
	r.RegisterHandlerMatchFunc(matcher.And(matcher.MsgText(core.CommandNewWLRequest), r.StateMatchFunc(fsm.StateIdle)), h.NewWLRequest)
	r.RegisterHandlerMatchFunc(matcher.And(matcher.MsgText(core.CommandViewPendingWLRequests), r.StateMatchFunc(fsm.StateIdle), matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...)), h.ViewPendingWLRequests)
	r.RegisterHandlerMatchFunc(r.StateMatchFunc(fsm.StateWaitingWLNickname), h.SubmitWLRequestNickname)

	r.RegisterHandlerMatchFunc(matcher.And(matcher.CallbackPrefix("approve"), matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...)), h.ApproveWLRequest)
	r.RegisterHandlerMatchFunc(matcher.And(matcher.CallbackPrefix("decline"), matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...)), h.DeclineWLRequest)

	b.Start(ctx)
}
