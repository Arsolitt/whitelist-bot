package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"whitelist-bot/internal/callbacks"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/router"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeclineWLRequest_Success(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()

	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	requesterID := requester.ID()

	arbiter, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(789012).
		ChatIDFromInt(1234567890).
		UsernameFromString("arbiter").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requesterID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: int64(arbiter.TelegramID())},
			Message: models.MaybeInaccessibleMessage{
				Message: &models.Message{
					ID:   100,
					Chat: models.Chat{ID: 789},
				},
			},
		},
	}

	declinedRequest, err := wlRequest.Decline(
		domainWLRequest.ArbiterID(arbiter.ID()),
		domainWLRequest.DeclineReason("Отклонено администратором"),
	)
	require.NoError(t, err)

	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(wlRequest, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByTelegramID(mock.Anything, int64(arbiter.TelegramID())).
		Return(arbiter, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(mock.Anything, requesterID).
		Return(requester, nil).
		Once()

	mockWLRepo.EXPECT().
		UpdateWLRequest(mock.Anything, mock.MatchedBy(func(req domainWLRequest.WLRequest) bool {
			return req.Status() == domainWLRequest.StatusDeclined &&
				req.ArbiterID() == domainWLRequest.ArbiterID(arbiter.ID())
		})).
		Return(declinedRequest, nil).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.NoError(t, err)
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
	assert.Contains(t, callbackResponse.CallbackParams.Text, "Заявка отклонена")
	assert.NotNil(t, callbackResponse.EditParams)
}

func TestDeclineWLRequest_InvalidCallbackData(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: "invalid json data",
			From: models.User{ID: 789012},
		},
	}

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal callback data")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_InvalidAction(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()
	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requester.ID()).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	callbackData := callbacks.NewWLRequestCallbackData(wlRequest.ID(), core.ActionWLRequestApprove)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: 789012},
		},
	}

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_WLRequestNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()
	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requester.ID()).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: 789012},
		},
	}

	expectedErr := errors.New("wl request not found")
	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(domainWLRequest.WLRequest{}, expectedErr).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get wl request")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_ArbiterNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()
	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requester.ID()).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: 789012},
		},
	}

	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(wlRequest, nil).
		Once()

	expectedErr := errors.New("arbiter not found")
	mockUserRepo.EXPECT().
		UserByTelegramID(mock.Anything, int64(789012)).
		Return(domainUser.User{}, expectedErr).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get arbiter")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_RequesterNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()

	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	requesterID := requester.ID()

	arbiter, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(789012).
		ChatIDFromInt(1234567890).
		UsernameFromString("arbiter").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requesterID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: int64(arbiter.TelegramID())},
		},
	}

	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(wlRequest, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByTelegramID(mock.Anything, int64(arbiter.TelegramID())).
		Return(arbiter, nil).
		Once()

	expectedErr := errors.New("requester not found")
	mockUserRepo.EXPECT().
		UserByID(mock.Anything, requesterID).
		Return(domainUser.User{}, expectedErr).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get requester")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_DeclineError(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()

	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	requesterID := requester.ID()

	arbiter, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(789012).
		ChatIDFromInt(1234567890).
		UsernameFromString("arbiter").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requesterID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusApproved)).
		ArbiterIDFromUserID(arbiter.ID()).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: int64(arbiter.TelegramID())},
		},
	}

	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(wlRequest, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByTelegramID(mock.Anything, int64(arbiter.TelegramID())).
		Return(arbiter, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(mock.Anything, requesterID).
		Return(requester, nil).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decline wl request")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}

func TestDeclineWLRequest_UpdateWLRequestError(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := newMockiUserRepository(t)
	mockWLRepo := newMockiWLRequestRepository(t)

	now := time.Now()

	requester, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(123456).
		ChatIDFromInt(1234567890).
		UsernameFromString("requester").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	requesterID := requester.ID()

	arbiter, err := domainUser.NewBuilder().
		NewID().
		TelegramIDFromInt(789012).
		ChatIDFromInt(1234567890).
		UsernameFromString("arbiter").
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(requesterID).
		NicknameFromString("testnick").
		StatusFromString(string(domainWLRequest.StatusPending)).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	require.NoError(t, err)
	wlRequestID := wlRequest.ID()

	callbackData := callbacks.NewWLRequestCallbackData(wlRequestID, core.ActionWLRequestDecline)
	callbackDataJSON, err := json.Marshal(callbackData)
	require.NoError(t, err)

	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback123",
			Data: string(callbackDataJSON),
			From: models.User{ID: int64(arbiter.TelegramID())},
		},
	}

	mockWLRepo.EXPECT().
		WLRequestByID(mock.Anything, wlRequestID).
		Return(wlRequest, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByTelegramID(mock.Anything, int64(arbiter.TelegramID())).
		Return(arbiter, nil).
		Once()

	mockUserRepo.EXPECT().
		UserByID(mock.Anything, requesterID).
		Return(requester, nil).
		Once()

	expectedErr := errors.New("database error")
	mockWLRepo.EXPECT().
		UpdateWLRequest(mock.Anything, mock.AnythingOfType("wl_request.WLRequest")).
		Return(domainWLRequest.WLRequest{}, expectedErr).
		Once()

	handler := DeclineWLRequest(mockUserRepo, mockWLRepo)
	state, response, err := handler(ctx, nil, update, fsm.StateIdle)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update wl request")
	assert.Equal(t, fsm.StateIdle, state)
	require.NotNil(t, response)

	callbackResponse, ok := response.(*router.CallbackResponse)
	require.True(t, ok)
	assert.NotNil(t, callbackResponse.CallbackParams)
}
