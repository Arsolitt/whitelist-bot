package logger

import "sync"

type logData struct {
	mu   sync.RWMutex
	data map[string]any
}

type keyType string

const (
	dataKey  = keyType("logData")
	levelKey = keyType("slogLevel")
)

const (
	ChatIDField          = "chat_id"
	UserIDField          = "user_id"
	CorrelationIDField   = "correlation_id"
	RequestIDField       = "request_id"
	UserNameField        = "user_name"
	UserFirstNameField   = "user_first_name"
	UserLastNameField    = "user_last_name"
	UserTelegramIDField  = "user_telegram_id"
	UpdateIDField        = "update_id"
	MessageIDField       = "message_id"
	MessageChatIDField   = "message_chat_id"
	MessageChatTypeField = "message_chat_type"
	CurrentStateField    = "current_state"
	NextStateField       = "next_state"
	ErrorField           = "error"
)
