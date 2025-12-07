package msgs

import (
	"fmt"
	"strings"
	"whitelist-bot/internal/core"
)

func Start() string {
	var sb strings.Builder
	sb.WriteString("Привет! Я бот для подачи заявок в белый список.\n\n")
	sb.WriteString(fmt.Sprintf("Чтобы создать заявку, напиши: <b>%s</b>", core.CommandNewWLRequest))
	return sb.String()
}
