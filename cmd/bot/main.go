package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/db"
	"whitelist-bot/internal/core/kv"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	memoryFSM "whitelist-bot/internal/fsm/memory"
	"whitelist-bot/internal/handlers"
	memoryLocker "whitelist-bot/internal/locker/memory"
	"whitelist-bot/internal/router"
	"whitelist-bot/internal/router/matcher"

	postgresUserRepository "whitelist-bot/internal/repository/user/postgres"
	postgresWLRequestRepository "whitelist-bot/internal/repository/wl_request/postgres"
)

// TODO: add support to return multiple messages in one handler.
// TODO: write tests !!!!!!!!!!
// TODO: refactor FSM to store metadata for each state.
// TODO: add event system for notifications.
// TODO: add validation for nickname. Length, special characters, etc.
// TODO: refactor wl_request.go handlers for better readability.
// TODO: add emojis to messages.
// TODO: add middleware for checking permissions.
// TODO: add middleware for recovering panics.
// TODO: add requests limit for users.
// TODO: refactor to use Must methods for initialization.

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

	lockerService := memoryLocker.New()
	fsmService := memoryFSM.NewFSM()

	dbSqlite, err := db.GetSqliteDB(ctx, cfg.Sqlite.URL)
	if err != nil {
		slog.Error("Failed to open sqlite database", "error", err.Error())
		os.Exit(1)
	}
	defer dbSqlite.Close()

	dbPG, err := db.GetPostgresDB(ctx, cfg.Postgres.URL)
	if err != nil {
		slog.Error("Failed to connect to postgres database", "error", err.Error())
		os.Exit(1)
	}
	defer dbPG.Close()

	conn, err := kv.GetNatsConn(ctx, cfg.Nats)
	if err != nil {
		slog.Error("Failed to connect to NATS", "error", err.Error())
		os.Exit(1)
	}
	defer conn.Drain()

	userRepo := postgresUserRepository.NewUserRepository(dbPG)
	wlRequestRepo := postgresWLRequestRepository.NewWLRequestRepository(dbPG)

	// metastoreService, err := natsMetastore.New(ctx, conn, "whitelist-bot", cfg.Nats.MetastoreReplicas)
	if err != nil {
		slog.Error("Failed to create NATS metastore", "error", err.Error())
		os.Exit(1)
	}
	// h := handlers.New(userRepo, wlRequestRepo, metastoreService, cfg)
	r, err := router.NewTelegramRouter(fsmService,
		lockerService,
		userRepo,
		handlers.GlobalErrorHandler(),
		handlers.GlobalSuccessHandler(cfg),
		cfg.Telegram.Token,
		handlers.DefaultHandler(),
		func(err error) {
			if strings.Contains(err.Error(), "context canceled") {
				slog.Info("Bot stopped")
				return
			}
			slog.Error("Bot error", "error", err.Error())
		},
	)
	if err != nil {
		slog.Error("Failed to create telegram router", "error", err.Error())
		os.Exit(1)
	}

	// INFO HANDLER
	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.MsgText(core.CommandInfo),
			r.StateMatchFunc(fsm.StateIdle),
		),
		handlers.Info(userRepo),
	)

	// NEW WL REQUEST HANDLER
	r.RegisterHandlerMatchFunc(
		matcher.And(matcher.MsgText(core.CommandNewWLRequest), r.StateMatchFunc(fsm.StateIdle)),
		handlers.NewWLRequest(),
	)
	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.MsgText(core.CommandViewPendingWLRequests),
			r.StateMatchFunc(fsm.StateIdle),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.ViewPendingWLRequests(wlRequestRepo, r.Bot()),
	)
	r.RegisterHandlerMatchFunc(r.StateMatchFunc(fsm.StateWaitingWLNickname), handlers.SubmitWLRequestNickname(userRepo, wlRequestRepo))

	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.CallbackAction(core.ActionWLRequestApprove),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.ApproveWLRequest(userRepo, wlRequestRepo, r.Bot()))
	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.CallbackAction(core.ActionWLRequestDecline),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.DeclineWLRequest(userRepo, wlRequestRepo, r.Bot()))

	// r.RegisterHandlerMatchFunc(matcher.And(
	// 	r.StateMatchFunc(fsm.StateIdle),
	// 	matcher.MsgText(core.CommandAnketaStart),
	// ), h.AnketaStart)
	// r.RegisterHandlerMatchFunc(r.StateMatchFunc(fsm.StateAnketaName), h.AnketaName)
	// r.RegisterHandlerMatchFunc(r.StateMatchFunc(fsm.StateAnketaAge), h.AnketaAge)
	// r.RegisterHandlerMatchFunc(matcher.And(
	// 	r.StateMatchFunc(fsm.StateIdle),
	// 	matcher.MsgText(core.CommandAnketaInfo),
	// ), h.AnketaInfo)

	r.Start(ctx)
}
