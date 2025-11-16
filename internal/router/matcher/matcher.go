package matcher

import (
	"context"
	"whitelist/internal/fsm"
	"whitelist/internal/router"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func Text(text string) router.MatcherFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		return update.Message.Text == text
	}
}

func State(expectedState fsm.State) router.MatcherFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		return state == expectedState
	}
}

func And(matchers ...router.MatcherFunc) router.MatcherFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		for _, m := range matchers {
			if !m(ctx, b, update, state) {
				return false
			}
		}
		return true
	}
}

func Or(matchers ...router.MatcherFunc) router.MatcherFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update, state fsm.State) bool {
		for _, m := range matchers {
			if m(ctx, b, update, state) {
				return true
			}
		}
		return false
	}
}

func Command(cmd string) router.MatcherFunc {
	return Text("/" + cmd)
}
