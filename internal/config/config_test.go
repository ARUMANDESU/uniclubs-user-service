package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLoadByPath_HappyPath(t *testing.T) {
	cfg, err := LoadByPath("../../config/test.yaml")

	require.NoError(t, err, "no error must be returned")
	assert.Equal(t, cfg.Env, "local")
	assert.Equal(t, cfg.DatabaseDSN, "postgresql://user:password@localhost:5432/dbname")
	assert.Equal(t, cfg.RedisURL, "redis://:@localhost:port")
	assert.Equal(t, cfg.GRPC.Port, 5000)
	assert.Equal(t, cfg.GRPC.Timeout, time.Hour)
}

func TestLoadByPath_FailPath_WrongPath(t *testing.T) {
	_, err := LoadByPath("./some/wrong/path")

	require.Error(t, err, "have to return error")
}
