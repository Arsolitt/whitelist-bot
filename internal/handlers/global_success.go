package handlers

import (
	"context"
	"log/slog"
	"slices"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func GlobalSuccessHandler(
	cfg core.Config,
) func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State, msgParams *bot.SendMessageParams) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State, msgParams *bot.SendMessageParams) {
		if update.Message == nil {
			return
		}
		if msgParams == nil {
			return
		}
		if state == fsm.StateIdle {
			buttons := [][]models.KeyboardButton{
				{
					{Text: core.CommandInfo},
					{Text: core.CommandNewWLRequest},
					// {Text: core.CommandAnketaStart},
					// {Text: core.CommandAnketaInfo},
				},
			}
			if slices.Contains(cfg.Telegram.AdminIDs, update.Message.From.ID) {
				buttons[0] = append(buttons[0], models.KeyboardButton{Text: "Посмотреть заявки"})
			}
			if msgParams.ReplyMarkup == nil {
				slog.DebugContext(ctx, "Success handler called with new markup")
				msgParams.ReplyMarkup = &models.ReplyKeyboardMarkup{
					Keyboard:       buttons,
					ResizeKeyboard: true,
				}
			} else if msgParams.ReplyMarkup.(*models.ReplyKeyboardMarkup) != nil {
				slog.DebugContext(ctx, "Success handler called with old markup")
				oldMarkup := msgParams.ReplyMarkup.(*models.ReplyKeyboardMarkup)
				msgParams.ReplyMarkup = &models.ReplyKeyboardMarkup{
					Keyboard:              append(oldMarkup.Keyboard, buttons...),
					ResizeKeyboard:        oldMarkup.ResizeKeyboard,
					IsPersistent:          oldMarkup.IsPersistent,
					OneTimeKeyboard:       oldMarkup.OneTimeKeyboard,
					InputFieldPlaceholder: oldMarkup.InputFieldPlaceholder,
					Selective:             oldMarkup.Selective,
				}
			}
		} else if msgParams.ReplyMarkup == nil {
			msgParams.ReplyMarkup = &models.ReplyKeyboardRemove{
				RemoveKeyboard: true,
				Selective:      true,
			}
		}

		if msgParams.ChatID == nil || msgParams.ChatID == 0 {
			msgParams.ChatID = update.Message.Chat.ID
		}

		if msgParams.ParseMode == "" {
			msgParams.ParseMode = models.ParseModeHTML
		}

		_, _ = b.SendMessage(ctx, msgParams)
		slog.DebugContext(ctx, "Success handler called")
	}
}
