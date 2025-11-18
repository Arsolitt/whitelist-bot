package wl_request

import (
	"errors"
	"fmt"
	"time"
	"whitelist/internal/core/utils"
	domainUser "whitelist/internal/domain/user"

	"github.com/google/uuid"
)

type Builder struct {
	id            ID
	requesterID   RequesterID
	nickname      Nickname
	status        Status
	declineReason DeclineReason
	arbiterID     ArbiterID
	errors        []error
	createdAt     time.Time
	updatedAt     time.Time
}

func NewBuilder() Builder {
	return Builder{}
}

func (b Builder) NewID() Builder {
	b.id = NewID()
	return b
}

func (b Builder) IDFromString(id string) Builder {
	uuid, err := uuid.Parse(id)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse ID: %w", err))
		return b
	}
	b.id = ID(uuid)
	return b
}

func (b Builder) IDFromUUID(id uuid.UUID) Builder {
	b.id = ID(id)
	return b
}

func (b Builder) ID(id ID) Builder {
	b.id = id
	return b
}

func (b Builder) RequesterID(requesterID RequesterID) Builder {
	b.requesterID = requesterID
	return b
}

func (b Builder) RequesterIDFromUserID(userID domainUser.ID) Builder {
	b.requesterID = RequesterID(userID)
	return b
}

func (b Builder) RequesterIDFromString(requesterID string) Builder {
	requesterUUID, err := utils.UUIDFromString[ID](requesterID)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse ID: %w", err))
		return b
	}
	b.requesterID = RequesterID(requesterUUID)
	return b
}

func (b Builder) Nickname(nickname Nickname) Builder {
	b.nickname = nickname
	return b
}

func (b Builder) NicknameFromString(nickname string) Builder {
	b.nickname = Nickname(nickname)
	return b
}

func (b Builder) ArbiterID(arbiterID ArbiterID) Builder {
	b.arbiterID = arbiterID
	return b
}

func (b Builder) DeclineReason(declineReason DeclineReason) Builder {
	b.declineReason = declineReason
	return b
}

func (b Builder) DeclineReasonFromString(declineReason string) Builder {
	b.declineReason = DeclineReason(declineReason)
	return b
}

func (b Builder) Status(status Status) Builder {
	b.status = status
	return b
}

func (b Builder) StatusFromString(status string) Builder {
	b.status = Status(status)
	return b
}

func (b Builder) CreatedAt(createdAt time.Time) Builder {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	b.createdAt = createdAt
	return b
}

func (b Builder) CreatedAtFromString(createdAt string) Builder {
	createdAtTime, err := time.Parse("2006-01-02T15:04:05-0700", createdAt)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse createdAt: %w", err))
		return b
	}
	b.createdAt = createdAtTime
	return b
}

func (b Builder) UpdatedAt(updatedAt time.Time) Builder {
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	b.updatedAt = updatedAt
	return b
}

func (b Builder) UpdatedAtFromString(updatedAt string) Builder {
	updatedAtTime, err := time.Parse("2006-01-02T15:04:05-0700", updatedAt)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse updatedAt: %w", err))
		return b
	}
	b.updatedAt = updatedAtTime
	return b
}

func (b Builder) Build() (WLRequest, error) {
	if len(b.errors) > 0 {
		return WLRequest{}, errors.Join(b.errors...)
	}
	if b.id.IsZero() {
		return WLRequest{}, errors.New("ID is required")
	}
	if b.requesterID.IsZero() {
		return WLRequest{}, errors.New("requester ID is required")
	}
	if b.nickname.IsZero() {
		return WLRequest{}, errors.New("nickname is required")
	}

	return WLRequest{
		id:            b.id,
		requesterID:   b.requesterID,
		nickname:      b.nickname,
		status:        b.status,
		declineReason: b.declineReason,
		arbiterID:     b.arbiterID,
		createdAt:     b.createdAt,
		updatedAt:     b.updatedAt,
	}, nil
}
