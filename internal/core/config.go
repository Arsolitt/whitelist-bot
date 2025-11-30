package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Logs     LogsConfig     `env-prefix:"LOGS_"`
	Sqlite   SqliteConfig   `env-prefix:"SQLITE_"`
	Postgres PostgresConfig `env-prefix:"POSTGRES_"`
	Telegram TelegramConfig `env-prefix:"TELEGRAM_"`
	Server   ServerConfig   `env-prefix:"SERVER_"`
}

type LogsConfig struct {
	LogLevel    string `env:"LEVEL"        env-default:"info"  validate:"oneof=debug info warn error"`
	IsPretty    bool   `env:"IS_PRETTY"    env-default:"true"`
	WithContext bool   `env:"WITH_CONTEXT" env-default:"true"`
	WithSources bool   `env:"WITH_SOURCES" env-default:"false"`
}

type SqliteConfig struct {
	URL string `env:"URL" env-default:"./data/whitelist.db" validate:"required"`
}

type PostgresConfig struct {
	URL string `env:"URL" env-default:"postgres://app:app@127.0.0.1:5432/app" validate:"required"`
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
