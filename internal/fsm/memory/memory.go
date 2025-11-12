package memory

import (
	"sync"
	"whitelist/internal/fsm"
	"whitelist/internal/model"
)

type MemoryFSM struct {
	states map[model.UserID]fsm.State
	mu     sync.RWMutex
}

func NewMemoryFSM() *MemoryFSM {
	return &MemoryFSM{
		states: make(map[model.UserID]fsm.State),
	}
}

func (f *MemoryFSM) GetState(userID model.UserID) (fsm.State, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	state, ok := f.states[userID]
	if !ok {
		f.mu.RUnlock()

		f.mu.Lock()
		f.states[userID] = state
		f.mu.Unlock()

		f.mu.RLock()

		return fsm.StateStart, nil
	}

	return state, nil
}

func (f *MemoryFSM) SetState(userID model.UserID, state fsm.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.states[userID] = state
	return nil
}
