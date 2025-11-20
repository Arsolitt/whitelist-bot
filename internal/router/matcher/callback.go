package matcher

// import (
// 	"context"
// 	"strings"
// 	"whitelist/internal/router"

// 	"github.com/go-telegram/bot"
// 	"github.com/go-telegram/bot/models"
// )

// func CallbackCommand(cmd string) router.CallbackMatcherFunc {
// 	return func(_ context.Context, _ *bot.Bot, update *models.Update) bool {
// 		return update.CallbackQuery.Data == cmd
// 	}
// }

// func CallbackPrefix(prefix string) router.CallbackMatcherFunc {
// 	return func(_ context.Context, _ *bot.Bot, update *models.Update) bool {
// 		return strings.HasPrefix(update.CallbackQuery.Data, prefix)
// 	}
// }

// func CallbackMatchTelegramIDs(ids ...int64) router.CallbackMatcherFunc {
// 	return func(_ context.Context, _ *bot.Bot, update *models.Update) bool {
// 		for _, id := range ids {
// 			if update.CallbackQuery.From.ID == id {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }
