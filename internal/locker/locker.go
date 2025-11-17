package locker

import domainUser "whitelist/internal/domain/user"

type ILocker interface {
	Lock(userID domainUser.ID) error
	Unlock(userID domainUser.ID) error
}
