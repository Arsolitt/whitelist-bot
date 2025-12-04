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
	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, router.Response, error) {
		callbackData, err := parseCallbackData(update.CallbackQuery.Data)
		if err != nil {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("неверный формат callback data"),
			}, nil)
			return state, response, fmt.Errorf("failed to unmarshal callback data: %w", err)
		}

		if !callbackData.IsApprove() {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("неверный action"),
			}, nil)
			return state, response, fmt.Errorf("invalid action: expected approve, got %s", callbackData.Action())
		}

		ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, callbackData.ID().String())

		dbWLRequest, err := wlRequestRepo.WLRequestByID(ctx, callbackData.ID())
		if err != nil {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("заявка не найдена"),
			}, nil)
			return state, response, fmt.Errorf("failed to get wl request: %w", err)
		}
		slog.DebugContext(ctx, "WL request fetched from database")

		arbiter, err := userRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("не удалось получить арбитра"),
			}, nil)
			return state, response, fmt.Errorf("failed to get arbiter: %w", err)
		}
		ctx = logger.WithLogValue(ctx, logger.ArbiterIDField, arbiter.ID().String())
		slog.DebugContext(ctx, "Arbiter fetched from database")

		requester, err := userRepo.UserByID(ctx, domainUser.ID(dbWLRequest.RequesterID()))
		if err != nil {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("не удалось получить заявителя"),
			}, nil)
			return state, response, fmt.Errorf("failed to get requester: %w", err)
		}
		ctx = logger.WithLogValue(ctx, logger.RequesterIDField, requester.ID().String())
		slog.DebugContext(ctx, "Requester fetched from database")

		updatedRequest, err := dbWLRequest.Approve(domainWLRequest.ArbiterID(arbiter.ID()))
		if err != nil {
			response := router.NewMessageResponse(&bot.SendMessageParams{
				Text: msgs.CallbackError("ошибка при обновлении заявки"),
			})
			return state, response, fmt.Errorf("failed to build updated request: %w", err)
		}

		_, err = wlRequestRepo.UpdateWLRequest(ctx, updatedRequest)
		if err != nil {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("ошибка при сохранении изменений"),
			}, nil)
			return state, response, fmt.Errorf("failed to update wl request: %w", err)
		}

		response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
			Text: "✅ Заявка подтверждена",
		}, &bot.EditMessageTextParams{
			Text: msgs.ApprovedWLRequest(dbWLRequest, arbiter, requester),
		})

		return state, response, nil
	}
}

func parseCallbackData(data string) (callbacks.WLRequestCallbackData, error) {
	var callbackData callbacks.WLRequestCallbackData
	err := json.Unmarshal([]byte(data), &callbackData)
	return callbackData, err
}
