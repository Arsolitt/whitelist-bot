package msgs

import (
	"strings"
)

func Cancel() string {
	var sb strings.Builder
	sb.WriteString("Все действия отменены.\n\n")
	return sb.String()
}
