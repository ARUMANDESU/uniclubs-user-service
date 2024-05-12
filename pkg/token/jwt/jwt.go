package jwt

import (
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

var (
	ErrTokenIsNotValid         = errors.New("token is not valid")
	ErrInvalidTokenClaims      = errors.New("invalid token claims")
	ErrUserIDClaimNotFound     = errors.New("user_id claim not found or invalid")
	ErrUserIDMismatch          = errors.New("user ID from token does not match provided user ID")
	ErrTokenIsExpired          = errors.New("token is expired")
	ErrTokenSignatureIsInvalid = errors.New("token signature is invalid")
)

func GenerateTokenPair(userID int64, cfg config.JWTConfig) (map[string]string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = 1
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(cfg.AccessTokenDuration).Unix()

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString([]byte(cfg.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = 1
	rtClaims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	rt, err := refreshToken.SignedString([]byte(cfg.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  t,
		"refresh_token": rt,
	}, nil
}

func GetUserIDFromToken(tokenString string, secret string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret used to sign the token
		return []byte(secret), nil
	})
	if err != nil {
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "token signature is invalid"):
			return 0, ErrTokenSignatureIsInvalid
		case strings.Contains(errorMessage, "token is expired"):
			return 0, ErrTokenIsExpired
		default:
			return 0, err
		}
	}

	// Check if the token is valid
	if !token.Valid {
		return 0, ErrTokenIsNotValid
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidTokenClaims
	}

	// Extract user_id from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, ErrUserIDClaimNotFound
	}

	return int64(userID), nil
}
