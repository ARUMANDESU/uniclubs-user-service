package token

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	_, err := GenerateSessionToken()
	require.NoError(t, err, "must not return error")

}
