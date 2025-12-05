package router

import (
	"context"
	"log/slog"
	"slices"
	"whitelist-bot/internal/core"
	"whitelist-bot/internal/core/logger"
	"whitelist-bot/internal/core/utils"
	"whitelist-bot/internal/fsm"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Response interface {
	Answer(ctx context.Context, sender utils.IMessageSender, update *models.Update, currentState fsm.State, cfg core.Config) error
}

type MessageResponse struct {
	Params []*bot.SendMessageParams
}

func (r *MessageResponse) AddMessage(p *bot.SendMessageParams) {
	r.Params = append(r.Params, p)
}

func NewMessageResponse(params ...*bot.SendMessageParams) *MessageResponse {
	if len(params) == 0 {
		return &MessageResponse{
			Params: make([]*bot.SendMessageParams, 0),
		}
	}
	return &MessageResponse{
		Params: params,
	}
}

func (r *MessageResponse) Answer(ctx context.Context, sender utils.IMessageSender, update *models.Update, currentState fsm.State, cfg core.Config) error {
	if update.Message == nil {
		return core.ErrInvalidUpdate
	}
	for _, p := range r.Params {
		if p == nil {
			return nil
		}
		if p.ChatID == nil {
			p.ChatID = update.Message.Chat.ID
		}
		if p.ParseMode == "" {
			p.ParseMode = models.ParseModeHTML
		}
		if currentState == fsm.StateIdle {
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
			if p.ReplyMarkup == nil {
				slog.DebugContext(ctx, "Success handler called with new markup")
				p.ReplyMarkup = &models.ReplyKeyboardMarkup{
					Keyboard:       buttons,
					ResizeKeyboard: true,
				}
			} else if oldMarkup, ok := p.ReplyMarkup.(*models.ReplyKeyboardMarkup); ok {
				slog.DebugContext(ctx, "Success handler called with old markup")
				p.ReplyMarkup = &models.ReplyKeyboardMarkup{
					Keyboard:              append(oldMarkup.Keyboard, buttons...),
					ResizeKeyboard:        oldMarkup.ResizeKeyboard,
					IsPersistent:          oldMarkup.IsPersistent,
					OneTimeKeyboard:       oldMarkup.OneTimeKeyboard,
					InputFieldPlaceholder: oldMarkup.InputFieldPlaceholder,
					Selective:             oldMarkup.Selective,
				}
			}
		} else if p.ReplyMarkup == nil {
			p.ReplyMarkup = &models.ReplyKeyboardRemove{
				RemoveKeyboard: true,
				Selective:      true,
			}
		}
		_, err := sender.SendMessage(ctx, p)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to send message", logger.ErrorField, err.Error())
			continue
		}
	}
	return nil
}

type CallbackResponse struct {
	CallbackParams *bot.AnswerCallbackQueryParams
	EditParams     *bot.EditMessageTextParams
}

func NewCallbackResponse(callbackParams *bot.AnswerCallbackQueryParams, editParams *bot.EditMessageTextParams) *CallbackResponse {
	return &CallbackResponse{
		CallbackParams: callbackParams,
		EditParams:     editParams,
	}
}

func (r *CallbackResponse) Answer(ctx context.Context, sender utils.IMessageSender, update *models.Update, currentState fsm.State, cfg core.Config) error {
	if r.CallbackParams != nil {
		if r.CallbackParams.CallbackQueryID == "" {
			r.CallbackParams.CallbackQueryID = update.CallbackQuery.ID
		}
		_, err := sender.AnswerCallbackQuery(ctx, r.CallbackParams)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
		}
		r.CallbackParams.ShowAlert = true
	}
	if r.EditParams != nil {
		if r.EditParams.ChatID == nil {
			r.EditParams.ChatID = update.CallbackQuery.Message.Message.Chat.ID
		}
		if r.EditParams.MessageID == 0 {
			r.EditParams.MessageID = update.CallbackQuery.Message.Message.ID
		}
		if r.EditParams.ParseMode == "" {
			r.EditParams.ParseMode = models.ParseModeHTML
		}
		_, err := sender.EditMessageText(ctx, r.EditParams)
		return err
	}
	return nil
}
