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

const PENDING_WL_REQUESTS_LIMIT = 5

func (h *Handlers) NewWLRequest(ctx context.Context, b *bot.Bot, update *models.Update, _ fsm.State) (fsm.State, *bot.SendMessageParams, error) {
	msgParams := bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.WaitingForNickname(),
		ParseMode: "HTML",
	}
	return fsm.StateWaitingWLNickname, &msgParams, nil
}

func (h *Handlers) SubmitWLRequestNickname(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
	// TODO: add validation for nickname. Length, special characters, etc.
	user, err := h.useRepo.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		return fsm.StateWaitingWLNickname, nil, fmt.Errorf("failed to get user: %w", err)
	}

	nickname := ""
	if update.Message != nil && update.Message.Text != "" {
		nickname = update.Message.Text
	}

	dbWLRequest, err := h.wlRequestRepo.CreateWLRequest(ctx, domainWLRequest.RequesterID(user.ID()), domainWLRequest.Nickname(nickname))
	if err != nil {
		return fsm.StateWaitingWLNickname, nil, fmt.Errorf("failed to create wl request: %w", err)
	}

	logger.WithLogValue(ctx, logger.WLRequestIDField, dbWLRequest.ID().String())

	msgParams := bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.WLRequestCreated(dbWLRequest),
		ParseMode: "HTML",
	}
	return fsm.StateIdle, &msgParams, nil
}

func (h *Handlers) ViewPendingWLRequests(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
	wlRequests, err := h.wlRequestRepo.PendingWLRequests(ctx, PENDING_WL_REQUESTS_LIMIT)
	if err != nil {
		return state, nil, fmt.Errorf("failed to get  pending wl request: %w", err)
	}
	if len(wlRequests) == 0 {
		msgParams := &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      msgs.NoPendingWLRequests(),
			ParseMode: "HTML",
		}
		return state, msgParams, nil
	}

	for _, wlRequest := range wlRequests {
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
			slog.ErrorContext(ctx, "Failed to send message", logger.ErrorField, err.Error())
			continue
		}
	}

	return state, nil, nil
}

func (h *Handlers) HandleApproveWLRequest(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
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
		return state, nil, fmt.Errorf("failed to parse request ID: %w", err)
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
		return state, nil, fmt.Errorf("failed to get wl request: %w", err)
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
		return state, nil, fmt.Errorf("failed to get arbiter: %w", err)
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
		return state, nil, fmt.Errorf("failed to build updated request: %w", err)
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при сохранении изменений",
			ShowAlert:       true,
		})
		return state, nil, fmt.Errorf("failed to update wl request: %w", err)
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
	return state, nil, nil
}

// TODO: rewrite routing for callback queries.
func (h *Handlers) HandleDeclineWLRequest(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
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
		return fsm.StateIdle, nil, fmt.Errorf("failed to parse request ID: %w", err)
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
		return fsm.StateIdle, nil, fmt.Errorf("failed to get wl request: %w", err)
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
		return fsm.StateIdle, nil, fmt.Errorf("failed to get arbiter: %w", err)
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
		return fsm.StateIdle, nil, fmt.Errorf("failed to build updated request: %w", err)
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Ошибка при сохранении изменений",
			ShowAlert:       true,
		})
		return fsm.StateIdle, nil, fmt.Errorf("failed to update wl request: %w", err)
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
	return fsm.StateIdle, nil, nil
}
