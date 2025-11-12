package repository

import (
	"whitelist/internal/model"
)

type IRepository interface {
	UserByTelegramID(telegramID int64) (model.User, error)
	CreateUser(user model.User) error
	UpdateUser(user model.User) error
}
