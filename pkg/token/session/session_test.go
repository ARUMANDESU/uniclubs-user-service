package session

import (
	"crypto/rand"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockRand struct{}

// Read generates random bytes
func (r *MockRand) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(i) // Just for testing
	}
	return len(p), nil
}

func TestGenerateToken(t *testing.T) {
	// Replace rand.Reader with MockRand for predictable test results
	rand.Reader = &MockRand{}

	expectedToken := "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f" // this is hexadecimal representation of the 0..31 slice of bytes
	token, err := GenerateToken()

	if !assert.NoError(t, err) {
		assert.Equal(t, "", token, fmt.Sprintf("if error returned, should return empty string, but got: %s", token))
	}

	require.NoError(t, err, fmt.Sprintf("GenerateToken returned an error: %v", err))
	assert.Equal(t, expectedToken, token, fmt.Sprintf("GenerateToken did not return the expected token. Expected: %s, Got: %s", expectedToken, token))

}
