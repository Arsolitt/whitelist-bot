package memory

import (
	"sync"
	"whitelist/internal/core"
	"whitelist/internal/model"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	users map[model.UserID]model.User
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users: make(map[model.UserID]model.User),
	}
}

func (r *MemoryRepository) UserByTelegramID(telegramID int64) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.TelegramID == telegramID {
			return user, nil
		}
	}

	return model.User{}, core.ErrUserNotFound
}

func (r *MemoryRepository) CreateUser(user model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

func (r *MemoryRepository) UpdateUser(user model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}
