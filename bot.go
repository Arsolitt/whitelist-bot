package main

import (
	"fmt"
	"log"
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
	api, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:       api,
		db:        db,
		config:    config,
		userState: make(map[int64]string),
	}, nil
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID

	// Обработка команд
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(message)
		case "apply":
			b.handleApplyCommand(message)
		case "status":
			b.handleStatusCommand(message)
		case "pending":
			if userID == b.config.AdminID {
				b.handlePendingCommand(message)
			}
		case "cancel":
			delete(b.userState, userID)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Действие отменено.")
			b.api.Send(msg)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Используйте /start для просмотра доступных команд.")
			b.api.Send(msg)
		}
		return
	}

	// Обработка текстовых сообщений в зависимости от состояния
	state, exists := b.userState[userID]
	if exists && state == "waiting_nickname" {
		b.handleNicknameInput(message)
		return
	}

	// Если нет активного состояния
	msg := tgbotapi.NewMessage(message.Chat.ID, "Используйте /start для просмотра доступных команд.")
	b.api.Send(msg)
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	userID := message.From.ID
	var text string

	if userID == b.config.AdminID {
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
}

func (b *Bot) handleApplyCommand(message *tgbotapi.Message) {
	userID := message.From.ID

	// Проверяем, есть ли уже активная заявка
	lastRequest, err := b.db.GetUserLastRequest(userID)
	if err != nil {
		log.Printf("Error getting user last request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if lastRequest != nil && lastRequest.Status == StatusPending {
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"⏳ У вас уже есть активная заявка на рассмотрении. Дождитесь решения администратора.\n\nИспользуйте /status для проверки статуса.")
		b.api.Send(msg)
		return
	}

	b.userState[userID] = "waiting_nickname"
	msg := tgbotapi.NewMessage(message.Chat.ID, "📝 Отправьте свой никнейм для заявки на вайтлист.\n\nИспользуйте /cancel для отмены.")
	b.api.Send(msg)
}

func (b *Bot) handleNicknameInput(message *tgbotapi.Message) {
	userID := message.From.ID
	nickname := strings.TrimSpace(message.Text)

	if nickname == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Никнейм не может быть пустым. Попробуйте еще раз.")
		b.api.Send(msg)
		return
	}

	if len(nickname) > 100 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Никнейм слишком длинный. Максимум 100 символов.")
		b.api.Send(msg)
		return
	}

	username := message.From.UserName
	err := b.db.CreateRequest(userID, username, nickname)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка при создании заявки. Попробуйте позже.")
		b.api.Send(msg)
		delete(b.userState, userID)
		return
	}

	delete(b.userState, userID)

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
}

func (b *Bot) handleStatusCommand(message *tgbotapi.Message) {
	userID := message.From.ID

	lastRequest, err := b.db.GetUserLastRequest(userID)
	if err != nil {
		log.Printf("Error getting user last request: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if lastRequest == nil {
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

	text := fmt.Sprintf("%s Статус вашей заявки: %s\n\n"+
		"Никнейм: %s\n"+
		"Дата подачи: %s",
		statusEmoji, statusText, lastRequest.Nickname, lastRequest.CreatedAt.Format("02.01.2006 15:04"))

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	b.api.Send(msg)
}

func (b *Bot) handlePendingCommand(message *tgbotapi.Message) {
	requests, err := b.db.GetPendingRequests()
	if err != nil {
		log.Printf("Error getting pending requests: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
		b.api.Send(msg)
		return
	}

	if len(requests) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "ℹ️ Нет ожидающих заявок.")
		b.api.Send(msg)
		return
	}

	for _, req := range requests {
		b.sendRequestToAdmin(message.Chat.ID, &req)
	}
}

func (b *Bot) sendRequestToAdmin(chatID int64, req *WhitelistRequest) {
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
}

func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Проверяем, что это админ
	if callback.From.ID != b.config.AdminID {
		answer := tgbotapi.NewCallback(callback.ID, "У вас нет прав для выполнения этого действия.")
		b.api.Send(answer)
		return
	}

	parts := strings.Split(callback.Data, "_")
	if len(parts) != 2 {
		answer := tgbotapi.NewCallback(callback.ID, "Неверный формат данных.")
		b.api.Send(answer)
		return
	}

	action := parts[0]
	requestIDStr := parts[1]
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		answer := tgbotapi.NewCallback(callback.ID, "Неверный ID заявки.")
		b.api.Send(answer)
		return
	}

	// Получаем заявку
	request, err := b.db.GetRequestByID(requestID)
	if err != nil {
		log.Printf("Error getting request: %v", err)
		answer := tgbotapi.NewCallback(callback.ID, "Ошибка при получении заявки.")
		b.api.Send(answer)
		return
	}

	if request.Status != StatusPending {
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
		answer := tgbotapi.NewCallback(callback.ID, "Неизвестное действие.")
		b.api.Send(answer)
		return
	}

	// Обновляем статус в БД
	err = b.db.UpdateRequestStatus(requestID, newStatus)
	if err != nil {
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

	// Уведомляем пользователя
	userMsg := tgbotapi.NewMessage(request.UserID, userMessage)
	b.api.Send(userMsg)
}
