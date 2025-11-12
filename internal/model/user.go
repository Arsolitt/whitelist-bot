package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

type User struct {
	ID         UserID
	TelegramID int64
	CustomName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewUser(telegramID int64, customName string) (User, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return User{}, fmt.Errorf("failed to generate new UUID: %w", err)
	}

	return User{
		ID:         UserID(newUUID),
		TelegramID: telegramID,
		CustomName: customName,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}
