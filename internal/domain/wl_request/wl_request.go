package wl_request

import (
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

func (w *WLRequest) ID() ID {
	return w.id
}

func (w *WLRequest) RequesterID() RequesterID {
	return w.requesterID
}

func (w *WLRequest) Nickname() Nickname {
	return w.nickname
}

func (w *WLRequest) Status() Status {
	return w.status
}

func (w *WLRequest) DeclineReason() DeclineReason {
	return w.declineReason
}

func (w *WLRequest) ArbiterID() ArbiterID {
	return w.arbiterID
}

func (w *WLRequest) CreatedAt() time.Time {
	return w.createdAt
}

func (w *WLRequest) UpdatedAt() time.Time {
	return w.updatedAt
}
