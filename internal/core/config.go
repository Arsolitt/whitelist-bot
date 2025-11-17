package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Logs     LogsConfig     `env-prefix:"LOGS_"`
	Database DatabaseConfig `env-prefix:"DATABASE_"`
	Telegram TelegramConfig `env-prefix:"TELEGRAM_"`
	Server   ServerConfig   `env-prefix:"SERVER_"`
}

type LogsConfig struct {
	LogLevel    string `env:"LEVEL"        env-default:"info"  validate:"oneof=debug info warn error"`
	IsPretty    bool   `env:"IS_PRETTY"    env-default:"true"`
	WithContext bool   `env:"WITH_CONTEXT" env-default:"true"`
	WithSources bool   `env:"WITH_SOURCES" env-default:"false"`
}

type DatabaseConfig struct {
	Path         string `env:"PATH"           env-default:"whitelist.db" validate:"required"`
	MaxOpenConns int    `env:"MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns int    `env:"MAX_IDLE_CONNS" env-default:"5"`
}

type TelegramConfig struct {
	Token    string  `env:"TOKEN"     validate:"required"`
	AdminIDs []int64 `env:"ADMIN_IDS" validate:"required,min=1"`
	Debug    bool    `env:"DEBUG"                               env-default:"false"`
}

type ServerConfig struct {
	MaxRequestsPerUser int `env:"MAX_REQUESTS_PER_USER" env-default:"3" validate:"min=1"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load .env file: %w", err)
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
