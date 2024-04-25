package storage

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotExists  = errors.New("user does not exists")
	ErrTokenNotExists = errors.New("refresh token does not exists")
)
