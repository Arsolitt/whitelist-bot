package main

import (
	"log"
)

func main() {
	// Загружаем конфигурацию
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Подключаемся к базе данных
	db, err := NewDatabase("whitelist.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Создаем и запускаем бота
	bot, err := NewBot(config, db)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	log.Println("Bot started successfully")
	if err := bot.Start(); err != nil {
		log.Fatal("Bot stopped with error:", err)
	}
}
