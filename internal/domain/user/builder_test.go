package user

import (
	"errors"
	"testing"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/utils"

	"github.com/google/uuid"
)

func TestBuilder_Build_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()

	user, err := NewBuilder().
		ID(id).
		TelegramID(telegramID).
		FirstName(firstName).
		LastName(lastName).
		Username(username).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}

	if user.ID() != id {
		t.Errorf("expected user ID to be %s, got %s", id, user.ID())
	}
	if user.TelegramID() != telegramID {
		t.Errorf("expected user Telegram ID to be %d, got %d", telegramID, user.TelegramID())
	}
	if user.FirstName() != firstName {
		t.Errorf("expected user first name to be %s, got %s", firstName, user.FirstName())
	}
	if user.LastName() != lastName {
		t.Errorf("expected user last name to be %s, got %s", lastName, user.LastName())
	}
	if user.Username() != username {
		t.Errorf("expected user username to be %s, got %s", username, user.Username())
	}
	if user.CreatedAt() != now {
		t.Errorf("expected user created at to be %s, got %s", now, user.CreatedAt())
	}
	if user.UpdatedAt() != now {
		t.Errorf("expected user updated at to be %s, got %s", now, user.UpdatedAt())
	}
}

func TestBuilder_Build_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()
	tests := []struct {
		name          string
		builder       func() Builder
		expectedError error
	}{
		{
			name: "ID is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(ID(uuid.Nil)).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrIDRequired,
		},
		{
			name: "ID is invalid",
			builder: func() Builder {
				return NewBuilder().
					IDFromString("invalid").
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrFailedToParseID,
		},/*  */
		{
			name: "Telegram ID negative",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(TelegramID(-1)).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Telegram ID zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(TelegramID(0)).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Username required",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(Username("")).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrUsernameRequired,
		},
		{
			name: "CreatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(time.Time{}).
					UpdatedAt(now)
			},
			expectedError: ErrCreatedAtRequired,
		},
		{
			name: "CreatedAt not set",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					UpdatedAt(now)
			},
			expectedError: ErrCreatedAtRequired,
		},
		{
			name: "UpdatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(time.Time{})
			},
			expectedError: ErrUpdatedAtRequired,
		},
		{
			name: "UpdatedAt not set",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now)
			},
			expectedError: ErrUpdatedAtRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder().Build()
			if err != nil && !errors.Is(err, test.expectedError) {
				t.Errorf("expected error %v, got %v", test.expectedError, err)
			}
		})
	}
}
