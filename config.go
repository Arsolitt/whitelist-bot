package main

import (
	"os"
	"strconv"
)

type Config struct {
	BotToken string
	AdminID  int64
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("BOT_TOKEN environment variable is required")
	}

	adminIDStr := os.Getenv("ADMIN_ID")
	if adminIDStr == "" {
		panic("ADMIN_ID environment variable is required")
	}

	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		panic("ADMIN_ID must be a valid integer")
	}

	return &Config{
		BotToken: token,
		AdminID:  adminID,
	}, nil
}
