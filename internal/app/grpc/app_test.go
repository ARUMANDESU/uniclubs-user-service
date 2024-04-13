package grpcapp

import (
	"errors"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/grpc/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockAuthService := mocks.NewAuth(t)
	mockManagementService := mocks.NewManagement(t)

	port := 55044
	app := New(logger, port, mockAuthService, mockManagementService)

	require.NotNil(t, app)
	assert.Equal(t, logger, app.log)
	assert.Equal(t, port, app.port)
	assert.NotNil(t, app.gRPCServer)
}

func TestApp_Run_Error(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app := New(logger, 0, nil, nil)

	ErrSrvStopped := errors.New("grpc: the server has been stopped")

	go func() {
		err := app.Run()
		if err == nil || !strings.Contains(err.Error(), "grpcapp.Run") || !strings.Contains(err.Error(), ErrSrvStopped.Error()) {
			t.Errorf("Unexpected error: %v", err)
		}
	}()

	app.Stop()

}

func TestApp_RunAndStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockAuthService := mocks.NewAuth(t)
	mockManagementService := mocks.NewManagement(t)

	// Create new App instance
	port := 55066
	app := New(logger, port, mockAuthService, mockManagementService)

	require.NotNil(t, app)

	go func() {
		err := app.Run()
		assert.EqualError(t, err, "grpcapp.Run: grpc: the server has been stopped")
	}()

	app.Stop()
}

func TestApp_MustRun(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockAuthService := mocks.NewAuth(t)
	mockManagementService := mocks.NewManagement(t)

	// Create new App instance
	port := 55077
	app := New(logger, port, mockAuthService, mockManagementService)

	require.NotNil(t, app)

	go func() {
		app.MustRun()
	}()
}
