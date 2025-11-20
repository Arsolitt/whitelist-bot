package user

import (
	"time"
)

type User struct {
	id         ID
	telegramID TelegramID
	firstName  FirstName
	lastName   LastName
	username   Username
	createdAt  time.Time
	updatedAt  time.Time
}

func (u User) ID() ID {
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

func (u User) UpdateTimestamp() User {
	u.updatedAt = time.Now()
	return u
}
