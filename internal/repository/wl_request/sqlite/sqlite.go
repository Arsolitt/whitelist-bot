package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
)

const SQLITE_TIME_FORMAT = "2006-01-02T15:04:05-0700"

type iQueryable interface {
	Begin() (*sql.Tx, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
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
		ID:            newWLRequest.ID().String(),
		RequesterID:   newWLRequest.RequesterID().String(),
		Nickname:      newWLRequest.Nickname(),
		Status:        newWLRequest.Status(),
		DeclineReason: newWLRequest.DeclineReason(),
		CreatedAt:     newWLRequest.CreatedAt().Format(SQLITE_TIME_FORMAT),
		UpdatedAt:     newWLRequest.UpdatedAt().Format(SQLITE_TIME_FORMAT),
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
		createdAt, err := time.Parse(SQLITE_TIME_FORMAT, dbWLRequest.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse createdAt: %w", err)
		}
		updatedAt, err := time.Parse(SQLITE_TIME_FORMAT, dbWLRequest.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updatedAt: %w", err)
		}
		builder := domainWLRequest.NewBuilder().
			IDFromString(dbWLRequest.ID).
			Status(dbWLRequest.Status).
			DeclineReason(dbWLRequest.DeclineReason).
			RequesterIDFromString(dbWLRequest.RequesterID).
			Nickname(dbWLRequest.Nickname).
			CreatedAt(createdAt).
			UpdatedAt(updatedAt)

		if dbWLRequest.ArbiterID != "" {
			builder = builder.ArbiterIDFromString(dbWLRequest.ArbiterID)
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

	dbWLRequest, err := q.WLRequestByID(ctx, id.String())
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to get wl request by id: %w", err)
	}

	createdAt, err := time.Parse(SQLITE_TIME_FORMAT, dbWLRequest.CreatedAt)
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to parse createdAt: %w", err)
	}
	updatedAt, err := time.Parse(SQLITE_TIME_FORMAT, dbWLRequest.UpdatedAt)
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to parse updatedAt: %w", err)
	}

	builder := domainWLRequest.NewBuilder().
		IDFromString(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterIDFromString(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAt(createdAt).
		UpdatedAt(updatedAt)

	if dbWLRequest.ArbiterID != "" {
		builder = builder.ArbiterIDFromString(dbWLRequest.ArbiterID)
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
		ID:            wlRequest.ID().String(),
		RequesterID:   wlRequest.RequesterID().String(),
		Nickname:      wlRequest.Nickname(),
		Status:        wlRequest.Status(),
		DeclineReason: wlRequest.DeclineReason(),
		ArbiterID:     wlRequest.ArbiterID().String(),
		UpdatedAt:     wlRequest.UpdatedAt().Format(SQLITE_TIME_FORMAT),
	})
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to update wl request: %w", err)
	}

	return wlRequest, nil
}
