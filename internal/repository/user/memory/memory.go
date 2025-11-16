package memory

import (
	"sync"
	"whitelist/internal/core"
	domainUser "whitelist/internal/domain/user"
)

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[domainUser.UserID]domainUser.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[domainUser.UserID]domainUser.User),
	}
}

func (r *MemoryUserRepository) UserByTelegramID(telegramID int64) (domainUser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.TelegramID() == domainUser.TelegramID(telegramID) {
			return user, nil
		}
	}

	return domainUser.User{}, core.ErrUserNotFound
}

func (r *MemoryUserRepository) CreateUser(user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID()] = user
	return nil
}

func (r *MemoryUserRepository) UpdateUser(user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID()] = user
	return nil
}
