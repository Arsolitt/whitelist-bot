package handlers

import (
	"context"
	"fmt"
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
