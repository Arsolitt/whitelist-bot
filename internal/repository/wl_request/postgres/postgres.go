package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	domainWLRequest "whitelist-bot/internal/domain/wl_request"
)

type iQueryable interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
}

type WLRequestRepository struct {
	db iQueryable
}

func NewWLRequestRepository(db iQueryable) *WLRequestRepository {
	return &WLRequestRepository{db: db}
}

func (r *WLRequestRepository) CreateWLRequest(
	ctx context.Context,
	requesterID domainWLRequest.RequesterID,
	nickname domainWLRequest.Nickname,
) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	now := time.Now()

	newWLRequest, err := domainWLRequest.NewBuilder().
		NewID().
		Status(domainWLRequest.StatusPending).
		DeclineReasonFromString("").
		RequesterID(requesterID).
		Nickname(nickname).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}

	_, err = q.CreateWLRequest(ctx, CreateWLRequestParams{
		ID:            newWLRequest.ID(),
		RequesterID:   newWLRequest.RequesterID(),
		Nickname:      newWLRequest.Nickname(),
		Status:        newWLRequest.Status(),
		DeclineReason: newWLRequest.DeclineReason(),
		CreatedAt:     newWLRequest.CreatedAt(),
		UpdatedAt:     newWLRequest.UpdatedAt(),
	})
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to create wl request: %w", err)
	}
	return newWLRequest, nil
}

func (r *WLRequestRepository) PendingWLRequests(ctx context.Context, limit int64) ([]domainWLRequest.WLRequest, error) {
	q := New(r.db)

	dbWLRequests, err := q.PendingWLRequests(ctx, limit)
	pendingWLRequests := make([]domainWLRequest.WLRequest, len(dbWLRequests))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pendingWLRequests, nil
		}
		return nil, fmt.Errorf("failed to get pending wl requests: %w", err)
	}
	for i, dbWLRequest := range dbWLRequests {
		builder := domainWLRequest.NewBuilder().
			ID(dbWLRequest.ID).
			Status(dbWLRequest.Status).
			DeclineReason(dbWLRequest.DeclineReason).
			RequesterID(dbWLRequest.RequesterID).
			Nickname(dbWLRequest.Nickname).
			CreatedAt(dbWLRequest.CreatedAt).
			UpdatedAt(dbWLRequest.UpdatedAt)

		if !dbWLRequest.ArbiterID.IsZero() {
			builder = builder.ArbiterID(dbWLRequest.ArbiterID)
		}

		pendingWLRequests[i], err = builder.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build wl request: %s: %w", dbWLRequest.ID, err)
		}
	}
	return pendingWLRequests, nil
}

func (r *WLRequestRepository) WLRequestByID(
	ctx context.Context,
	id domainWLRequest.ID,
) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	dbWLRequest, err := q.WLRequestByID(ctx, id)
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to get wl request by id: %w", err)
	}

	builder := domainWLRequest.NewBuilder().
		ID(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterID(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAt(dbWLRequest.CreatedAt).
		UpdatedAt(dbWLRequest.UpdatedAt)

	if !dbWLRequest.ArbiterID.IsZero() {
		builder = builder.ArbiterID(dbWLRequest.ArbiterID)
	}

	wlRequest, err := builder.Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}

	return wlRequest, nil
}

func (r *WLRequestRepository) UpdateWLRequest(
	ctx context.Context,
	wlRequest domainWLRequest.WLRequest,
) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	wlRequest = wlRequest.UpdateTimestamp()

	_, err := q.UpdateWLRequest(ctx, UpdateWLRequestParams{
		ID:            wlRequest.ID(),
		RequesterID:   wlRequest.RequesterID(),
		Nickname:      wlRequest.Nickname(),
		Status:        wlRequest.Status(),
		DeclineReason: wlRequest.DeclineReason(),
		ArbiterID:     wlRequest.ArbiterID(),
		UpdatedAt:     wlRequest.UpdatedAt(),
	})
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to update wl request: %w", err)
	}

	return wlRequest, nil
}
