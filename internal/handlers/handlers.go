package handlers

import (
	"context"

	"whitelist-bot/internal/core"
	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
	"whitelist-bot/internal/metastore"
	repository "whitelist-bot/internal/repository/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type iUserGetter interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type iWLRequestRepository interface {
	CreateWLRequest(
		ctx context.Context,
		requesterID domainWLRequest.RequesterID,
		nickname domainWLRequest.Nickname,
	) (domainWLRequest.WLRequest, error)
	PendingWLRequests(ctx context.Context, limit int64) ([]domainWLRequest.WLRequest, error)
	PendingWLRequestsWithRequester(ctx context.Context, limit int64) ([]repository.PendingWLRequestWithRequester, error)
	WLRequestByID(ctx context.Context, id domainWLRequest.ID) (domainWLRequest.WLRequest, error)
	UpdateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error)
}

type iMessageSender interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	AnswerCallbackQuery(ctx context.Context, params *bot.AnswerCallbackQueryParams) (bool, error)
	EditMessageText(ctx context.Context, params *bot.EditMessageTextParams) (*models.Message, error)
}

type Handlers struct {
	userRepo      iUserRepository
	wlRequestRepo iWLRequestRepository
	metastore     metastore.Metastore
	config        core.Config
}

func New(userRepo iUserRepository, wlRequestRepo iWLRequestRepository, metastore metastore.Metastore, config core.Config) *Handlers {
	return &Handlers{userRepo: userRepo, wlRequestRepo: wlRequestRepo, metastore: metastore, config: config}
}
