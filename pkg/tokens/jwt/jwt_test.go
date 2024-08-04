package jwt

import (
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGenerateTokenPair_HappyPath(t *testing.T) {

	userIDs := []int64{1, 2, 3, 4, 5, -1, 1000000203, 123421453453}
	cfg := config.JWTConfig{
		AccessTokenDuration: 0,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}
	for _, userID := range userIDs {
		tokenPair, err := GenerateTokenPair(userID, cfg)

		require.NoError(t, err, "GenerateTokenPair should not return an error")

		assert.NotNil(t, tokenPair["access_token"], "access_token should not be nil")
		assert.NotNil(t, tokenPair["refresh_token"], "refresh_token should not be nil")
	}
}

func TestGetUserIDFromToken_HappyPath(t *testing.T) {
	userIDs := []int64{1, 2, 3, 4, 5, -1, 1000000203, 123421453453}
	cfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}
	for _, userID := range userIDs {
		tokenPair, err := GenerateTokenPair(userID, cfg)

		require.NoError(t, err, "GenerateTokenPair should not return an error")

		assert.NotNil(t, tokenPair["access_token"], "access_token should not be nil")
		assert.NotNil(t, tokenPair["refresh_token"], "refresh_token should not be nil")

		userIDFromToken, err := GetUserIDFromToken(tokenPair["access_token"], cfg.AccessTokenSecret)
		require.NoError(t, err, "GetUserIDFromToken should not return an error")
		assert.Equal(t, userID, userIDFromToken, "userID should be equal to userIDFromToken")
	}
}

func TestGetUserIDFromToken_ExpiredToken(t *testing.T) {
	userIDs := []int64{1, 2, 3, 4, 5, -1, 1000000203, 123421453453}
	cfg := config.JWTConfig{
		AccessTokenDuration: 0,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}

	for _, userID := range userIDs {
		t.Run(fmt.Sprintf("userID: %d", userID), func(t *testing.T) {
			tokenPair, err := GenerateTokenPair(userID, cfg)

			require.NoError(t, err, "GenerateTokenPair should not return an error")

			assert.NotNil(t, tokenPair["access_token"], "access_token should not be nil")
			assert.NotNil(t, tokenPair["refresh_token"], "refresh_token should not be nil")

			time.Sleep(time.Nanosecond)

			userIDFromToken, err := GetUserIDFromToken(tokenPair["access_token"], cfg.AccessTokenSecret)
			require.Error(t, err, "GetUserIDFromToken should return an error")
			assert.Equal(t, ErrTokenIsExpired, err, "Error should be domain.ErrTokenIsExpired")
			assert.Equal(t, int64(0), userIDFromToken, "userIDFromToken should be 0")
		})
	}

}

func TestGetUserIDFromToken_TokenIsInvalid(t *testing.T) {
	userIDs := []int64{1, 2, 3, 4, 5, -1, 1000000203, 123421453453}
	cfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}
	invalidTokenCfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "invalid",
		RefreshTokenSecret:  "invalid",
	}

	for _, userID := range userIDs {
		t.Run(fmt.Sprintf("userID: %d", userID), func(t *testing.T) {
			tokenPair, err := GenerateTokenPair(userID, cfg)

			require.NoError(t, err, "GenerateTokenPair should not return an error")

			assert.NotNil(t, tokenPair["access_token"], "access_token should not be nil")
			assert.NotNil(t, tokenPair["refresh_token"], "refresh_token should not be nil")

			userIDFromToken, err := GetUserIDFromToken(tokenPair["access_token"], invalidTokenCfg.AccessTokenSecret)
			require.Error(t, err, "GetUserIDFromToken should return an error")
			assert.ErrorIs(t, ErrTokenSignatureIsInvalid, err, "Error should be domain.ErrTokenSignatureIsInvalid")
			assert.Equal(t, int64(0), userIDFromToken, "userIDFromToken should be 0")
		})
	}
}

func TestGetUserIDFromToken_InvalidToken(t *testing.T) {
	tokenString := "invalidToken"
	secret := "secret"

	_, err := GetUserIDFromToken(tokenString, secret)

	require.Error(t, err, "GetUserIDFromToken should return an error for invalid token")
}

func BenchmarkGenerateTokenPair(b *testing.B) {
	cfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}
	for i := 0; i < b.N; i++ {
		GenerateTokenPair(1, cfg)
	}
}

func BenchmarkGetUserIDFromToken(b *testing.B) {
	cfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "",
		RefreshTokenSecret:  "",
	}
	tokenPair, err := GenerateTokenPair(1, cfg)
	require.NoError(b, err, "GenerateTokenPair should not return an error")

	for i := 0; i < b.N; i++ {
		GetUserIDFromToken(tokenPair["access_token"], cfg.AccessTokenSecret)
	}
}
