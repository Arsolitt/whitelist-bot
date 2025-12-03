package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/callbacks"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func ApproveWLRequest(
	userRepo iUserRepository,
	wlRequestRepo iWLRequestRepository,
) router.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
		callbackData, err := parseCallbackData(update.CallbackQuery.Data)
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "неверный формат callback data")
			return state, nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
		}

		if !callbackData.IsApprove() {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "неверный action")
			return state, nil, fmt.Errorf("invalid action: expected approve, got %s", callbackData.Action())
		}

		ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, callbackData.ID().String())

		dbWLRequest, err := wlRequestRepo.WLRequestByID(ctx, callbackData.ID())
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "заявка не найдена")
			return state, nil, fmt.Errorf("failed to get wl request: %w", err)
		}
		slog.DebugContext(ctx, "WL request fetched from database")

		arbiter, err := userRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "не удалось получить арбитра")
			return state, nil, fmt.Errorf("failed to get arbiter: %w", err)
		}
		ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
		slog.DebugContext(ctx, "Arbiter fetched from database")

		requester, err := userRepo.UserByID(ctx, domainUser.ID(dbWLRequest.RequesterID()))
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "не удалось получить заявителя")
			return state, nil, fmt.Errorf("failed to get requester: %w", err)
		}
		ctx = logger.WithLogValue(ctx, logger.RequesterIDField, requester.ID().String())
		slog.DebugContext(ctx, "Requester fetched from database")

		updatedRequest, err := dbWLRequest.Approve(domainWLRequest.ArbiterID(arbiter.ID()))
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при обновлении заявки")
			return state, nil, fmt.Errorf("failed to build updated request: %w", err)
		}

		_, err = wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при сохранении изменений")
			return state, nil, fmt.Errorf("failed to update wl request: %w", err)
		}

		_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
			Text:      msgs.ApprovedWLRequest(dbWLRequest, arbiter, requester),
			ParseMode: "HTML",
		})
		if err != nil {
			sendCallbackError(ctx, b, update.CallbackQuery.ID, "ошибка при редактировании сообщения")
			return state, nil, fmt.Errorf("failed to edit message: %w", err)
		}

		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "✅ Заявка подтверждена",
			ShowAlert:       false,
		})

		return state, nil, nil
	}
}

func parseCallbackData(data string) (callbacks.WLRequestCallbackData, error) {
	var callbackData callbacks.WLRequestCallbackData
	err := json.Unmarshal([]byte(data), &callbackData)
	return callbackData, err
}

func sendCallbackError(ctx context.Context, sender iMessageSender, callbackQueryID, errorText string) {
	_, _ = sender.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQueryID,
		Text:            msgs.CallbackError(errorText),
		ShowAlert:       true,
	})
}
