package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type (
	TelegramID int64
	FirstName  string
	LastName   string
	Username   string
)

type User struct {
	id         UserID
	telegramID TelegramID
	firstName  FirstName
	lastName   LastName
	username   Username
	createdAt  time.Time
	updatedAt  time.Time
}

func (u User) ID() UserID {
	return u.id
}

func (u User) TelegramID() TelegramID {
	return u.telegramID
}

func (u User) FirstName() FirstName {
	return u.firstName
}

func (u User) LastName() LastName {
	return u.lastName
}

func (u User) Username() Username {
	return u.username
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}

type UserBuilder struct {
	id         UserID
	telegramID TelegramID
	firstName  FirstName
	lastName   LastName
	username   Username
	errors     []error
}

func NewUserBuilder() UserBuilder {
	return UserBuilder{}
}

func (b UserBuilder) ID(id uuid.UUID) UserBuilder {
	if id == uuid.Nil {
		id = uuid.New()
	}
	b.id = UserID(id)
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

func (b UserBuilder) Build() (User, error) {
	if b.id.IsZero() {
		b.id = NewUserID()
	}
	if len(b.errors) > 0 {
		return User{}, errors.Join(b.errors...)
	}

	now := time.Now()
	return User{
		id:         b.id,
		telegramID: b.telegramID,
		firstName:  b.firstName,
		lastName:   b.lastName,
		username:   b.username,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}
