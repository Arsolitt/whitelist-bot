package handlers

import (
	"context"

	"whitelist/internal/core"
	domainUser "whitelist/internal/domain/user"
	domainWLRequest "whitelist/internal/domain/wl_request"
)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
}

type iWLRequestRepository interface {
	CreateWLRequest(ctx context.Context, requesterID domainWLRequest.RequesterID, nickname domainWLRequest.Nickname) (domainWLRequest.WLRequest, error)
	PendingWLRequests(ctx context.Context, limit int64) ([]domainWLRequest.WLRequest, error)
	WLRequestByID(ctx context.Context, id domainWLRequest.ID) (domainWLRequest.WLRequest, error)
	UpdateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error)
}

type Handlers struct {
	useRepo       iUserRepository
	wlRequestRepo iWLRequestRepository
	config        core.Config
}

func New(useRepo iUserRepository, wlRequestRepo iWLRequestRepository, config core.Config) *Handlers {
	return &Handlers{useRepo: useRepo, wlRequestRepo: wlRequestRepo, config: config}
}
