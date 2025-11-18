package msgs

import (
	"fmt"
	"strings"
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
