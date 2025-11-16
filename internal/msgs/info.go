package msgs

import (
	"fmt"
	"strings"
	"time"

	"whitelist/internal/domain/user"
)

func UserInfo(u user.User) string {
	var sb strings.Builder

	sb.WriteString("<b>👤 Информация о пользователе</b>\n\n")

	// Basic info
	if u.FirstName() != "" || u.LastName() != "" {
		sb.WriteString("📝 <b>Имя:</b> ")
		if u.FirstName() != "" {
			sb.WriteString(string(u.FirstName()))
		}
		if u.LastName() != "" {
			if u.FirstName() != "" {
				sb.WriteString(" ")
			}
			sb.WriteString(string(u.LastName()))
		}
		sb.WriteString("\n")
	}

	if u.Username() != "" {
		sb.WriteString(fmt.Sprintf("🔗 <b>Username:</b> @%s\n", u.Username()))
	}

	sb.WriteString(fmt.Sprintf("🆔 <b>Telegram ID:</b> <code>%d</code>\n", u.TelegramID()))
	sb.WriteString(fmt.Sprintf("🔑 <b>User ID:</b> <code>%s</code>\n", u.ID()))

	// Timestamps
	sb.WriteString("\n<b>⏰ Временные метки</b>\n")
	sb.WriteString(fmt.Sprintf("📅 <b>Создан:</b> %s\n", formatTime(u.CreatedAt())))
	sb.WriteString(fmt.Sprintf("🔄 <b>Обновлён:</b> %s\n", formatTime(u.UpdatedAt())))

	return sb.String()
}

// formatTime formats time.Time to a human-readable string
func formatTime(t time.Time) string {
	return t.Format("02.01.2006 15:04:05")
}
