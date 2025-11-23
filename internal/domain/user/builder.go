package user

import (
	"errors"
	"fmt"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/utils"

	"github.com/google/uuid"
)

var (
	ErrIDRequired         = errors.New("ID is required")
	ErrTelegramIDRequired = errors.New("telegram ID required")
	ErrUsernameRequired   = errors.New("username required")
	ErrCreatedAtRequired  = errors.New("createdAt is required")
	ErrUpdatedAtRequired  = errors.New("updatedAt is required")
)

type Builder struct {
	id         ID
	telegramID TelegramID
	firstName  FirstName
	lastName   LastName
	username   Username
	errors     []error
	createdAt  time.Time
	updatedAt  time.Time
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

func (b Builder) TelegramID(telegramID TelegramID) Builder {
	if telegramID.IsZero() {
		b.errors = append(b.errors, ErrTelegramIDRequired)
		return b
	}
	b.telegramID = telegramID
	return b
}

func (b Builder) TelegramIDFromInt(telegramID int64) Builder {
	return b.TelegramID(TelegramID(telegramID))
}

func (b Builder) FirstName(firstName FirstName) Builder {
	if len(firstName) > maxFirstNameLength {
		firstNameRunes := []rune(firstName)
		firstName = FirstName(firstNameRunes[:maxFirstNameLength])
	}
	b.firstName = firstName
	return b
}

func (b Builder) FirstNameFromString(firstName string) Builder {
	return b.FirstName(FirstName(firstName))
}

func (b Builder) LastName(lastName LastName) Builder {
	if len(lastName) > maxLastNameLength {
		lastNameRunes := []rune(lastName)
		lastName = LastName(lastNameRunes[:maxLastNameLength])
	}
	b.lastName = lastName
	return b
}

func (b Builder) LastNameFromString(lastName string) Builder {
	return b.LastName(LastName(lastName))
}

func (b Builder) Username(username Username) Builder {
	if username.IsZero() {
		b.errors = append(b.errors, ErrUsernameRequired)
		return b
	}
	b.username = username
	return b
}

func (b Builder) UsernameFromString(username string) Builder {
	return b.Username(Username(username))
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

func (b Builder) Build() (User, error) {
	if len(b.errors) > 0 {
		return User{}, errors.Join(b.errors...)
	}

	return User{
		id:         b.id,
		telegramID: b.telegramID,
		firstName:  b.firstName,
		lastName:   b.lastName,
		username:   b.username,
		createdAt:  b.createdAt,
		updatedAt:  b.updatedAt,
	}, nil
}
