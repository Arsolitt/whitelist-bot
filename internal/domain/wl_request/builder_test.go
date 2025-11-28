package wl_request

import (
	"errors"
	"strings"
	"testing"
	"time"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	requesterID := RequesterID(utils.NewUniqueID())
	arbiterID := ArbiterID(utils.NewUniqueID())
	nickname := Nickname("PlayerNick")
	status := StatusApproved
	declineReason := DeclineReason("")
	now := time.Now()

	wlRequest, err := NewBuilder().
		ID(id).
		RequesterID(requesterID).
		Nickname(nickname).
		Status(status).
		ArbiterID(arbiterID).
		DeclineReason(declineReason).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		assert.Fail(t, "failed to build wl_request: %v", err)
	}

	if wlRequest.ID() != id {
		assert.Fail(t, "expected wl_request ID to be %s, got %s", id, wlRequest.ID())
	}
	if wlRequest.RequesterID() != requesterID {
		assert.Fail(t, "expected wl_request RequesterID to be %s, got %s", requesterID, wlRequest.RequesterID())
	}
	if wlRequest.Nickname() != nickname {
		assert.Fail(t, "expected wl_request Nickname to be %s, got %s", nickname, wlRequest.Nickname())
	}
	if wlRequest.Status() != status {
		assert.Fail(t, "expected wl_request Status to be %s, got %s", status, wlRequest.Status())
	}
	if wlRequest.ArbiterID() != arbiterID {
		assert.Fail(t, "expected wl_request ArbiterID to be %s, got %s", arbiterID, wlRequest.ArbiterID())
	}
	if wlRequest.DeclineReason() != declineReason {
		assert.Fail(t, "expected wl_request DeclineReason to be %s, got %s", declineReason, wlRequest.DeclineReason())
	}
	if wlRequest.CreatedAt() != now {
		assert.Fail(t, "expected wl_request CreatedAt to be %s, got %s", now, wlRequest.CreatedAt())
	}
	if wlRequest.UpdatedAt() != now {
		assert.Fail(t, "expected wl_request UpdatedAt to be %s, got %s", now, wlRequest.UpdatedAt())
	}
}

func TestBuilder_Build_Success_Pending(t *testing.T) {
	id := ID(utils.NewUniqueID())
	requesterID := RequesterID(utils.NewUniqueID())
	nickname := Nickname("PlayerNick")
	status := StatusPending
	now := time.Now()

	wlRequest, err := NewBuilder().
		ID(id).
		RequesterID(requesterID).
		Nickname(nickname).
		Status(status).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		assert.Fail(t, "failed to build wl_request: %v", err)
	}

	if wlRequest.Status() != status {
		assert.Fail(t, "expected wl_request Status to be %s, got %s", status, wlRequest.Status())
	}
	if !wlRequest.IsPending() {
		assert.Fail(t, "expected wl_request to be pending")
	}
}

func TestBuilder_Build_Success_Declined(t *testing.T) {
	id := ID(utils.NewUniqueID())
	requesterID := RequesterID(utils.NewUniqueID())
	arbiterID := ArbiterID(utils.NewUniqueID())
	nickname := Nickname("PlayerNick")
	status := StatusDeclined
	declineReason := DeclineReason("Too young")
	now := time.Now()

	wlRequest, err := NewBuilder().
		ID(id).
		RequesterID(requesterID).
		Nickname(nickname).
		Status(status).
		ArbiterID(arbiterID).
		DeclineReason(declineReason).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
	if err != nil {
		assert.Fail(t, "failed to build wl_request: %v", err)
	}

	if wlRequest.Status() != status {
		assert.Fail(t, "expected wl_request Status to be %s, got %s", status, wlRequest.Status())
	}
	if wlRequest.DeclineReason() != declineReason {
		assert.Fail(t, "expected wl_request DeclineReason to be %s, got %s", declineReason, wlRequest.DeclineReason())
	}
}

func TestBuilder_Build_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	requesterID := RequesterID(utils.NewUniqueID())
	arbiterID := ArbiterID(utils.NewUniqueID())
	nickname := Nickname("PlayerNick")
	status := StatusApproved
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
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
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
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "RequesterID is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(RequesterID(uuid.Nil)).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrRequesterIDRequired,
		},
		{
			name: "RequesterID is invalid",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterIDFromString("invalid").
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "Nickname is empty",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(Nickname("")).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrNicknameRequired,
		},
		{
			name: "Nickname is too long",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(Nickname(strings.Repeat("a", maxNicknameLength+1))).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrInvalidLength,
		},
		{
			name: "Status is empty",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(Status("")).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrStatusRequired,
		},
		{
			name: "Status is invalid",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(Status("invalid")).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrInvalidStatus,
		},
		{
			name: "CreatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
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
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
					UpdatedAt(now)
			},
			expectedError: ErrCreatedAtRequired,
		},
		{
			name: "UpdatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
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
					RequesterID(requesterID).
					Nickname(nickname).
					Status(status).
					ArbiterID(arbiterID).
					CreatedAt(now)
			},
			expectedError: ErrUpdatedAtRequired,
		},
		{
			name: "ArbiterID required for approved status",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusApproved).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrArbiterRequiredForNonPending,
		},
		{
			name: "ArbiterID required for declined status",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusDeclined).
					DeclineReason(DeclineReason("Too young")).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrArbiterRequiredForNonPending,
		},
		{
			name: "DeclineReason required for declined status",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusDeclined).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrDeclineReasonRequiredForDeclined,
		},
		{
			name: "DeclineReason not allowed for approved status",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusApproved).
					DeclineReason(DeclineReason("Too young")).
					ArbiterID(arbiterID).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrDeclineReasonNotAllowedForApproved,
		},
		{
			name: "DeclineReason is too long",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusDeclined).
					ArbiterID(arbiterID).
					DeclineReason(DeclineReason(strings.Repeat("a", maxDeclineReasonLength+1))).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrInvalidLength,
		},
		{
			name: "ArbiterID is invalid",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					RequesterID(requesterID).
					Nickname(nickname).
					Status(StatusApproved).
					ArbiterIDFromString("invalid").
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "Empty builder",
			builder: func() Builder {
				return NewBuilder()
			},
			expectedError: errors.Join(
				ErrIDRequired,
				ErrRequesterIDRequired,
				ErrNicknameRequired,
				ErrStatusRequired,
				ErrCreatedAtRequired,
				ErrUpdatedAtRequired,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder().Build()
			if err == nil && test.expectedError != nil {
				t.Errorf("expected error %v, got nil", test.expectedError)
			}
			if test.name == "Empty builder" {
				var joinErr interface{ Unwrap() []error }
				if !errors.As(err, &joinErr) {
					assert.Fail(t, "expected a join error, but got a different type", err)
				}
				actualErrors := joinErr.Unwrap()

				var expectedErr interface{ Unwrap() []error }
				if !errors.As(test.expectedError, &expectedErr) {
					assert.Fail(t, "expected a join error, but got a different type", test.expectedError)
				}
				expectedErrors := expectedErr.Unwrap()
				for _, expected := range expectedErrors {
					found := false
					for _, actual := range actualErrors {
						if errors.Is(actual, expected) {
							found = true
							break
						}
					}
					if !found {
						assert.Fail(t, "expected error %q was not found in the joined error", expected)
					}
				}
			} else {
				assert.ErrorIs(t, err, test.expectedError)
			}
		})
	}
}
