package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist/internal/core/logger"
	"whitelist/internal/core/utils"
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

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, dbWLRequest.ID().String())

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

func (h *Handlers) HandleApproveWLRequest(ctx context.Context, b *bot.Bot, update *models.Update) error {
	// Extract request ID from callback data (format: "approve:uuid")
	callbackData := update.CallbackQuery.Data
	requestIDStr := callbackData[8:] // Remove "approve:" prefix

	// Parse request ID
	requestID, err := utils.UUIDFromString[domainWLRequest.ID](requestIDStr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse request ID", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: неверный ID заявки",
			ShowAlert:       true,
		})
		return fmt.Errorf("failed to parse request ID: %w", err)
	}

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, requestID.String())
	slog.DebugContext(ctx, "WL request ID parsed")

	// Get request from database
	dbWLRequest, err := h.wlRequestRepo.WLRequestByID(ctx, requestID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: заявка не найдена",
			ShowAlert:       true,
		})
		return fmt.Errorf("failed to get wl request: %w", err)
	}
	slog.DebugContext(ctx, "WL request fetched from database")
	arbiter, err := h.useRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get arbiter", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: не удалось получить арбитра",
			ShowAlert:       true,
		})
		return fmt.Errorf("failed to get arbiter: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
	slog.DebugContext(ctx, "Arbiter fetched from database")

	// Update request status to approved
	updatedRequest, err := domainWLRequest.NewBuilder().
		ID(dbWLRequest.ID()).
		RequesterID(dbWLRequest.RequesterID()).
		Nickname(dbWLRequest.Nickname()).
		Status(domainWLRequest.StatusApproved).
		DeclineReason(dbWLRequest.DeclineReason()).
		ArbiterIDFromUserID(arbiter.ID()).
		CreatedAt(dbWLRequest.CreatedAt()).
		Build()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to build updated request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при обновлении заявки",
			ShowAlert:       true,
		})
		return fmt.Errorf("failed to build updated request: %w", err)
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при сохранении изменений",
			ShowAlert:       true,
		})
		return fmt.Errorf("failed to update wl request: %w", err)
	}

	// Answer callback query
	_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Заявка подтверждена!",
		ShowAlert:       false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
	}

	// Update message
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      fmt.Sprintf("✅ <b>Заявка подтверждена!</b>\n\n👤 <b>Ник:</b> %s\n🆔 <b>ID заявки:</b> <code>%s</code>", dbWLRequest.Nickname(), dbWLRequest.ID()),
		ParseMode: "HTML",
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
	}

	// TODO: Send notification to requester
	return nil
}

// TODO: rewrite routing for callback queries.
func (h *Handlers) HandleDeclineWLRequest(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Extract request ID from callback data (format: "decline:uuid")
	callbackData := update.CallbackQuery.Data
	requestIDStr := callbackData[8:] // Remove "decline:" prefix

	// Parse request ID
	requestID, err := utils.UUIDFromString[domainWLRequest.ID](requestIDStr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse request ID", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: неверный ID заявки",
			ShowAlert:       true,
		})
		return
	}

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, requestID.String())
	slog.DebugContext(ctx, "WL request ID parsed")

	// Get request from database
	dbWLRequest, err := h.wlRequestRepo.WLRequestByID(ctx, requestID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: заявка не найдена",
			ShowAlert:       true,
		})
		return
	}
	slog.DebugContext(ctx, "WL request fetched from database")
	arbiter, err := h.useRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get arbiter", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка: не удалось получить арбитра",
			ShowAlert:       true,
		})
		return
	}
	ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
	slog.DebugContext(ctx, "Arbiter fetched from database")

	// Update request status to declined
	updatedRequest, err := domainWLRequest.NewBuilder().
		ID(dbWLRequest.ID()).
		RequesterID(dbWLRequest.RequesterID()).
		Nickname(dbWLRequest.Nickname()).
		Status(domainWLRequest.StatusDeclined).
		DeclineReason(dbWLRequest.DeclineReason()).
		ArbiterIDFromUserID(arbiter.ID()).
		CreatedAt(dbWLRequest.CreatedAt()).
		Build()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to build updated request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при обновлении заявки",
			ShowAlert:       true,
		})
		return
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при сохранении изменений",
			ShowAlert:       true,
		})
		return
	}

	// Answer callback query
	_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "Заявка отклонена!",
		ShowAlert:       false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
	}

	// Update message
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      fmt.Sprintf("❌ <b>Заявка отклонена!</b>\n\n👤 <b>Ник:</b> %s\n🆔 <b>ID заявки:</b> <code>%s</code>", dbWLRequest.Nickname(), dbWLRequest.ID()),
		ParseMode: "HTML",
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
	}

	// TODO: Send notification to requester
}
