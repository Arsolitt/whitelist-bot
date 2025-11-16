package fsm

import domainUser "whitelist/internal/domain/user"

type State string

const (
	StateStart           State = "start"
	StateIdle            State = "idle"
	StateWaitingNickname State = "waiting_nickname"
)

type IFSM interface {
	GetState(userID domainUser.UserID) (State, error)
	SetState(userID domainUser.UserID, state State) error
}
