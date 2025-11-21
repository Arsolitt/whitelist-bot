package wl_request

import (
	"fmt"
	"time"
)

type WLRequest struct {
	id            ID
	requesterID   RequesterID
	nickname      Nickname
	status        Status
	declineReason DeclineReason
	arbiterID     ArbiterID
	createdAt     time.Time
	updatedAt     time.Time
}

func (w WLRequest) ID() ID {
	return w.id
}

func (w WLRequest) RequesterID() RequesterID {
	return w.requesterID
}

func (w WLRequest) Nickname() Nickname {
	return w.nickname
}

func (w WLRequest) Status() Status {
	return w.status
}

func (w WLRequest) DeclineReason() DeclineReason {
	return w.declineReason
}

func (w WLRequest) ArbiterID() ArbiterID {
	return w.arbiterID
}

func (w WLRequest) CreatedAt() time.Time {
	return w.createdAt
}

func (w WLRequest) UpdatedAt() time.Time {
	return w.updatedAt
}

func (w WLRequest) UpdateTimestamp() WLRequest {
	w.updatedAt = time.Now()
	return w
}

func (w WLRequest) Approve(arbiterID ArbiterID) (WLRequest, error) {
	newWLRequest, err := NewBuilder().
		ID(w.ID()).
		RequesterID(w.RequesterID()).
		Nickname(w.Nickname()).
		Status(StatusApproved).
		DeclineReason(w.DeclineReason()).
		ArbiterID(arbiterID).
		CreatedAt(w.CreatedAt()).
		UpdatedAt(w.UpdatedAt()).
		Build()
	if err != nil {
		return WLRequest{}, fmt.Errorf("failed to approve wl request: %w", err)
	}
	return newWLRequest, nil
}

func (w WLRequest) Decline(arbiterID ArbiterID, declineReason DeclineReason) (WLRequest, error) {
	newWLRequest, err := NewBuilder().
		ID(w.ID()).
		RequesterID(w.RequesterID()).
		Nickname(w.Nickname()).
		Status(StatusDeclined).
		DeclineReason(declineReason).
		ArbiterID(arbiterID).
		CreatedAt(w.CreatedAt()).
		UpdatedAt(w.UpdatedAt()).
		Build()
	if err != nil {
		return WLRequest{}, fmt.Errorf("failed to decline wl request: %w", err)
	}
	return newWLRequest, nil
}
