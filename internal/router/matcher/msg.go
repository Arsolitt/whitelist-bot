package matcher

// import (
// 	"context"
// 	"whitelist/internal/fsm"
// 	"whitelist/internal/router"

// 	"github.com/go-telegram/bot"
// 	"github.com/go-telegram/bot/models"
// )

// func Text(text string) router.MsgMatcherFunc {
// 	return func(_ context.Context, _ *bot.Bot, update *models.Update, _ fsm.State) bool {
// 		return update.Message.Text == text
// 	}
// }

// func State(expectedState fsm.State) router.MsgMatcherFunc {
// 	return func(_ context.Context, _ *bot.Bot, _ *models.Update, state fsm.State) bool {
// 		return state == expectedState
// 	}
// }

// func And(matchers ...router.MsgMatcherFunc) router.MsgMatcherFunc {
// 	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
// 		for _, m := range matchers {
// 			if !m(ctx, b, update, state) {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// }

// func Or(matchers ...router.MsgMatcherFunc) router.MsgMatcherFunc {
// 	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
// 		for _, m := range matchers {
// 			if m(ctx, b, update, state) {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }

// func Command(cmd string) router.MsgMatcherFunc {
// 	return Text("/" + cmd)
// }

// func MatchTelegramIDs(ids ...int64) router.MsgMatcherFunc {
// 	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
// 		for _, id := range ids {
// 			if update.Message.From.ID == id {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }
