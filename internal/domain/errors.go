package domain

import "errors"

var (
	ErrInternal = errors.New("internal error")

	ErrUserExists        = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserNonAuthorized = errors.New("user is not authorized")

	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrNotFound          = errors.New("not found")
)

var (
	ErrTokenNotFound           = errors.New("token not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrRefreshTokenNotFound    = errors.New("refresh token not found")
	ErrActivationTokenNotFound = errors.New("activation token not found")
)
