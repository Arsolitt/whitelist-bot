package user

import "github.com/google/uuid"

type UserID uuid.UUID

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

func (u UserID) IsZero() bool {
	if u == UserID(uuid.Nil) || u.String() == "" {
		return true
	}
	return false
}

func NewUserID() UserID {
	newID, err := uuid.NewV7()
	if err != nil {
		newID = uuid.New()
	}
	return UserID(newID)
}
