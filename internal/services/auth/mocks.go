package auth

import (
	"context"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/stretchr/testify/mock"
	"io"
	"log/slog"
	"testing"
	"time"
)

type Suite struct {
	Auth *Auth
	*MockUserStorage
	*MockTokenStorage
	*MockAmqp
}

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) SaveUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorage) GetUserByEmail(ctx context.Context, email string) (user *domain.User, err error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorage) GetUserRoleByID(ctx context.Context, userID int64) (role string, err error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockUserStorage) ActivateUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockTokenStorage struct {
	mock.Mock
}

func (m *MockTokenStorage) Create(ctx context.Context, token string, userID int64, duration time.Duration) error {
	args := m.Called(ctx, token, userID, duration)
	return args.Error(0)
}

func (m *MockTokenStorage) Get(ctx context.Context, token string) (userID int64, err error) {
	args := m.Called(ctx, token)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenStorage) Delete(ctx context.Context, sessionToken string) error {
	args := m.Called(ctx, sessionToken)
	return args.Error(0)
}

type MockAmqp struct {
	mock.Mock
}

func (m *MockAmqp) Publish(ctx context.Context, exchangeName rabbitmq.ExchangeName, routingKey rabbitmq.RoutingKey, msg any) error {
	args := m.Called(ctx, exchangeName, routingKey, msg)
	return args.Error(0)
}

func Setup(t *testing.T) *Suite {
	t.Helper()
	mockUserStorage := new(MockUserStorage)
	mockTokenStorage := new(MockTokenStorage)
	mockAmqp := new(MockAmqp)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	jwtCfg := config.JWTConfig{
		AccessTokenDuration: time.Minute,
		AccessTokenSecret:   "secret",
		RefreshTokenSecret:  "secret",
	}

	return &Suite{
		Auth:             New(logger, jwtCfg, mockUserStorage, mockTokenStorage, mockTokenStorage, mockAmqp),
		MockUserStorage:  mockUserStorage,
		MockTokenStorage: mockTokenStorage,
		MockAmqp:         mockAmqp,
	}

}
