package fsm

import "whitelist/internal/model"

type State string

const (
	StateStart           State = "start"
	StateIdle            State = "idle"
	StateWaitingNickname State = "waiting_nickname"
)

type IFSM interface {
	GetState(userID model.UserID) (State, error)
	SetState(userID model.UserID, state State) error
}
