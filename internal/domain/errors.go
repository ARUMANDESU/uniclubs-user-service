package domain

import "errors"

var (
	ErrTokenIsNotValid         = errors.New("token is not valid")
	ErrInvalidTokenClaims      = errors.New("invalid token claims")
	ErrUserIDClaimNotFound     = errors.New("user_id claim not found or invalid")
	ErrUserIDMismatch          = errors.New("user ID from token does not match provided user ID")
	ErrTokenIsExpired          = errors.New("token is expired")
	ErrTokenSignatureIsInvalid = errors.New("token signature is invalid")

	ErrImageQuality = errors.New("image quality must be between 1 and 100")
	ErrImageFormat  = errors.New("unsupported image format")
	ErrImageIsEmpty = errors.New("image buffer is empty")
)
