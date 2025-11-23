package memory

import (
	"errors"
	"sync"
	domainUser "whitelist-bot/internal/domain/user"
)

var (
	ErrUserLockNotFound = errors.New("user lock not found")
)

type Locker struct {
	mu    sync.RWMutex
	locks map[domainUser.ID]*sync.RWMutex
}

func New() *Locker {
	return &Locker{
		locks: make(map[domainUser.ID]*sync.RWMutex),
	}
}

func (l *Locker) Lock(userID domainUser.ID) error {
	l.mu.Lock()

	userLock, ok := l.locks[userID]
	if !ok {
		l.locks[userID] = &sync.RWMutex{}
		userLock = l.locks[userID]
	}

	l.mu.Unlock()

	userLock.Lock()
	return nil
}

func (l *Locker) Unlock(userID domainUser.ID) error {
	l.mu.RLock()

	userLock, ok := l.locks[userID]

	l.mu.RUnlock()

	if !ok {
		return ErrUserLockNotFound
	}

	userLock.Unlock()
	return nil
}
