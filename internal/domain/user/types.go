package user

import (
	"whitelist/internal/core/utils"

	"github.com/google/uuid"
)

type (
	ID         uuid.UUID
	TelegramID int64
	FirstName  string
	LastName   string
	Username   string
)

func (t TelegramID) IsZero() bool {
	return t == 0
}

func (u Username) IsZero() bool {
	return u == ""
}

func NewID() ID {
	return ID(utils.NewUniqueID())
}

func (u ID) String() string {
	return utils.UUIDString(u)
}

func (u ID) IsZero() bool {
	return utils.UUIDIsZero(u)
}
