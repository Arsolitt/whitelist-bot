package handlers

import (
	"context"
	"fmt"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"
	"whitelist-bot/internal/msgs"
	"whitelist-bot/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Info(
	userGetter iUserGetter,
) router.HandlerFunc {
	return func(
		ctx context.Context,
		b *bot.Bot,
		update *models.Update,
		currentState fsm.State,
	) (fsm.State, *bot.SendMessageParams, error) {
		if currentState != fsm.StateIdle {
			return currentState, nil, core.ErrInvalidUserState
		}

		user, err := userGetter.UserByTelegramID(ctx, update.Message.From.ID)
		if err != nil {
			return fsm.StateIdle, nil, fmt.Errorf("failed to get user: %w", err)
		}

		msgParams := &bot.SendMessageParams{
			Text: msgs.UserInfo(user),
		}
		return fsm.StateIdle, msgParams, nil
	}
}
