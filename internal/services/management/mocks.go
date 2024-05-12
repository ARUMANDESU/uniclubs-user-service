package management

import (
	"context"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/stretchr/testify/mock"
	"io"
	"log/slog"
	"testing"
)

type Suite struct {
	Management *Management
	*MockUserStorage
	*MockAmqp
}

func Setup(t *testing.T) *Suite {
	t.Helper()
	mockUserStorage := new(MockUserStorage)
	mockAmqp := new(MockAmqp)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &Suite{
		Management:      New(logger, mockUserStorage, mockAmqp),
		MockUserStorage: mockUserStorage,
		MockAmqp:        mockAmqp,
	}
}

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserStorage) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserStorage) UpdateUserRole(ctx context.Context, userID int64, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func (m *MockUserStorage) DeleteUserByID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserStorage) GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error) {
	args := m.Called(ctx, query, filters)
	return args.Get(0).([]*domain.User), args.Get(1).(domain.Metadata), args.Error(2)
}

type MockAmqp struct {
	mock.Mock
}

func (m *MockAmqp) Publish(ctx context.Context, exchangeName string, routingKey string, msg interface{}) error {
	args := m.Called(ctx, exchangeName, routingKey, msg)
	return args.Error(0)

}
