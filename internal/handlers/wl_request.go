package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist/internal/core/logger"
	"whitelist/internal/fsm"
	"whitelist/internal/msgs"

	domainWLRequest "whitelist/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *Handlers) NewWLRequest(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, error) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.WaitingForNickname(),
		ParseMode: "HTML",
	})
	if err != nil {
		return fsm.StateIdle, fmt.Errorf("failed to send message: %w", err)
	}
	return fsm.StateWaitingWLNickname, nil

}

func (h *Handlers) HandleWLRequestNickname(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
	// TODO: add validation for nickname. Length, special characters, etc.
	user, err := h.useRepo.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		return fsm.StateWaitingWLNickname, fmt.Errorf("failed to get user: %w", err)
	}

	wlRequest, err := domainWLRequest.NewBuilder().
		NewID().
		RequesterIDFromUserID(user.ID()).
		NicknameFromString(update.Message.Text).
		Build()
	if err != nil {
		return fsm.StateWaitingWLNickname, fmt.Errorf("failed to build wl request: %w", err)
	}

	dbWLRequest, err := h.wlRequestRepo.CreateWLRequest(ctx, wlRequest)
	if err != nil {
		return fsm.StateWaitingWLNickname, fmt.Errorf("failed to create wl request: %w", err)
	}

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, dbWLRequest.ID())

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.WLRequestCreated(dbWLRequest),
		ParseMode: "HTML",
	})
	if err != nil {
		return fsm.StateWaitingWLNickname, fmt.Errorf("failed to send message: %w", err)
	}

	// TODO: send message to admins about new wl request. Asynchronously.

	return fsm.StateIdle, err
}

func (h *Handlers) HandlePendingWLRequest(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, error) {
	wlRequest, err := h.wlRequestRepo.PendingWLRequest(ctx)
	if err != nil {
		// Если нет заявок, отправляем соответствующее сообщение
		if err.Error() == "failed to get  pending wl request: sql: no rows in result set" {
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      msgs.NoPendingWLRequests(),
				ParseMode: "HTML",
			})
			if err != nil {
				return state, fmt.Errorf("failed to send no requests message: %w", err)
			}
			return state, nil
		}
		return state, fmt.Errorf("failed to get  pending wl request: %w", err)
	}

	// Создаем inline клавиатуру с кнопками подтверждения и отказа
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "✅ Подтвердить",
					CallbackData: fmt.Sprintf("approve:%s", wlRequest.ID()),
				},
				{
					Text:         "❌ Отказать",
					CallbackData: fmt.Sprintf("decline:%s", wlRequest.ID()),
				},
			},
		},
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        msgs.PendingWLRequest(wlRequest),
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	})
	if err != nil {
		return state, fmt.Errorf("failed to send message: %w", err)
	}

	return state, nil
}

func (h *Handlers) HandleApproveWLRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: Implement approve logic
	// 1. Extract request ID from callback data
	// 2. Update request status to approved
	// 3. Send notification to requester
	// 4. Answer callback query

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Заявка подтверждена!",
		ShowAlert:       false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "✅ <b>Заявка подтверждена!</b>",
		ParseMode: "HTML",
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err)
	}
}

func (h *Handlers) HandleDeclineWLRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: Implement decline logic
	// 1. Extract request ID from callback data
	// 2. Update request status to declined
	// 3. Send notification to requester
	// 4. Answer callback query

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Заявка отклонена!",
		ShowAlert:       false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "❌ <b>Заявка отклонена!</b>",
		ParseMode: "HTML",
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err)
	}
}
