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
	defer l.mu.Unlock()

	if _, ok := l.locks[userID]; !ok {
		l.locks[userID] = &sync.RWMutex{}
	}

	l.locks[userID].Lock()
	return nil
}

func (l *MemoryLocker) Unlock(userID domainUser.UserID) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.locks[userID]; !ok {
		return fmt.Errorf("user lock not found")
	}

	l.locks[userID].Unlock()
	return nil
}
