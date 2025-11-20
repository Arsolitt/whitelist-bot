package core

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUnknownCommand = errors.New("unknown command")
)
