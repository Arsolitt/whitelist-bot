package handlers

import (
	"context"
	"errors"
	"testing"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlers_ViewPendingWLRequests_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	// Test data
	now := time.Now()
	user, err := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440000").
		TelegramIDFromInt(123456).
		UsernameFromString("testuser").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	userID := user.ID()

	wlRequest, err := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440001").
		RequesterIDFromUserID(userID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	// Setup expectations
	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]domainWLRequest.WLRequest{wlRequest}, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(ctx, userID).
		Return(user, nil).
		Once()

	mockSender.EXPECT().
		SendMessage(ctx, mock.MatchedBy(func(params *bot.SendMessageParams) bool {
			return params.ChatID == int64(789) &&
				params.ParseMode == "HTML" &&
				params.ReplyMarkup != nil
		})).
		Return(&models.Message{ID: 1}, nil).
		Once()

	// Create handler
	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	// Execute
	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, msgParams)
}

func TestHandlers_ViewPendingWLRequests_MultipleRequests(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	// Create multiple users and requests
	now := time.Now()
	user1, _ := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440010").
		TelegramIDFromInt(111).
		UsernameFromString("user1").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user1ID := user1.ID()

	user2, _ := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440011").
		TelegramIDFromInt(222).
		UsernameFromString("user2").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user2ID := user2.ID()

	wlReq1, _ := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440020").
		RequesterIDFromUserID(user1ID).
		NicknameFromString("nick1").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	wlReq2, _ := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440021").
		RequesterIDFromUserID(user2ID).
		NicknameFromString("nick2").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	// Setup expectations
	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]domainWLRequest.WLRequest{wlReq1, wlReq2}, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(ctx, user1ID).
		Return(user1, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(ctx, user2ID).
		Return(user2, nil).
		Once()

	mockSender.EXPECT().
		SendMessage(ctx, mock.AnythingOfType("*bot.SendMessageParams")).
		Return(&models.Message{ID: 1}, nil).
		Twice()

	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, msgParams)
}

func TestHandlers_ViewPendingWLRequests_NoRequests(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	// Empty list of requests
	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]domainWLRequest.WLRequest{}, nil).
		Once()

	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, msgParams)
	assert.Equal(t, int64(789), msgParams.ChatID)
	assert.Equal(t, models.ParseMode("HTML"), msgParams.ParseMode)
	assert.NotEmpty(t, msgParams.Text)
}

func TestHandlers_ViewPendingWLRequests_RepositoryError(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	expectedErr := errors.New("database connection failed")
	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return(nil, expectedErr).
		Once()

	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pending wl requests")
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, msgParams)
}

func TestHandlers_ViewPendingWLRequests_UserNotFoundError(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	now := time.Now()
	user, _ := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440030").
		TelegramIDFromInt(123).
		UsernameFromString("testuser").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	userID := user.ID()

	wlRequest, _ := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440031").
		RequesterIDFromUserID(userID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]domainWLRequest.WLRequest{wlRequest}, nil).
		Once()

	userErr := errors.New("user not found")
	mockUserRepo.EXPECT().
		UserByID(ctx, userID).
		Return(domainUser.User{}, userErr).
		Once()

	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get requester")
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, msgParams)
}

func TestHandlers_ViewPendingWLRequests_SendMessageError_ContinuesProcessing(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)
	mockSender := newMockiMessageSender(t)

	// Create two users and requests
	now := time.Now()
	user1, _ := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440040").
		TelegramIDFromInt(111).
		UsernameFromString("user1").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user1ID := user1.ID()

	user2, _ := domainUser.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440041").
		TelegramIDFromInt(222).
		UsernameFromString("user2").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user2ID := user2.ID()

	wlReq1, _ := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440050").
		RequesterIDFromUserID(user1ID).
		NicknameFromString("nick1").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	wlReq2, _ := domainWLRequest.NewBuilder().
		IDFromString("550e8400-e29b-41d4-a716-446655440051").
		RequesterIDFromUserID(user2ID).
		NicknameFromString("nick2").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	mockWLRepo.EXPECT().
		PendingWLRequests(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]domainWLRequest.WLRequest{wlReq1, wlReq2}, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(ctx, user1ID).
		Return(user1, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(ctx, user2ID).
		Return(user2, nil).
		Once()

	// First message fails
	mockSender.EXPECT().
		SendMessage(ctx, mock.AnythingOfType("*bot.SendMessageParams")).
		Return(nil, errors.New("telegram API error")).
		Once()

	// Second message succeeds
	mockSender.EXPECT().
		SendMessage(ctx, mock.AnythingOfType("*bot.SendMessageParams")).
		Return(&models.Message{ID: 2}, nil).
		Once()

	h := NewWithSender(mockUserRepo, mockWLRepo, mockSender, core.Config{})

	state, msgParams, err := h.ViewPendingWLRequests(ctx, nil, update, fsm.StateIdle)

	// Should not return error, continues processing
	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, msgParams)
}
