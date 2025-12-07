package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/core/utils"
	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
	"whitelist-bot/internal/metastore"
	"whitelist-bot/internal/msgs"

	eBus "whitelist-bot/internal/eventbus"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	keyPrefixNotUnique        = "not_unique"
	keyWLRequestAdminNotified = "wl_request_admin_notified"
	ttlWLRequestAdminNotified = 24 * time.Hour
)

type WLRequestCreatedEvent struct {
	ID        utils.UniqueID            `json:"id"`
	WLRequest domainWLRequest.WLRequest `json:"wl_request"`
	Requester domainUser.User           `json:"requester"`
}

func HandleWLRequestCreatedEvent(
	mg metastore.IMetastoreGetter,
	ms metastore.IMetastoreSetter,
	sender utils.IMessageSender,
	adminChatIDs []int64,
) eBus.ConsumerUnitHandler {
	return func(ctx context.Context, data []byte) error {
		var event WLRequestCreatedEvent

		err := json.Unmarshal(data, &event)
		if err != nil {
			return fmt.Errorf("failed to unmarshal wl request created event: %w", err)
		}

		ctx = logger.WithLogValue(ctx, logger.EventIDField, event.ID.String())
		ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, event.WLRequest.ID().String())
		ctx = logger.WithLogValue(ctx, logger.RequesterIDField, event.Requester.ID().String())
		slog.InfoContext(ctx, "Handling wl request created event")

		lastNotificationTimeRaw, err := mg.GetString(ctx, keyPrefixNotUnique, keyWLRequestAdminNotified)

		if err != nil && !errors.Is(err, metastore.ErrKeyNotFound) {
			return fmt.Errorf("failed to get last notification time: %w", err)
		}

		if !isRawTimeExpired(ctx, lastNotificationTimeRaw) {
			return nil
		}

		// slice of errors
		sendingErrors := make([]error, 0, len(adminChatIDs))

		for i, chatID := range adminChatIDs {
			_, err = sender.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    chatID,
				Text:      msgs.WLRequestAdminNotification(event.WLRequest),
				ParseMode: models.ParseModeHTML,
			})
			if err != nil {
				sendingErrors[i] = fmt.Errorf("failed to send wl request admin notification message: %w", err)
			}
		}
		if len(sendingErrors) == len(adminChatIDs) {
			return errors.Join(sendingErrors...)
		}

		err = ms.SetStringWithTTL(ctx, keyPrefixNotUnique, keyWLRequestAdminNotified, time.Now().Format(time.RFC3339), ttlWLRequestAdminNotified)
		if err != nil {
			return fmt.Errorf("failed to set wl request admin notified: %w", err)
		}
		return nil
	}
}

func isRawTimeExpired(ctx context.Context, rawTime string) bool {
	if len(rawTime) == 0 {
		slog.InfoContext(ctx, "Last notification time is empty")
		return true
	}
	parsedTime, err := time.Parse(time.RFC3339, string(rawTime))
	if err != nil {
		slog.WarnContext(ctx, "failed to parse last notification time", logger.ErrorField, err.Error())
		return true
	}
	return parsedTime.IsZero() || time.Since(parsedTime) > ttlWLRequestAdminNotified
}
