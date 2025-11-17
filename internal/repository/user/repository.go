package repository

import (
	"context"
	"whitelist/internal/domain/user"
)

type IUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (user.User, error)
	CreateUser(ctx context.Context, user user.User) error
	UpdateUser(ctx context.Context, user user.User) error
}
