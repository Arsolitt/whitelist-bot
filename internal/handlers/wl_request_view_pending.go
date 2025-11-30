package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/callbacks"
	"whitelist-bot/internal/core/logger"
	domainUser "whitelist-bot/internal/domain/user"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h Handlers) ViewPendingWLRequests(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	state fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	messages, err := h.preparePendingWLRequestMessages(ctx)
	if err != nil {
		return state, nil, err
	}

	if len(messages) == 0 {
		msgParams := &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      msgs.NoPendingWLRequests(),
			ParseMode: "HTML",
		}
		return state, msgParams, nil
	}

	for _, msg := range messages {
		_, err = h.botSendMessage(ctx, b, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        msg.Text,
			ParseMode:   "HTML",
			ReplyMarkup: msg.ReplyMarkup,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to send message", logger.ErrorField, err.Error())
			continue
		}
	}

	return state, nil, nil
}

func (h Handlers) preparePendingWLRequestMessages(ctx context.Context) ([]pendingWLRequestMessage, error) {
	wlRequests, err := h.wlRequestRepo.PendingWLRequests(ctx, PENDING_WL_REQUESTS_LIMIT)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending wl requests: %w", err)
	}

	if len(wlRequests) == 0 {
		return nil, nil
	}

	messages := make([]pendingWLRequestMessage, 0, len(wlRequests))
	for _, wlRequest := range wlRequests {
		requester, err := h.useRepo.UserByID(ctx, domainUser.ID(wlRequest.RequesterID()))
		if err != nil {
			return nil, fmt.Errorf("failed to get requester: %w", err)
		}

		keyboard := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "✅ Подтвердить",
						CallbackData: callbacks.ApproveWLRequestData(ctx, wlRequest.ID()),
					},
					{
						Text:         "❌ Отказать",
						CallbackData: callbacks.DeclineWLRequestData(ctx, wlRequest.ID()),
					},
				},
			},
		}

		messages = append(messages, pendingWLRequestMessage{
			Text:        msgs.PendingWLRequest(wlRequest, requester),
			ReplyMarkup: keyboard,
		})
	}

	return messages, nil
}
