package memory

import (
	"sync"
	domainUser "whitelist-bot/internal/domain/user"
	"whitelist-bot/internal/fsm"
)

type FSM struct {
	states map[domainUser.ID]fsm.State
	mu     sync.RWMutex
}

func NewFSM() *FSM {
	return &FSM{
		states: make(map[domainUser.ID]fsm.State),
	}
}

func (f *FSM) GetState(userID domainUser.ID) (fsm.State, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	state, ok := f.states[userID]
	if !ok {
		return fsm.StateIdle, nil
	}

	return state, nil
}

func (f *FSM) SetState(userID domainUser.ID, state fsm.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.states[userID] = state
	return nil
}
