package activate

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	b := make([]byte, 16)

	// Generate cryptographically secure random bytes
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Return the encoded string in hexadecimal format
	return hex.EncodeToString(b), nil
}
