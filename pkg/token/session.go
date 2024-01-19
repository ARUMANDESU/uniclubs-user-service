package token

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)

	// Generate cryptographically secure random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Return the encoded string in hexadecimal format
	return hex.EncodeToString(b), nil
}
