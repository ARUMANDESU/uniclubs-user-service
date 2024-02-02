package session

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	_, err := GenerateToken()
	require.NoError(t, err, "must not return error")

}
