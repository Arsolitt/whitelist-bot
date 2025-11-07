package main

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	BotToken string
	AdminID  int64
}

func LoadConfig() (*Config, error) {
	slog.Info("Loading configuration from environment variables")

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		slog.Error("BOT_TOKEN environment variable is not set")
		panic("BOT_TOKEN environment variable is required")
	}
	slog.Debug("BOT_TOKEN found", "token_length", len(token))

	adminIDStr := os.Getenv("ADMIN_ID")
	if adminIDStr == "" {
		slog.Error("ADMIN_ID environment variable is not set")
		panic("ADMIN_ID environment variable is required")
	}
	slog.Debug("ADMIN_ID found", "value", adminIDStr)

	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		slog.Error("Failed to parse ADMIN_ID", "error", err, "value", adminIDStr)
		panic("ADMIN_ID must be a valid integer")
	}

	slog.Info("Configuration loaded successfully")

	return &Config{
		BotToken: token,
		AdminID:  adminID,
	}, nil
}
