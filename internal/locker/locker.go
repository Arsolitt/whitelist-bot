package locker

import "whitelist/internal/model"

type ILocker interface {
	Lock(userID model.UserID) error
	Unlock(userID model.UserID) error
}
