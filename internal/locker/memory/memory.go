package memory

import (
	"fmt"
	"sync"
	domainUser "whitelist/internal/domain/user"
)

type MemoryLocker struct {
	mu    sync.RWMutex
	locks map[domainUser.UserID]*sync.RWMutex
}

func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{
		locks: make(map[domainUser.UserID]*sync.RWMutex),
	}
}

func (l *MemoryLocker) Lock(userID domainUser.UserID) error {
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

func (l *MemoryLocker) Unlock(userID domainUser.UserID) error {
	l.mu.RLock()

	userLock, ok := l.locks[userID]

	l.mu.RUnlock()

	if !ok {
		return fmt.Errorf("user lock not found")
	}

	userLock.Unlock()
	return nil
}
