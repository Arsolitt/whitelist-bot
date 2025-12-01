package repository

import (
	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
)

type PendingWLRequestWithRequester struct {
	WlRequest domainWLRequest.WLRequest
	User      domainUser.User
}
