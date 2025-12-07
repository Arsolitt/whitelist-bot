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
	"whitelist-bot/internal/eventbus"
	"whitelist-bot/internal/fsm"
	memoryFSM "whitelist-bot/internal/fsm/memory"
	"whitelist-bot/internal/handlers"
	memoryLocker "whitelist-bot/internal/locker/memory"
	"whitelist-bot/internal/router"
	"whitelist-bot/internal/router/matcher"
	"whitelist-bot/internal/wp"

	bh "whitelist-bot/internal/eventbus/handlers"

	memoryEventBus "whitelist-bot/internal/eventbus/memory"
	natsMetastore "whitelist-bot/internal/metastore/nats"
	postgresUserRepository "whitelist-bot/internal/repository/user/postgres"
	postgresWLRequestRepository "whitelist-bot/internal/repository/wl_request/postgres"
)

// TODO: write tests !!!!!!!!!!
// TODO: add validation for nickname. Length, special characters, etc.
// TODO: add middleware for checking permissions.
// TODO: add middleware for recovering panics.
// TODO: add requests limit for users.
// TODO: refactor to use Must methods for initialization.
// TODO: add custom update context, set user to context.
// TODO: add wrapper for bot sending message methods. Retry logic, error handling, default parse mode.

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
	fsmService := memoryFSM.New()

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

	metastoreService, err := natsMetastore.New(ctx, conn, "whitelist-bot", cfg.Nats.MetastoreReplicas)
	if err != nil {
		slog.Error("Failed to create NATS metastore", "error", err.Error())
		os.Exit(1)
	}

	eBus := memoryEventBus.New(10)
	defer eBus.Close()

	sem, err := wp.NewSemaphore(10)
	if err != nil {
		slog.Error("Failed to create semaphore", "error", err.Error())
		os.Exit(1)
	}
	r, err := router.NewTelegramRouter(
		fsmService,
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
			r.StateMatchFunc(ctx, fsm.StateIdle),
		),
		handlers.Info(userRepo),
	)

	// NEW WL REQUEST HANDLER
	r.RegisterHandlerMatchFunc(
		matcher.And(matcher.MsgText(core.CommandNewWLRequest), r.StateMatchFunc(ctx, fsm.StateIdle)),
		handlers.NewWLRequest(),
	)
	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.MsgText(core.CommandViewPendingWLRequests),
			r.StateMatchFunc(ctx, fsm.StateIdle),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.ViewPendingWLRequests(wlRequestRepo),
	)
	r.RegisterHandlerMatchFunc(
		r.StateMatchFunc(ctx, fsm.StateWaitingWLNickname),
		handlers.SubmitWLRequestNickname(userRepo, wlRequestRepo, eBus),
	)

	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.CallbackAction(core.ActionWLRequestApprove),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.ApproveWLRequest(userRepo, wlRequestRepo))
	r.RegisterHandlerMatchFunc(
		matcher.And(
			matcher.CallbackAction(core.ActionWLRequestDecline),
			matcher.MatchTelegramIDs(cfg.Telegram.AdminIDs...),
		),
		handlers.DeclineWLRequest(userRepo, wlRequestRepo))

	consumerPool := eventbus.NewConsumerPool(eBus, []eventbus.ConsumerUnit{
		{
			Topic:   core.TopicWLRequestCreated,
			Handler: bh.HandleWLRequestCreatedEvent(metastoreService, metastoreService, r.Bot(), cfg.Telegram.AdminIDs),
		},
	}, sem)
	err = consumerPool.Start(ctx)
	if err != nil {
		slog.Error("Failed to start consumer pool", "error", err.Error())
		os.Exit(1)
	}

	r.Start(ctx)

	consumerPool.Wait()
}
