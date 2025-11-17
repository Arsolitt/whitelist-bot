package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	maxFirstNameLength = 64
	maxLastNameLength  = 64
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

func (b Builder) TelegramID(telegramID int64) Builder {
	if telegramID == 0 {
		b.errors = append(b.errors, errors.New("telegram ID required"))
	}
	b.telegramID = TelegramID(telegramID)
	return b
}

func (b Builder) FirstName(firstName string) Builder {
	if len(firstName) > maxFirstNameLength {
		firstNameRunes := []rune(firstName)
		firstName = string(firstNameRunes[:64])
	}
	b.firstName = FirstName(firstName)
	return b
}

func (b Builder) LastName(lastName string) Builder {
	if len(lastName) > maxLastNameLength {
		lastNameRunes := []rune(lastName)
		lastName = string(lastNameRunes[:64])
	}
	b.lastName = LastName(lastName)
	return b
}

func (b Builder) Username(username string) Builder {
	if username == "" {
		b.errors = append(b.errors, errors.New("username required"))
	}
	b.username = Username(username)
	return b
}

func (b Builder) CreatedAt(createdAt time.Time) Builder {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	b.createdAt = createdAt
	return b
}

func (b Builder) UpdatedAt(updatedAt time.Time) Builder {
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	b.updatedAt = updatedAt
	return b
}

func (b Builder) Build() (User, error) {
	if len(b.errors) > 0 {
		return User{}, errors.Join(b.errors...)
	}
	if b.id.IsZero() {
		return User{}, errors.New("ID is required")
	}
	if b.createdAt.IsZero() {
		return User{}, errors.New("createdAt is required")
	}
	if b.updatedAt.IsZero() {
		return User{}, errors.New("updatedAt is required")
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
