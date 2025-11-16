package locker

import domainUser "whitelist/internal/domain/user"

type ILocker interface {
	Lock(userID domainUser.UserID) error
	Unlock(userID domainUser.UserID) error
}
