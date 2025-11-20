package core

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUnknownCommand = errors.New("unknown command")
	ErrInvalidLength  = errors.New("invalid length")
	ErrInvalidState   = errors.New("invalid state")
)
