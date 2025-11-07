package main

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	db        *Database
	config    *Config
	userState map[int64]string
}

func NewBot(config *Config, db *Database) (*Bot, error) {
	slog.Info("Initializing Telegram bot")

	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		slog.Error("Failed to create bot API", "error", err)
		return nil, err
	}

	slog.Info("Bot authorized successfully",
		"username", api.Self.UserName,
		"bot_id", api.Self.ID)
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:       api,
		db:        db,
		config:    config,
		userState: make(map[int64]string),
	}, nil
}

func (b *Bot) Start() error {
	slog.Info("Starting bot polling", "timeout", 60)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	slog.Info("Bot polling started, listening for updates")

	for update := range updates {
		if update.Message != nil {
			slog.Debug("Received message update",
				"update_id", update.UpdateID,
				"user_id", update.Message.From.ID,
				"username", update.Message.From.UserName)
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			slog.Debug("Received callback query",
				"update_id", update.UpdateID,
				"user_id", update.CallbackQuery.From.ID,
				"data", update.CallbackQuery.Data)
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName

	isAdmin := userID == b.config.AdminID

	// Обработка команд
	if message.IsCommand() {
		cmd := message.Command()
		slog.Info("User sent command",
			"command", cmd,
			"user_id", userID,
			"username", username,
			"is_admin", isAdmin,
			"chat_id", message.Chat.ID)

		switch cmd {
		case "start":
			b.handleStart(message)
		case "apply":
			b.handleApplyCommand(message)
		case "status":
			b.handleStatusCommand(message)
		case "pending":
			if isAdmin {
				b.handlePendingCommand(message)
			} else {
				slog.Warn("Non-admin user tried to access admin command",
					"user_id", userID,
					"username", username,
					"command", "pending")
			}
		case "cancel":
			slog.Info("User cancelled action",
				"user_id", userID,
				"username", username,
				"previous_state", b.userState[userID])
			delete(b.userState, userID)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Действие отменено.")
			b.api.Send(msg)
		default:
			slog.Info("User sent unknown command",
				"command", cmd,
				"user_id", userID,
				"username", username)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используйте /start для просмотра доступных команд.")
			b.api.Send(msg)
		}
		return
	}

	// Обработка текстовых сообщений в зависимости от состояния
	state, exists := b.userState[userID]
	if exists && state == "waiting_nickname" {
		slog.Info("Processing nickname input",
			"user_id", userID,
			"username", username)
		b.handleNicknameInput(message)
		return
	}

	// Если нет активного состояния
	slog.Debug("User sent message without active state",
		"user_id", userID,
		"username", username,
		"text_length", len(message.Text))
	msg := tgbotapi.NewMessage(message.Chat.ID, "Используйте /start для просмотра доступных команд.")
	b.api.Send(msg)
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName
	isAdmin := userID == b.config.AdminID

	slog.Info("Handling /start command",
		"user_id", userID,
		"username", username,
		"is_admin", isAdmin)

	var text string

	if isAdmin {
		text = `👋 Добро пожаловать, администратор!

Доступные команды:
/apply - Подать заявку на вайтлист
/status - Проверить статус заявки
/pending - Просмотреть все ожидающие заявки

Как администратор, вы можете одобрять или отклонять заявки пользователей.`
	} else {
		text = `👋 Добро пожаловать в бота для управления вайтлистом!

Доступные команды:
/apply - Подать заявку на вайтлист
/status - Проверить статус вашей заявки
/cancel - Отменить текущее действие`
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)

	slog.Info("Welcome message sent",
		"user_id", userID,
		"is_admin", isAdmin)
}

func (b *Bot) handleApplyCommand(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName

	slog.Info("User initiated whitelist application",
		"user_id", userID,
		"username", username)

	// Проверяем, есть ли уже активная заявка
	lastRequest, err := b.db.GetUserLastRequest(userID)
	if err != nil {
		slog.Error("Error getting user last request during apply",
			"error", err,
			"user_id", userID)
		log.Printf("Error getting user last request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if lastRequest != nil && lastRequest.Status == StatusPending {
		slog.Info("User already has pending request",
			"user_id", userID,
			"request_id", lastRequest.ID)
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"⏳ У вас уже есть активная заявка на рассмотрении. Дождитесь решения администратора.\n\nИспользуйте /status для проверки статуса.")
		b.api.Send(msg)
		return
	}

	b.userState[userID] = "waiting_nickname"
	slog.Info("User state set to waiting_nickname",
		"user_id", userID,
		"username", username)

	msg := tgbotapi.NewMessage(message.Chat.ID, "📝 Отправьте свой никнейм для заявки на вайтлист.\n\nИспользуйте /cancel для отмены.")
	b.api.Send(msg)
}

func (b *Bot) handleNicknameInput(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName
	nickname := strings.TrimSpace(message.Text)

	slog.Info("Processing nickname input",
		"user_id", userID,
		"username", username,
		"nickname", nickname,
		"nickname_length", len(nickname))

	if nickname == "" {
		slog.Warn("User submitted empty nickname",
			"user_id", userID,
			"username", username)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Никнейм не может быть пустым. Попробуйте еще раз.")
		b.api.Send(msg)
		return
	}

	if len(nickname) > 100 {
		slog.Warn("User submitted too long nickname",
			"user_id", userID,
			"username", username,
			"nickname_length", len(nickname))
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Никнейм слишком длинный. Максимум 100 символов.")
		b.api.Send(msg)
		return
	}

	err := b.db.CreateRequest(userID, username, nickname)
	if err != nil {
		slog.Error("Failed to create whitelist request",
			"error", err,
			"user_id", userID,
			"nickname", nickname)
		log.Printf("Error creating request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при создании заявки. Попробуйте позже.")
		b.api.Send(msg)
		delete(b.userState, userID)
		return
	}

	delete(b.userState, userID)
	slog.Info("Whitelist request submitted successfully",
		"user_id", userID,
		"username", username,
		"nickname", nickname)

	// Уведомляем пользователя
	msg := tgbotapi.NewMessage(message.Chat.ID,
		"✅ Ваша заявка успешно отправлена!\n\n"+
			"Никнейм: "+nickname+"\n\n"+
			"Администратор рассмотрит вашу заявку в ближайшее время. Используйте /status для проверки статуса.")
	b.api.Send(msg)

	// Уведомляем админа
	b.notifyAdminNewRequest(userID, username, nickname)
}

func (b *Bot) notifyAdminNewRequest(userID int64, username, nickname string) {
	slog.Info("Notifying admin about new request",
		"admin_id", b.config.AdminID,
		"user_id", userID,
		"username", username,
		"nickname", nickname)

	userInfo := fmt.Sprintf("ID: %d", userID)
	if username != "" {
		userInfo += fmt.Sprintf("\nUsername: @%s", username)
	}

	text := fmt.Sprintf("🔔 Новая заявка на вайтлист!\n\n"+
		"Пользователь:\n%s\n\n"+
		"Никнейм: %s\n\n"+
		"Используйте /pending для просмотра всех ожидающих заявок.", userInfo, nickname)

	msg := tgbotapi.NewMessage(b.config.AdminID, text)
	b.api.Send(msg)

	slog.Info("Admin notification sent",
		"admin_id", b.config.AdminID,
		"user_id", userID)
}

func (b *Bot) handleStatusCommand(message *tgbotapi.Message) {
	userID := message.From.ID
	username := message.From.UserName

	slog.Info("User checking request status",
		"user_id", userID,
		"username", username)

	lastRequest, err := b.db.GetUserLastRequest(userID)
	if err != nil {
		slog.Error("Error getting user last request for status check",
			"error", err,
			"user_id", userID)
		log.Printf("Error getting user last request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if lastRequest == nil {
		slog.Info("User has no requests",
			"user_id", userID,
			"username", username)
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"ℹ️ У вас нет заявок.\n\nИспользуйте /apply для подачи заявки на вайтлист.")
		b.api.Send(msg)
		return
	}

	var statusText string
	var statusEmoji string
	switch lastRequest.Status {
	case StatusPending:
		statusEmoji = "⏳"
		statusText = "На рассмотрении"
	case StatusApproved:
		statusEmoji = "✅"
		statusText = "Одобрена"
	case StatusRejected:
		statusEmoji = "❌"
		statusText = "Отклонена"
	}

	slog.Info("User status checked",
		"user_id", userID,
		"username", username,
		"request_id", lastRequest.ID,
		"status", lastRequest.Status,
		"nickname", lastRequest.Nickname)

	text := fmt.Sprintf("%s Статус вашей заявки: %s\n\n"+
		"Никнейм: %s\n"+
		"Дата подачи: %s",
		statusEmoji, statusText, lastRequest.Nickname, lastRequest.CreatedAt.Format("02.01.2006 15:04"))

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

func (b *Bot) handlePendingCommand(message *tgbotapi.Message) {
	adminID := message.From.ID

	slog.Info("Admin requested pending requests list",
		"admin_id", adminID)

	requests, err := b.db.GetPendingRequests()
	if err != nil {
		slog.Error("Error getting pending requests for admin",
			"error", err,
			"admin_id", adminID)
		log.Printf("Error getting pending requests: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if len(requests) == 0 {
		slog.Info("No pending requests found",
			"admin_id", adminID)
		msg := tgbotapi.NewMessage(message.Chat.ID, "ℹ️ Нет ожидающих заявок.")
		b.api.Send(msg)
		return
	}

	slog.Info("Sending pending requests to admin",
		"admin_id", adminID,
		"requests_count", len(requests))

	for _, req := range requests {
		b.sendRequestToAdmin(message.Chat.ID, &req)
	}
}

func (b *Bot) sendRequestToAdmin(chatID int64, req *WhitelistRequest) {
	slog.Info("Sending request details to admin",
		"admin_chat_id", chatID,
		"request_id", req.ID,
		"user_id", req.UserID,
		"nickname", req.Nickname)

	userInfo := fmt.Sprintf("ID: %d", req.UserID)
	if req.Username != "" {
		userInfo += fmt.Sprintf("\nUsername: @%s", req.Username)
	}

	text := fmt.Sprintf("📋 Заявка #%d\n\n"+
		"Пользователь:\n%s\n\n"+
		"Никнейм: %s\n"+
		"Дата: %s",
		req.ID, userInfo, req.Nickname, req.CreatedAt.Format("02.01.2006 15:04"))

	// Создаем клавиатуру с кнопками
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Одобрить", fmt.Sprintf("approve_%d", req.ID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_%d", req.ID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)

	slog.Debug("Request card sent to admin",
		"request_id", req.ID,
		"admin_chat_id", chatID)
}

func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	adminID := callback.From.ID
	adminUsername := callback.From.UserName

	slog.Info("Received callback query",
		"callback_id", callback.ID,
		"user_id", adminID,
		"username", adminUsername,
		"data", callback.Data)

	// Проверяем, что это админ
	if adminID != b.config.AdminID {
		slog.Warn("Non-admin user tried to use callback query",
			"user_id", adminID,
			"username", adminUsername,
			"data", callback.Data)
		answer := tgbotapi.NewCallback(callback.ID, "У вас нет прав для выполнения этого действия.")
		b.api.Send(answer)
		return
	}

	parts := strings.Split(callback.Data, "_")
	if len(parts) != 2 {
		slog.Error("Invalid callback data format",
			"admin_id", adminID,
			"data", callback.Data)
		answer := tgbotapi.NewCallback(callback.ID, "Неверный формат данных.")
		b.api.Send(answer)
		return
	}

	action := parts[0]
	requestIDStr := parts[1]
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		slog.Error("Failed to parse request ID from callback",
			"error", err,
			"admin_id", adminID,
			"request_id_str", requestIDStr)
		answer := tgbotapi.NewCallback(callback.ID, "Неверный ID заявки.")
		b.api.Send(answer)
		return
	}

	slog.Info("Admin processing request",
		"admin_id", adminID,
		"admin_username", adminUsername,
		"action", action,
		"request_id", requestID)

	// Получаем заявку
	request, err := b.db.GetRequestByID(requestID)
	if err != nil {
		slog.Error("Error getting request for callback processing",
			"error", err,
			"admin_id", adminID,
			"request_id", requestID)
		log.Printf("Error getting request: %v", err)
		answer := tgbotapi.NewCallback(callback.ID, "Ошибка при получении заявки.")
		b.api.Send(answer)
		return
	}

	if request.Status != StatusPending {
		slog.Warn("Admin tried to process already processed request",
			"admin_id", adminID,
			"request_id", requestID,
			"current_status", request.Status)
		answer := tgbotapi.NewCallback(callback.ID, "Эта заявка уже обработана.")
		b.api.Send(answer)
		return
	}

	var newStatus RequestStatus
	var statusText string
	var userMessage string

	switch action {
	case "approve":
		newStatus = StatusApproved
		statusText = "✅ Одобрена"
		userMessage = "🎉 Поздравляем! Ваша заявка на вайтлист была одобрена!\n\nНикнейм: " + request.Nickname
	case "reject":
		newStatus = StatusRejected
		statusText = "❌ Отклонена"
		userMessage = "😔 К сожалению, ваша заявка на вайтлист была отклонена.\n\nНикнейм: " + request.Nickname
	default:
		slog.Warn("Unknown action in callback",
			"admin_id", adminID,
			"action", action,
			"request_id", requestID)
		answer := tgbotapi.NewCallback(callback.ID, "Неизвестное действие.")
		b.api.Send(answer)
		return
	}

	slog.Info("Admin decision made",
		"admin_id", adminID,
		"admin_username", adminUsername,
		"request_id", requestID,
		"user_id", request.UserID,
		"user_username", request.Username,
		"nickname", request.Nickname,
		"decision", action,
		"new_status", newStatus)

	// Обновляем статус в БД
	err = b.db.UpdateRequestStatus(requestID, newStatus)
	if err != nil {
		slog.Error("Failed to update request status",
			"error", err,
			"admin_id", adminID,
			"request_id", requestID,
			"new_status", newStatus)
		log.Printf("Error updating request status: %v", err)
		answer := tgbotapi.NewCallback(callback.ID, "Ошибка при обновлении статуса.")
		b.api.Send(answer)
		return
	}

	// Уведомляем админа
	answer := tgbotapi.NewCallback(callback.ID, fmt.Sprintf("Заявка #%d %s", requestID, statusText))
	b.api.Send(answer)

	// Обновляем сообщение админа
	userInfo := fmt.Sprintf("ID: %d", request.UserID)
	if request.Username != "" {
		userInfo += fmt.Sprintf("\nUsername: @%s", request.Username)
	}

	editText := fmt.Sprintf("📋 Заявка #%d - %s\n\n"+
		"Пользователь:\n%s\n\n"+
		"Никнейм: %s\n"+
		"Дата: %s",
		request.ID, statusText, userInfo, request.Nickname, request.CreatedAt.Format("02.01.2006 15:04"))

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, editText)
	b.api.Send(edit)

	slog.Info("Notifying user about decision",
		"request_id", requestID,
		"user_id", request.UserID,
		"decision", newStatus)

	// Уведомляем пользователя
	userMsg := tgbotapi.NewMessage(request.UserID, userMessage)
	b.api.Send(userMsg)

	slog.Info("Request processing completed",
		"admin_id", adminID,
		"request_id", requestID,
		"user_id", request.UserID,
		"final_status", newStatus)
}
