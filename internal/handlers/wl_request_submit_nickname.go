package handlers

import (
	"context"
	"fmt"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	domainWLRequest "whitelist-bot/internal/domain/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func SubmitWLRequestNickname(
	userRepo iUserRepository,
	wlRequestRepo iWLRequestRepository,
) router.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update, state fsm.State) (fsm.State, *bot.SendMessageParams, error) {
		// TODO: add validation for nickname. Length, special characters, etc.
		user, err := userRepo.UserByTelegramID(ctx, update.Message.From.ID)
		if err != nil {
			return fsm.StateWaitingWLNickname, nil, fmt.Errorf("failed to get user: %w", err)
		}

		nickname := ""
		if update.Message != nil && update.Message.Text != "" {
			nickname = update.Message.Text
		}

		dbWLRequest, err := wlRequestRepo.CreateWLRequest(
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
}
