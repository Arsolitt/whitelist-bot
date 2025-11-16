package repository

import "whitelist/internal/domain/user"

type IUserRepository interface {
	UserByTelegramID(telegramID int64) (user.User, error)
	CreateUser(user user.User) error
	UpdateUser(user user.User) error
}
