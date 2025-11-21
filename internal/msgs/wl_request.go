package msgs

import (
	"fmt"
	"strings"
	domainUser "whitelist/internal/domain/user"
	domainWLRequest "whitelist/internal/domain/wl_request"
)

func WaitingForNickname() string {
	return "Привет! Отправь свой ник, чтобы подать заявку в белый список."
}

func WLRequestCreated(wlRequest domainWLRequest.WLRequest) string {
	var sb strings.Builder
	sb.WriteString("<b>Заявка в белый список успешно отправлена</b>\n\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Ник:</b> %s\n", wlRequest.Nickname()))
	sb.WriteString(fmt.Sprintf("🆔 <b>ID:</b> %s\n", wlRequest.ID()))
	sb.WriteString(fmt.Sprintf("🔄 <b>Статус:</b> %s\n", wlRequest.Status()))
	sb.WriteString(fmt.Sprintf("🔄 <b>Создано:</b> %s\n", wlRequest.CreatedAt()))
	return sb.String()
}

func PendingWLRequest(wlRequest domainWLRequest.WLRequest) string {
	var sb strings.Builder
	sb.WriteString("📋 <b>Ожидающая заявка</b>\n\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Ник:</b> %s\n", wlRequest.Nickname()))
	sb.WriteString(fmt.Sprintf("🆔 <b>ID заявки:</b> <code>%s</code>\n", wlRequest.ID()))
	sb.WriteString(fmt.Sprintf("👥 <b>ID заявителя:</b> <code>%s</code>\n", wlRequest.RequesterID()))
	sb.WriteString(fmt.Sprintf("📅 <b>Создана:</b> %s\n", wlRequest.CreatedAt().Format("02.01.2006 15:04:05")))
	sb.WriteString(fmt.Sprintf("🔄 <b>Статус:</b> %s\n", wlRequest.Status()))
	return sb.String()
}

func NoPendingWLRequests() string {
	return "✅ <b>Нет ожидающих заявок</b>\n\nВсе заявки обработаны!"
}

func CallbackError(errorText string) string {
	return fmt.Sprintf("❌ <b>Ошибка:</b> %s", errorText)
}

func CallbackSuccess(successText string) string {
	return fmt.Sprintf("✅ <b>Успех:</b> %s", successText)
}

func ApprovedWLRequest(wlRequest domainWLRequest.WLRequest, arbiter domainUser.User, requester domainUser.User) string {
	var sb strings.Builder
	sb.WriteString("✅ <b>Заявка подтверждена!</b>\n\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Ник:</b> %s\n", wlRequest.Nickname()))
	sb.WriteString(fmt.Sprintf("🔗 <b>Заявитель:</b> @%s\n", requester.Username()))
	sb.WriteString(fmt.Sprintf("🔗 <b>Арбитр:</b> @%s\n", arbiter.Username()))
	sb.WriteString(fmt.Sprintf("🆔 <b>ID заявки:</b> <code>%s</code>\n", wlRequest.ID()))
	sb.WriteString(fmt.Sprintf("📅 <b>Создана:</b> %s\n", wlRequest.CreatedAt().Format("02.01.2006 15:04:05")))
	return sb.String()
}

func DeclinedWLRequest(wlRequest domainWLRequest.WLRequest, arbiter domainUser.User, requester domainUser.User) string {
	var sb strings.Builder
	sb.WriteString("❌ <b>Заявка отклонена!</b>\n\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Ник:</b> %s\n", wlRequest.Nickname()))
	sb.WriteString(fmt.Sprintf("🔄 <b>Причина отказа:</b> %s\n", wlRequest.DeclineReason()))
	sb.WriteString(fmt.Sprintf("🔗 <b>Заявитель:</b> @%s\n", requester.Username()))
	sb.WriteString(fmt.Sprintf("🔗 <b>Арбитр:</b> @%s\n", arbiter.Username()))
	sb.WriteString(fmt.Sprintf("🆔 <b>ID заявки:</b> <code>%s</code>\n", wlRequest.ID()))
	sb.WriteString(fmt.Sprintf("📅 <b>Создана:</b> %s\n", wlRequest.CreatedAt().Format("02.01.2006 15:04:05")))
	return sb.String()
}
