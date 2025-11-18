package wl_request

import (
	"time"

	"whitelist/internal/core/utils"

	"github.com/google/uuid"
)

type Status string

func (s Status) IsZero() bool {
	return s == ""
}

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDeclined Status = "declined"
)

type (
	ID            uuid.UUID
	RequesterID   uuid.UUID
	Nickname      string
	DeclineReason string
	ArbiterID     uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
)

func NewID() ID {
	return ID(utils.NewUniqueID())
}

func (u ID) String() string {
	return utils.UUIDString(u)
}

func (u ID) IsZero() bool {
	return utils.UUIDIsZero(u)
}

func NewRequesterID() RequesterID {
	return RequesterID(utils.NewUniqueID())
}

func (u RequesterID) String() string {
	return utils.UUIDString(u)
}

func (u RequesterID) IsZero() bool {
	return utils.UUIDIsZero(u)
}

func NewArbiterID() ArbiterID {
	return ArbiterID(utils.NewUniqueID())
}

func (u ArbiterID) String() string {
	return utils.UUIDString(u)
}

func (u ArbiterID) IsZero() bool {
	return utils.UUIDIsZero(u)
}

func (u Nickname) IsZero() bool {
	return u == ""
}
