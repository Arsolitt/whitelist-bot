package handlers

import (
	"context"
	"errors"
	"testing"
	"time"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/router"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
	wlRequestRepo "whitelist-bot/internal/repository/wl_request"

	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewPendingWLRequests_Success(t *testing.T) {
	ctx := context.Background()

	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()
	user, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("testuser").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	userID := user.ID()

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
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

	mockWLRepo.EXPECT().
		PendingWLRequestsWithRequester(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]wlRequestRepo.PendingWLRequestWithRequester{
			{WlRequest: wlRequest, User: user},
		}, nil).
		Once()

	handler := ViewPendingWLRequests(mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	msgResponse, ok := response.(*router.MessageResponse)
	require.True(t, ok)
	assert.Len(t, msgResponse.Params, 1)
	assert.NotNil(t, msgResponse.Params[0].ReplyMarkup)
}

func TestViewPendingWLRequests_MultipleRequests(t *testing.T) {
	ctx := context.Background()

	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()
	user1, _ := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(111).
		UsernameFromString("user1").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user1ID := user1.ID()

	user2, _ := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(222).
		UsernameFromString("user2").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	user2ID := user2.ID()

	wlReq1, _ := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(user1ID).
		NicknameFromString("nick1").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()

	wlReq2, _ := domainWLRequest.NewBuilder().
		NewID().
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
		PendingWLRequestsWithRequester(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]wlRequestRepo.PendingWLRequestWithRequester{
			{WlRequest: wlReq1, User: user1},
			{WlRequest: wlReq2, User: user2},
		}, nil).
		Once()

	handler := ViewPendingWLRequests(mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	msgResponse, ok := response.(*router.MessageResponse)
	require.True(t, ok)
	assert.Len(t, msgResponse.Params, 2)
}

func TestViewPendingWLRequests_NoRequests(t *testing.T) {
	ctx := context.Background()

	mockWLRepo := newMockiWLRequestRepository(t)

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	mockWLRepo.EXPECT().
		PendingWLRequestsWithRequester(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return([]wlRequestRepo.PendingWLRequestWithRequester{}, nil).
		Once()

	handler := ViewPendingWLRequests(mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	msgResponse, ok := response.(*router.MessageResponse)
	require.True(t, ok)
	assert.Len(t, msgResponse.Params, 1)
	assert.NotEmpty(t, msgResponse.Params[0].Text)
}

func TestViewPendingWLRequests_RepositoryError(t *testing.T) {
	ctx := context.Background()

	mockWLRepo := newMockiWLRequestRepository(t)

	update := &models.Update{
		Message: &models.Message{
			Chat: models.Chat{ID: 789},
		},
	}

	expectedErr := errors.New("database connection failed")
	mockWLRepo.EXPECT().
		PendingWLRequestsWithRequester(ctx, int64(PENDING_WL_REQUESTS_LIMIT)).
		Return(nil, expectedErr).
		Once()

	handler := ViewPendingWLRequests(mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pending wl requests")
	assert.Equal(t, fsm.StateIdle, state)
	assert.Nil(t, response)
}
