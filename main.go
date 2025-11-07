package main

import (
	"log"
	"log/slog"
	"os"
)

func main() {
	// Настраиваем структурированное логирование
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Whitelist Bot application")

	// Загружаем конфигурацию
	config, err := LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		log.Fatal("Failed to load config:", err)
	}

	slog.Info("Configuration loaded successfully",
		"admin_id", config.AdminID,
		"bot_token_length", len(config.BotToken))

	// Подключаемся к базе данных
	db, err := NewDatabase("whitelist.db")
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	slog.Info("Database initialized successfully", "database", "whitelist.db")

	// Создаем и запускаем бота
	bot, err := NewBot(config, db)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		log.Fatal("Failed to create bot:", err)
	}

	slog.Info("Bot started successfully", "polling", true)
	if err := bot.Start(); err != nil {
		slog.Error("Bot stopped with error", "error", err)
		log.Fatal("Bot stopped with error:", err)
	}
}
