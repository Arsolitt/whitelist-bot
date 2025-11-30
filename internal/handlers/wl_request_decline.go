package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h Handlers) DeclineWLRequest(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	state fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	callbackData, err := h.parseCallbackData(update.CallbackQuery.Data)
	if err != nil {
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "неверный формат callback data")
		return state, nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
	}

	if !callbackData.IsDecline() {
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "неверный action")
		return state, nil, fmt.Errorf("invalid action: expected decline, got %s", callbackData.Action())
	}

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, callbackData.ID().String())

	dbWLRequest, err := h.wlRequestRepo.WLRequestByID(ctx, callbackData.ID())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get wl request", logger.ErrorField, err.Error())
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "заявка не найдена")
		return state, nil, fmt.Errorf("failed to get wl request: %w", err)
	}
	slog.DebugContext(ctx, "WL request fetched from database")

	arbiter, err := h.userRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get arbiter", logger.ErrorField, err.Error())
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "не удалось получить арбитра")
		return state, nil, fmt.Errorf("failed to get arbiter: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
	slog.DebugContext(ctx, "Arbiter fetched from database")

	requester, err := h.userRepo.UserByID(ctx, domainUser.ID(dbWLRequest.RequesterID()))
	if err != nil {
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "не удалось получить заявителя")
		return state, nil, fmt.Errorf("failed to get requester: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.RequesterIDField, requester.ID().String())
	slog.DebugContext(ctx, "Requester fetched from database")

	declinedRequest, err := dbWLRequest.Decline(
		domainWLRequest.ArbiterID(arbiter.ID()),
		domainWLRequest.DeclineReason("Отклонено администратором"),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decline wl request", logger.ErrorField, err.Error())
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при обновлении заявки")
		return state, nil, fmt.Errorf("failed to decline wl request: %w", err)
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, declinedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при сохранении изменений")
		return state, nil, fmt.Errorf("failed to update wl request: %w", err)
	}

	_, err = h.botEditMessageText(ctx, b, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      msgs.DeclinedWLRequest(declinedRequest, arbiter, requester),
		ParseMode: "HTML",
	})
	if err != nil {
		h.sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при редактировании сообщения")
		return state, nil, fmt.Errorf("failed to edit message: %w", err)
	}

	_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "❌ Заявка отклонена!",
		ShowAlert:       false,
	})

	return state, nil, nil
}
