package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/callbacks"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func ViewPendingWLRequests(
	wlRequestRepo iWLRequestRepository,
	sender iMessageSender,
) router.HandlerFunc {
	preparePendingWLRequestMessages := func(ctx context.Context) ([]pendingWLRequestMessage, error) {
		wlRequests, err := wlRequestRepo.PendingWLRequestsWithRequester(ctx, PENDING_WL_REQUESTS_LIMIT)
		if err != nil {
			return nil, fmt.Errorf("failed to get pending wl requests: %w", err)
		}

		if len(wlRequests) == 0 {
			return nil, nil
		}

		messages := make([]pendingWLRequestMessage, 0, len(wlRequests))
		for _, wlRequest := range wlRequests {

			keyboard := &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{
							Text:         "✅ Подтвердить",
							CallbackData: callbacks.ApproveWLRequestData(ctx, wlRequest.WlRequest.ID()),
						},
						{
							Text:         "❌ Отказать",
							CallbackData: callbacks.DeclineWLRequestData(ctx, wlRequest.WlRequest.ID()),
						},
					},
				},
			}

			messages = append(messages, pendingWLRequestMessage{
				Text:        msgs.PendingWLRequest(wlRequest.WlRequest, wlRequest.User),
				ReplyMarkup: keyboard,
			})
		}

		return messages, nil
	}

	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
		messages, err := preparePendingWLRequestMessages(ctx)
		if err != nil {
			return state, nil, err
		}

		if len(messages) == 0 {
			msgParams := &bot.SendMessageParams{
				Text: msgs.NoPendingWLRequests(),
			}
			return state, msgParams, nil
		}

		// TODO: return multiple messages instead if send one by one.
		for _, msg := range messages {
			_, err = sender.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        msg.Text,
				ReplyMarkup: msg.ReplyMarkup,
				ParseMode:   models.ParseModeHTML,
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to send message", logger.ErrorField, err.Error())
				continue
			}
		}

		return state, nil, nil
	}
}
