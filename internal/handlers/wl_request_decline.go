package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func DeclineWLRequest(
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

		// TODO: ??????????????
		if !callbackData.IsDecline() {
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("неверный action"),
			}, nil)
			return state, response, fmt.Errorf("invalid action: expected decline, got %s", callbackData.Action())
		}

		ctx = logger.WithLogValue(ctx, logger.WLRequestIDField, callbackData.ID().String())

		dbWLRequest, err := wlRequestRepo.WLRequestByID(ctx, callbackData.ID())
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get wl request", logger.ErrorField, err.Error())
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("заявка не найдена"),
			}, nil)
			return state, response, fmt.Errorf("failed to get wl request: %w", err)
		}
		slog.DebugContext(ctx, "WL request fetched from database")

		arbiter, err := userRepo.UserByTelegramID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get arbiter", logger.ErrorField, err.Error())
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

		declinedRequest, err := dbWLRequest.Decline(
			domainWLRequest.ArbiterID(arbiter.ID()),
			domainWLRequest.DeclineReason("Отклонено администратором"),
		)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to decline wl request", logger.ErrorField, err.Error())
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("ошибка при обновлении заявки"),
			}, nil)
			return state, response, fmt.Errorf("failed to decline wl request: %w", err)
		}

		_, err = wlRequestRepo.UpdateWLRequest(ctx, declinedRequest)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update wl request", logger.ErrorField, err.Error())
			response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
				Text: msgs.CallbackError("ошибка при сохранении изменений"),
			}, nil)
			return state, response, fmt.Errorf("failed to update wl request: %w", err)
		}

		response := router.NewCallbackResponse(&bot.AnswerCallbackQueryParams{
			Text: "❌ Заявка отклонена!",
		}, &bot.EditMessageTextParams{
			Text: msgs.DeclinedWLRequest(declinedRequest, arbiter, requester),
		})

		return state, response, nil
	}
}
