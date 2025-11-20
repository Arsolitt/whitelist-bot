package wl_request

import (
	"fmt"
	"time"

	"whitelist/internal/core"
	"whitelist/internal/core/utils"

	"github.com/google/uuid"
)

const (
	maxNicknameLength      = 20
	maxDeclineReasonLength = 255
)

var (
	ErrInvalidNicknameLength = func(nickaname Nickname) error {
		return fmt.Errorf("%w: nickname: %s is too long: %d", core.ErrInvalidLength, nickaname, len(nickaname))
	}
	ErrInvalidDeclineReasonLength = func(length int) error {
		return fmt.Errorf("%w: decline reason is too long: %d", core.ErrInvalidLength, length)
	}
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

func (u DeclineReason) IsZero() bool {
	return u == ""
}
