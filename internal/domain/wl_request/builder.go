package wl_request

import (
	"errors"
	"fmt"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/utils"
	domainUser "whitelist-bot/internal/domain/user"

	"github.com/google/uuid"
)

var (
	ErrIDRequired                         = errors.New("ID required")
	ErrRequesterIDRequired                = errors.New("requester ID required")
	ErrNicknameRequired                   = errors.New("nickname required")
	ErrStatusRequired                     = errors.New("status required")
	ErrInvalidStatus                      = errors.New("invalid status")
	ErrCreatedAtRequired                  = errors.New("createdAt required")
	ErrUpdatedAtRequired                  = errors.New("updatedAt required")
	ErrArbiterRequiredForNonPending       = errors.New("arbiter ID is required for non-pending status")
	ErrDeclineReasonRequiredForDeclined   = errors.New("decline reason is required for declined status")
	ErrDeclineReasonNotAllowedForApproved = errors.New("decline reason is not allowed for approved status")
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
	return b.ID(ID(utils.NewUniqueID()))
}

func (b Builder) IDFromString(id string) Builder {
	idUUID, err := utils.UUIDFromString[ID](id)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("%w: %w", core.ErrFailedToParseID, err))
		return b
	}
	return b.ID(ID(idUUID))
}

func (b Builder) IDFromUUID(id uuid.UUID) Builder {
	return b.ID(ID(id))
}

func (b Builder) ID(id ID) Builder {
	if id.IsZero() {
		b.errors = append(b.errors, ErrIDRequired)
		return b
	}
	b.id = id
	return b
}

func (b Builder) RequesterID(requesterID RequesterID) Builder {
	if requesterID.IsZero() {
		b.errors = append(b.errors, ErrRequesterIDRequired)
		return b
	}
	b.requesterID = requesterID
	return b
}

func (b Builder) RequesterIDFromUserID(userID domainUser.ID) Builder {
	return b.RequesterID(RequesterID(userID))
}

func (b Builder) RequesterIDFromString(requesterID string) Builder {
	requesterUUID, err := utils.UUIDFromString[ID](requesterID)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("%w: %w", core.ErrFailedToParseID, err))
		return b
	}
	return b.RequesterID(RequesterID(requesterUUID))
}

func (b Builder) Nickname(nickname Nickname) Builder {
	if nickname.IsZero() {
		b.errors = append(b.errors, ErrNicknameRequired)
		return b
	}
	if len(nickname) > maxNicknameLength {
		b.errors = append(b.errors, ErrInvalidNicknameLength(nickname))
		return b
	}
	b.nickname = nickname
	return b
}

func (b Builder) NicknameFromString(nickname string) Builder {
	return b.Nickname(Nickname(nickname))
}

func (b Builder) ArbiterID(arbiterID ArbiterID) Builder {
	b.arbiterID = arbiterID
	return b
}

func (b Builder) ArbiterIDFromString(arbiterID string) Builder {
	arbiterUUID, err := utils.UUIDFromString[ID](arbiterID)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("%w: %w", core.ErrFailedToParseID, err))
		return b
	}
	return b.ArbiterID(ArbiterID(arbiterUUID))
}

func (b Builder) ArbiterIDFromUserID(userID domainUser.ID) Builder {
	return b.ArbiterID(ArbiterID(userID))
}

func (b Builder) DeclineReason(declineReason DeclineReason) Builder {
	if len(declineReason) > maxDeclineReasonLength {
		b.errors = append(b.errors, ErrInvalidDeclineReasonLength(len(declineReason)))
		return b
	}
	b.declineReason = declineReason
	return b
}

func (b Builder) DeclineReasonFromString(declineReason string) Builder {
	return b.DeclineReason(DeclineReason(declineReason))
}

func (b Builder) Status(status Status) Builder {
	if status.IsZero() {
		b.errors = append(b.errors, ErrStatusRequired)
		return b
	}
	if status != StatusPending &&
		status != StatusApproved &&
		status != StatusDeclined {
		b.errors = append(b.errors, fmt.Errorf("%w: %s", ErrInvalidStatus, status))
		return b
	}
	b.status = status
	return b
}

func (b Builder) StatusFromString(status string) Builder {
	return b.Status(Status(status))
}

func (b Builder) CreatedAt(createdAt time.Time) Builder {
	if createdAt.IsZero() {
		b.errors = append(b.errors, ErrCreatedAtRequired)
		return b
	}
	b.createdAt = createdAt
	return b
}

func (b Builder) UpdatedAt(updatedAt time.Time) Builder {
	if updatedAt.IsZero() {
		b.errors = append(b.errors, ErrUpdatedAtRequired)
		return b
	}
	b.updatedAt = updatedAt
	return b
}

func (b Builder) Build() (WLRequest, error) {
	if len(b.errors) > 0 {
		return WLRequest{}, errors.Join(b.errors...)
	}
	if b.id.IsZero() {
		b.errors = append(b.errors, ErrIDRequired)
	}
	if b.requesterID.IsZero() {
		b.errors = append(b.errors, ErrRequesterIDRequired)
	}
	if b.nickname.IsZero() {
		b.errors = append(b.errors, ErrNicknameRequired)
	}
	if b.status.IsZero() {
		b.errors = append(b.errors, ErrStatusRequired)
	}
	if b.createdAt.IsZero() {
		b.errors = append(b.errors, ErrCreatedAtRequired)
	}
	if b.updatedAt.IsZero() {
		b.errors = append(b.errors, ErrUpdatedAtRequired)
	}
	if len(b.errors) > 0 {
		return WLRequest{}, errors.Join(b.errors...)
	}
	if b.arbiterID.IsZero() && b.status != StatusPending {
		return WLRequest{}, ErrArbiterRequiredForNonPending
	}
	if b.status == StatusDeclined && b.declineReason.IsZero() {
		return WLRequest{}, ErrDeclineReasonRequiredForDeclined
	}
	if b.status == StatusApproved && !b.declineReason.IsZero() {
		return WLRequest{}, ErrDeclineReasonNotAllowedForApproved
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
