package callbacks

import (
	"context"
	"encoding/json"
	"log/slog"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/core/utils"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
)

type WLRequestCallbackData struct {
	id     domainWLRequest.ID
	action string
}

func (c WLRequestCallbackData) Action() string {
	return c.action
}

func (c WLRequestCallbackData) IsApprove() bool {
	return c.action == core.ActionWLRequestApprove
}

func (c WLRequestCallbackData) IsDecline() bool {
	return c.action == core.ActionWLRequestDecline
}

func (c WLRequestCallbackData) ID() domainWLRequest.ID {
	return c.id
}

func (c WLRequestCallbackData) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID     string `json:"id"`
		Action string `json:"action"`
	}{
		ID:     c.id.String(),
		Action: c.action,
	}

	return json.Marshal(aux)
}

func (c *WLRequestCallbackData) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID     string `json:"id"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	id, err := utils.UUIDFromString[domainWLRequest.ID](aux.ID)
	if err != nil {
		return err
	}
	c.id = id
	c.action = aux.Action
	return nil
}

func NewWLRequestCallbackData(id domainWLRequest.ID, action string) WLRequestCallbackData {
	return WLRequestCallbackData{
		id:     id,
		action: action,
	}
}

func ApproveWLRequestData(ctx context.Context, id domainWLRequest.ID) string {
	json, err := json.Marshal(NewWLRequestCallbackData(id, core.ActionWLRequestApprove))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal approve WL request data", logger.ErrorField, err.Error())
		return ""
	}
	slog.DebugContext(ctx, "Approve WL request data marshalled", "data", string(json))
	return string(json)
}

func DeclineWLRequestData(ctx context.Context, id domainWLRequest.ID) string {
	json, err := json.Marshal(NewWLRequestCallbackData(id, core.ActionWLRequestDecline))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal decline WL request data", logger.ErrorField, err.Error())
		return ""
	}
	slog.DebugContext(ctx, "Decline WL request data marshalled", "data", string(json))
	return string(json)
}
