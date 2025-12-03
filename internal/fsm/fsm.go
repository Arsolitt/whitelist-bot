package fsm

import domainUser "whitelist-bot/internal/domain/user"

type State string

const (
	StateStart             State = "start"
	StateIdle              State = "idle"
	StateWaitingWLNickname State = "waiting_wl_nickname"
	// StateAnketaName        State = "anketa_name"
	// StateAnketaAge         State = "anketa_age"
)

type IFSM interface {
	GetState(userID domainUser.ID) (State, error)
	SetState(userID domainUser.ID, state State) error
}
