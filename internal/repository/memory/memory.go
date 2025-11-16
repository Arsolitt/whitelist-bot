package memory

import (
	"sync"
	"whitelist/internal/core"
	domainUser "whitelist/internal/domain/user"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	users map[domainUser.UserID]domainUser.User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: make(map[domainUser.UserID]domainUser.User),
	}
}

func (r *MemoryRepository) UserByTelegramID(telegramID int64) (domainUser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.TelegramID() == domainUser.TelegramID(telegramID) {
			return user, nil
		}
	}

	return domainUser.User{}, core.ErrUserNotFound
}

func (r *MemoryRepository) CreateUser(user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID()] = user
	return nil
}

func (r *MemoryRepository) UpdateUser(user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID()] = user
	return nil
}
