package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserBuilder struct {
	id         UserID
	telegramID TelegramID
	firstName  FirstName
	lastName   LastName
	username   Username
	errors     []error
	createdAt  time.Time
	updatedAt  time.Time
}

func NewUserBuilder() UserBuilder {
	return UserBuilder{}
}

func (b UserBuilder) NewID() UserBuilder {
	b.id = NewUserID()
	return b
}

func (b UserBuilder) IDFromString(id string) UserBuilder {
	uuid, err := uuid.Parse(id)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("failed to parse ID: %w", err))
		return b
	}
	b.id = UserID(uuid)
	return b
}

func (b UserBuilder) IDFromUUID(id uuid.UUID) UserBuilder {
	b.id = UserID(id)
	return b
}

func (b UserBuilder) ID(id UserID) UserBuilder {
	b.id = id
	return b
}

func (b UserBuilder) TelegramID(telegramID int64) UserBuilder {
	if telegramID == 0 {
		b.errors = append(b.errors, errors.New("telegram ID required"))
	}
	b.telegramID = TelegramID(telegramID)
	return b
}

func (b UserBuilder) FirstName(firstName string) UserBuilder {
	if len(firstName) > 64 {
		firstNameRunes := []rune(firstName)
		firstName = string(firstNameRunes[:64])
	}
	b.firstName = FirstName(firstName)
	return b
}

func (b UserBuilder) LastName(lastName string) UserBuilder {
	if len(lastName) > 64 {
		lastNameRunes := []rune(lastName)
		lastName = string(lastNameRunes[:64])
	}
	b.lastName = LastName(lastName)
	return b
}

func (b UserBuilder) Username(username string) UserBuilder {
	if username == "" {
		b.errors = append(b.errors, errors.New("username required"))
	}
	b.username = Username(username)
	return b
}

func (b UserBuilder) CreatedAt(createdAt time.Time) UserBuilder {
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	b.createdAt = createdAt
	return b
}

func (b UserBuilder) UpdatedAt(updatedAt time.Time) UserBuilder {
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	b.updatedAt = updatedAt
	return b
}

func (b UserBuilder) Build() (User, error) {
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
