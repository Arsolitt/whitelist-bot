package fsm

import domainUser "whitelist/internal/domain/user"

type State string

const (
	StateStart             State = "start"
	StateIdle              State = "idle"
	StateWaitingWLNickname State = "waiting_wl_nickname"
)

type IFSM interface {
	GetState(userID domainUser.ID) (State, error)
	SetState(userID domainUser.ID, state State) error
}
