package memory

import (
	"context"
	"log/slog"
	"sync"
	"whitelist-bot/internal/core"
	domainUser "whitelist-bot/internal/domain/user"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[domainUser.ID]domainUser.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[domainUser.ID]domainUser.User),
	}
}

func (r *UserRepository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slog.DebugContext(ctx, "Getting user by telegram ID")

	for _, user := range r.users {
		if user.TelegramID() == domainUser.TelegramID(telegramID) {
			return user, nil
		}
	}

	return domainUser.User{}, core.ErrUserNotFound
}

func (r *UserRepository) CreateUser(ctx context.Context, user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.DebugContext(ctx, "Creating user")

	r.users[user.ID()] = user
	return nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.DebugContext(ctx, "Updating user")

	r.users[user.ID()] = user
	return nil
}
