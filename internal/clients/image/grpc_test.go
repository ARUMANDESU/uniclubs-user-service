package image

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := config.ClientsConfig{
		Image: struct {
			Address      string        `yaml:"address" env:"IMAGE_SERVICE_ADDRESS"`
			Timeout      time.Duration `yaml:"timeout" env:"IMAGE_SERVICE_TIMEOUT"`
			RetriesCount int           `yaml:"retries_count" env:"IMAGE_SERVICE_RETRIES_COUNT"`
		}{
			Address:      "example.com:50051",
			Timeout:      5 * time.Second,
			RetriesCount: 3,
		},
	}

	client, err := New(ctx, logger, cfg)

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestInterceptorLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	interceptorLogger := InterceptorLogger(logger)

	assert.NotNil(t, interceptorLogger)

}
