package memory

import (
	"sync"
	domainUser "whitelist/internal/domain/user"
	"whitelist/internal/fsm"
)

type MemoryFSM struct {
	states map[domainUser.UserID]fsm.State
	mu     sync.RWMutex
}

func NewMemoryFSM() *MemoryFSM {
	return &MemoryFSM{
		states: make(map[domainUser.UserID]fsm.State),
	}
}

func (f *MemoryFSM) GetState(userID domainUser.UserID) (fsm.State, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	state, ok := f.states[userID]
	if !ok {
		return fsm.StateStart, nil
	}

	return state, nil
}

func (f *MemoryFSM) SetState(userID domainUser.UserID, state fsm.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.states[userID] = state
	return nil
}
