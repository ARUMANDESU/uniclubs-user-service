package storage

import "errors"

var (
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotExists    = errors.New("user does not exists")
	ErrSessionNotExists = errors.New("session does not exists")
	ErrSessinoExists    = errors.New("session already exists")
)
