package user

import "github.com/google/uuid"

type ID uuid.UUID

func NewID() ID {
	newID, err := uuid.NewV7()
	if err != nil {
		newID = uuid.New()
	}
	return ID(newID)
}

func (u ID) String() string {
	return uuid.UUID(u).String()
}

func (u ID) IsZero() bool {
	if u == ID(uuid.Nil) || u.String() == "" {
		return true
	}
	return false
}
