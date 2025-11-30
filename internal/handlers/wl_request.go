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

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const PENDING_WL_REQUESTS_LIMIT = 5

// TODO: rewrite routing for callback queries.

type pendingWLRequestMessage struct {
	Text        string
	ReplyMarkup *models.InlineKeyboardMarkup
}

func (h Handlers) NewWLRequest(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	_ fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	msgParams := bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msgs.WaitingForNickname(),
		ParseMode: "HTML",
	}
	return fsm.StateWaitingWLNickname, &msgParams, nil
}

func (h Handlers) SubmitWLRequestNickname(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	state fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	// TODO: add validation for nickname. Length, special characters, etc.
	user, err := h.useRepo.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		return fsm.StateWaitingWLNickname, nil, fmt.Errorf("failed to get user: %w", err)
	}

	nickname := ""
	if update.Message != nil && update.Message.Text != "" {
		nickname = update.Message.Text
	}

	dbWLRequest, err := h.wlRequestRepo.CreateWLRequest(
		ctx,
		domainWLRequest.RequesterID(user.ID()),
		domainWLRequest.Nickname(nickname),
	)
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

func (h Handlers) DeclineWLRequest(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	state fsm.State,
) (fsm.State, *bot.SendMessageParams, error) {
	var callbackData callbacks.WLRequestCallbackData
	err := json.Unmarshal([]byte(update.CallbackQuery.Data), &callbackData)
	if err != nil {
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("неверный формат callback data"),
			ShowAlert:       true,
		})
		return state, nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
	}
	if !callbackData.IsDecline() {
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("неверный action"),
			ShowAlert:       true,
		})
		return state, nil, fmt.Errorf("failed to unmarshal callback data: %w", err)
	}

	ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, callbackData.ID().String())

	dbWLRequest, err := h.wlRequestRepo.WLRequestByID(ctx, callbackData.ID())
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get wl request", logger.ErrorField, err.Error())
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("заявка не найдена"),
			ShowAlert:       true,
		})
		return fsm.StateIdle, nil, fmt.Errorf("failed to get wl request: %w", err)
	}
	slog.DebugContext(ctx, "WL request fetched from database")
	arbiter, err := h.useRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get arbiter", logger.ErrorField, err.Error())
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("не удалось получить арбитра"),
			ShowAlert:       true,
		})
		return fsm.StateIdle, nil, fmt.Errorf("failed to get arbiter: %w", err)
	}
	ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
	slog.DebugContext(ctx, "Arbiter fetched from database")

	requester, err := h.useRepo.UserByID(ctx, domainUser.ID(dbWLRequest.RequesterID()))
	if err != nil {
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("не удалось получить заявителя"),
			ShowAlert:       true,
		})
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
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("ошибка при обновлении заявки"),
			ShowAlert:       true,
		})
		return fsm.StateIdle, nil, fmt.Errorf("failed to decline wl request: %w", err)
	}

	_, err = h.wlRequestRepo.UpdateWLRequest(ctx, declinedRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("ошибка при сохранении изменений"),
			ShowAlert:       true,
		})
		return fsm.StateIdle, nil, fmt.Errorf("failed to update wl request: %w", err)
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      msgs.DeclinedWLRequest(declinedRequest, arbiter, requester),
		ParseMode: "HTML",
	})
	if err != nil {
		_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            msgs.CallbackError("ошибка при редактировании сообщения"),
			ShowAlert:       true,
		})
		return state, nil, fmt.Errorf("failed to edit message: %w", err)
	}
	// TODO: ?????????????
	_, _ = h.botAnswerCallbackQuery(ctx, b, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "❌ Заявка отклонена!",
		ShowAlert:       false,
	})

	// TODO: Send notification to requester
	return fsm.StateIdle, nil, nil
}
