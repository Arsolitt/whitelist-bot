package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	domainWLRequest "whitelist/internal/domain/wl_request"
)

type IQueryable interface {
	Begin() (*sql.Tx, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type WLRequestRepository struct {
	db IQueryable
}

func NewWLRequestRepository(db IQueryable) *WLRequestRepository {
	return &WLRequestRepository{db: db}
}

func (r *WLRequestRepository) CreateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	now := time.Now()
	nowFormatted := now.Format("2006-01-02T15:04:05-0700")

	dbWLRequest, err := q.CreateWLRequest(ctx, CreateWLRequestParams{
		ID:            wlRequest.ID().String(),
		RequesterID:   wlRequest.RequesterID().String(),
		Nickname:      wlRequest.Nickname(),
		Status:        domainWLRequest.StatusPending,
		DeclineReason: wlRequest.DeclineReason(),
		CreatedAt:     nowFormatted,
		UpdatedAt:     nowFormatted,
	})
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to create wl request: %w", err)
	}
	newWLRequest, err := domainWLRequest.NewBuilder().
		IDFromString(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterIDFromString(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}
	return newWLRequest, nil
}

func (r *WLRequestRepository) PendingWLRequests(ctx context.Context) ([]domainWLRequest.WLRequest, error) {
	q := New(r.db)

	dbWLRequests, err := q.PendingWLRequests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending wl requests: %w", err)
	}
	newWLRequests := make([]domainWLRequest.WLRequest, len(dbWLRequests))
	for i, dbWLRequest := range dbWLRequests {
		builder := domainWLRequest.NewBuilder().
			IDFromString(dbWLRequest.ID).
			Status(dbWLRequest.Status).
			DeclineReason(dbWLRequest.DeclineReason).
			RequesterIDFromString(dbWLRequest.RequesterID).
			Nickname(dbWLRequest.Nickname).
			CreatedAtFromString(dbWLRequest.CreatedAt).
			UpdatedAtFromString(dbWLRequest.UpdatedAt)

		if dbWLRequest.ArbiterID != nil {
			builder = builder.ArbiterIDFromString(*dbWLRequest.ArbiterID)
		}

		newWLRequests[i], err = builder.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build wl request: %w", err)
		}
	}
	return newWLRequests, nil
}

func (r *WLRequestRepository) PendingWLRequest(ctx context.Context) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	dbWLRequest, err := q.PendingWLRequest(ctx)
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to get  pending wl request: %w", err)
	}

	builder := domainWLRequest.NewBuilder().
		IDFromString(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterIDFromString(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAtFromString(dbWLRequest.CreatedAt).
		UpdatedAtFromString(dbWLRequest.UpdatedAt)

	if dbWLRequest.ArbiterID != nil {
		builder = builder.ArbiterIDFromString(*dbWLRequest.ArbiterID)
	}

	wlRequest, err := builder.Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}

	return wlRequest, nil
}

func (r *WLRequestRepository) WLRequestByID(ctx context.Context, id domainWLRequest.ID) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	dbWLRequest, err := q.WLRequestByID(ctx, id.String())
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to get wl request by id: %w", err)
	}

	builder := domainWLRequest.NewBuilder().
		IDFromString(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterIDFromString(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAtFromString(dbWLRequest.CreatedAt).
		UpdatedAtFromString(dbWLRequest.UpdatedAt)

	if dbWLRequest.ArbiterID != nil {
		builder = builder.ArbiterIDFromString(*dbWLRequest.ArbiterID)
	}

	wlRequest, err := builder.Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}

	return wlRequest, nil
}

func (r *WLRequestRepository) UpdateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error) {
	q := New(r.db)

	now := time.Now()
	nowFormatted := now.Format("2006-01-02T15:04:05-0700")

	var arbiterID *string
	if !wlRequest.ArbiterID().IsZero() {
		arbiterIDStr := wlRequest.ArbiterID().String()
		arbiterID = &arbiterIDStr
	}

	dbWLRequest, err := q.UpdateWLRequest(ctx, UpdateWLRequestParams{
		ID:            wlRequest.ID().String(),
		RequesterID:   wlRequest.RequesterID().String(),
		Nickname:      wlRequest.Nickname(),
		Status:        wlRequest.Status(),
		DeclineReason: wlRequest.DeclineReason(),
		ArbiterID:     arbiterID,
		UpdatedAt:     nowFormatted,
	})
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to update wl request: %w", err)
	}

	builder := domainWLRequest.NewBuilder().
		IDFromString(dbWLRequest.ID).
		Status(dbWLRequest.Status).
		DeclineReason(dbWLRequest.DeclineReason).
		RequesterIDFromString(dbWLRequest.RequesterID).
		Nickname(dbWLRequest.Nickname).
		CreatedAtFromString(dbWLRequest.CreatedAt).
		UpdatedAtFromString(dbWLRequest.UpdatedAt)

	if dbWLRequest.ArbiterID != nil {
		builder = builder.ArbiterIDFromString(*dbWLRequest.ArbiterID)
	}

	updatedWLRequest, err := builder.Build()
	if err != nil {
		return domainWLRequest.WLRequest{}, fmt.Errorf("failed to build wl request: %w", err)
	}

	return updatedWLRequest, nil
}
