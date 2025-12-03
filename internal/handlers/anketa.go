package handlers

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
// 	"strings"
// 	"whitelist-bot/internal/fsm"
// 	"whitelist-bot/internal/metastore"

// 	"github.com/go-telegram/bot"
// 	"github.com/go-telegram/bot/models"
// )

// type anketaData struct {
// 	Name string `json:"name"`
// 	Age  int    `json:"age"`
// }

// func (h *Handlers) AnketaStart(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, *bot.SendMessageParams, error) {
// 	msgParams := &bot.SendMessageParams{
// 		ChatID:    update.Message.Chat.ID,
// 		Text:      "Привет! Как тебя зовут?",
// 		ParseMode: "HTML",
// 	}

// 	return fsm.StateAnketaName, msgParams, nil
// }

// func (h *Handlers) AnketaName(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, *bot.SendMessageParams, error) {
// 	user, err := h.userRepo.UserByTelegramID(ctx, update.Message.From.ID)
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to get user: %w", err)
// 	}

// 	h.metastore.Set(ctx, user.ID().String(), "anketa", anketaData{
// 		Name: update.Message.Text,
// 	})

// 	msgParams := &bot.SendMessageParams{
// 		ChatID:    update.Message.Chat.ID,
// 		Text:      "Сколько тебе лет?",
// 		ParseMode: "HTML",
// 	}

// 	return fsm.StateAnketaAge, msgParams, nil
// }

// func (h *Handlers) AnketaAge(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, *bot.SendMessageParams, error) {
// 	user, err := h.userRepo.UserByTelegramID(ctx, update.Message.From.ID)
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to get user: %w", err)
// 	}

// 	anketa, err := metastore.TypedJSONMeta[anketaData](ctx, h.metastore, user.ID().String(), "anketa")
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to get anketa: %w", err)
// 	}
// 	anketa.Age, err = strconv.Atoi(update.Message.Text)
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to convert age to int: %w", err)
// 	}
// 	h.metastore.Set(ctx, user.ID().String(), "anketa", anketa)

// 	msgParams := &bot.SendMessageParams{
// 		ChatID:    update.Message.Chat.ID,
// 		Text:      anketaMsg(anketa),
// 		ParseMode: "HTML",
// 	}

// 	return fsm.StateIdle, msgParams, nil
// }

// func (h *Handlers) AnketaInfo(ctx context.Context, b *bot.Bot, update *models.Update, currentState fsm.State) (fsm.State, *bot.SendMessageParams, error) {
// 	user, err := h.userRepo.UserByTelegramID(ctx, update.Message.From.ID)
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to get user: %w", err)
// 	}

// 	anketa, err := metastore.TypedJSONMeta[anketaData](ctx, h.metastore, user.ID().String(), "anketa")
// 	if err != nil {
// 		return currentState, nil, fmt.Errorf("failed to get anketa: %w", err)
// 	}
// 	msgParams := &bot.SendMessageParams{
// 		ChatID:    update.Message.Chat.ID,
// 		Text:      anketaMsg(anketa),
// 		ParseMode: "HTML",
// 	}
// 	return currentState, msgParams, nil
// }

// func anketaMsg(anketa anketaData) string {
// 	var sb strings.Builder

// 	sb.WriteString("<b>👤 Анкета</b>\n\n")

// 	if anketa.Name != "" {
// 		sb.WriteString("📝 <b>Имя:</b> ")
// 		sb.WriteString(anketa.Name)
// 		sb.WriteString("\n")
// 	}
// 	if anketa.Age != 0 {
// 		sb.WriteString("📝 <b>Возраст:</b> ")
// 		sb.WriteString(strconv.Itoa(anketa.Age))
// 		sb.WriteString("\n")
// 	}

// 	return sb.String()
// }
