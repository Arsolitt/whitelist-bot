package handlers

import (
	"context"

	domainUser "whitelist/internal/domain/user"
	domainWLRequest "whitelist/internal/domain/wl_request"
)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
}

type iWLRequestRepository interface {
	CreateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error)
	PendingWLRequest(ctx context.Context) (domainWLRequest.WLRequest, error)
}

type Handlers struct {
	useRepo       iUserRepository
	wlRequestRepo iWLRequestRepository
}

func New(useRepo iUserRepository, wlRequestRepo iWLRequestRepository) *Handlers {
	return &Handlers{useRepo: useRepo, wlRequestRepo: wlRequestRepo}
}
