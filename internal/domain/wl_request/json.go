package wl_request

import (
	"encoding/json"
	"time"
)

func (w WLRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID            ID            `json:"id"`
		RequesterID   RequesterID   `json:"requester_id"`
		Nickname      Nickname      `json:"nickname"`
		Status        Status        `json:"status"`
		DeclineReason DeclineReason `json:"decline_reason"`
		ArbiterID     ArbiterID     `json:"arbiter_id"`
		CreatedAt     time.Time     `json:"created_at"`
		UpdatedAt     time.Time     `json:"updated_at"`
	}{
		ID:            w.id,
		RequesterID:   w.requesterID,
		Nickname:      w.nickname,
		Status:        w.status,
		DeclineReason: w.declineReason,
		ArbiterID:     w.arbiterID,
		CreatedAt:     w.createdAt,
		UpdatedAt:     w.updatedAt,
	})
}

func (w *WLRequest) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID            ID            `json:"id"`
		RequesterID   RequesterID   `json:"requester_id"`
		Nickname      Nickname      `json:"nickname"`
		Status        Status        `json:"status"`
		DeclineReason DeclineReason `json:"decline_reason"`
		ArbiterID     ArbiterID     `json:"arbiter_id"`
		CreatedAt     time.Time     `json:"created_at"`
		UpdatedAt     time.Time     `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	wlRequest, err := NewBuilder().
		ID(aux.ID).
		RequesterID(aux.RequesterID).
		Nickname(aux.Nickname).
		Status(aux.Status).
		DeclineReason(aux.DeclineReason).
		ArbiterID(aux.ArbiterID).
		CreatedAt(aux.CreatedAt).
		UpdatedAt(aux.UpdatedAt).
		Build()
	if err != nil {
		return err
	}

	*w = wlRequest
	return nil
}
