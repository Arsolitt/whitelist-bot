package handlers

import (
	"context"
	"errors"
	"testing"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/domain/user"
	"whitelist-bot/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testUser := createTestUser(t)

	tests := []struct {
		name          string
		currentState  fsm.State
		setupMock     func(*mockiUserGetter)
		expectedState fsm.State
		expectedError error
		validateMsg   func(*testing.T, *bot.SendMessageParams)
	}{
		{
			name:         "success",
			currentState: fsm.StateIdle,
			setupMock: func(m *mockiUserGetter) {
				m.EXPECT().
					UserByTelegramID(ctx, int64(testUser.TelegramID())).
					Return(testUser, nil).
					Once()
			},
			expectedState: fsm.StateIdle,
			expectedError: nil,
			validateMsg: func(t *testing.T, msg *bot.SendMessageParams) {
				require.NotNil(t, msg)
				assert.Equal(t, int64(123), msg.ChatID)
				assert.Equal(t, models.ParseModeHTML, msg.ParseMode)
				assert.Contains(t, msg.Text, "Информация о пользователе")
				assert.Contains(t, msg.Text, "testuser")
			},
		},
		{
			name:          "invalid_state",
			currentState:  fsm.StateWaitingWLNickname,
			setupMock:     func(m *mockiUserGetter) {},
			expectedState: fsm.StateWaitingWLNickname,
			expectedError: core.ErrInvalidUserState,
			validateMsg: func(t *testing.T, msg *bot.SendMessageParams) {
				assert.Nil(t, msg)
			},
		},
		{
			name:         "repository_error",
			currentState: fsm.StateIdle,
			setupMock: func(m *mockiUserGetter) {
				m.EXPECT().
					UserByTelegramID(ctx, int64(testUser.TelegramID())).
					Return(user.User{}, errors.New("db error")).
					Once()
			},
			expectedState: fsm.StateIdle,
			expectedError: errors.New("failed to get user: db error"),
			validateMsg: func(t *testing.T, msg *bot.SendMessageParams) {
				assert.Nil(t, msg)
			},
		},
		{
			name:         "user_not_found",
			currentState: fsm.StateIdle,
			setupMock: func(m *mockiUserGetter) {
				m.EXPECT().
					UserByTelegramID(ctx, int64(testUser.TelegramID())).
					Return(user.User{}, core.ErrUserNotFound).
					Once()
			},
			expectedState: fsm.StateIdle,
			expectedError: core.ErrUserNotFound,
			validateMsg: func(t *testing.T, msg *bot.SendMessageParams) {
				assert.Nil(t, msg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGetter := newMockiUserGetter(t)
			tt.setupMock(mockGetter)

			handler := Info(mockGetter)

			update := &models.Update{
				Message: &models.Message{
					From: &models.User{
						ID: int64(testUser.TelegramID()),
					},
					Chat: models.Chat{
						ID: 123,
					},
				},
			}

			state, msg, err := handler(ctx, nil, update, tt.currentState)

			assert.Equal(t, tt.expectedState, state)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			tt.validateMsg(t, msg)
		})
	}
}

func createTestUser(t *testing.T) user.User {
	t.Helper()

	u, err := user.NewBuilder().
		IDFromUUID(uuid.New()).
		TelegramIDFromInt(12345).
		ChatIDFromInt(1234567890).
		FirstNameFromString("Test").
		LastNameFromString("User").
		UsernameFromString("testuser").
		CreatedAt(time.Now()).
		UpdatedAt(time.Now()).
		Build()
	require.NoError(t, err)

	return u
}
