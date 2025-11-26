package matcher

import (
	"encoding/json"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func MsgText(text string) bot.MatchFunc {
	return func(update *models.Update) bool {
		if update.Message == nil {
			return false
		}
		return update.Message.Text == text
	}
}

func And(matchers ...bot.MatchFunc) bot.MatchFunc {
	return func(update *models.Update) bool {
		for _, m := range matchers {
			if !m(update) {
				return false
			}
		}
		return true
	}
}

func Or(matchers ...bot.MatchFunc) bot.MatchFunc {
	return func(update *models.Update) bool {
		for _, m := range matchers {
			if m(update) {
				return true
			}
		}
		return false
	}
}

func Command(cmd string) bot.MatchFunc {
	return func(update *models.Update) bool {
		if update.Message == nil {
			return false
		}
		text := strings.TrimSpace(update.Message.Text)
		return strings.HasPrefix(text, "/"+cmd)
	}
}

func MatchTelegramIDs(ids ...int64) bot.MatchFunc {
	return func(update *models.Update) bool {
		var userID int64
		if update.Message != nil && update.Message.From != nil {
			userID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			userID = update.CallbackQuery.From.ID
		} else {
			return false
		}

		for _, id := range ids {
			if userID == id {
				return true
			}
		}
		return false
	}
}

func CallbackPrefix(prefix string) bot.MatchFunc {
	return func(update *models.Update) bool {
		if update.CallbackQuery == nil {
			return false
		}
		return strings.HasPrefix(update.CallbackQuery.Data, prefix)
	}
}

func CallbackAction(expectedAction string) bot.MatchFunc {
	return func(update *models.Update) bool {
		if update.CallbackQuery == nil {
			return false
		}
		var raw map[string]interface{}
		err := json.Unmarshal([]byte(update.CallbackQuery.Data), &raw)
		if err != nil {
			return false
		}
		action, exists := raw["action"]
		if !exists {
			return false
		}
		actionStr, ok := action.(string)
		if !ok {
			return false
		}
		return expectedAction == actionStr
	}
}
