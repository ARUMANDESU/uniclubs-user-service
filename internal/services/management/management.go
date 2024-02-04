package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"log/slog"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotExist       = errors.New("user does not exist")
)

type Management struct {
	log        *slog.Logger
	usrStorage UserStorage
}

type UserStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user *models.User, err error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUserByID(ctx context.Context, userID int64) error
	GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error)
}

func New(log *slog.Logger, storage UserStorage) *Management {
	return &Management{
		log:        log,
		usrStorage: storage,
	}
}

func (m Management) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	const op = "Management.GetUser"
	log := m.log.With(slog.String("op", op))

	user, err := m.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

	}

	return domain.ModelUserToDomainUser(*user), nil

}

func (m Management) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "Management.UpdateUser"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.UpdateUser(ctx, *domain.UserToModelUser(*user))
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user not found", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to update user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (m Management) DeleteUser(ctx context.Context, userID int64) error {
	const op = "Management.DeleteUser"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.DeleteUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user not found", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to delete user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (m Management) SearchUsers(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error) {
	const op = "Management.SearchUsers"
	log := m.log.With(slog.String("op", op))

	users, metadata, err := m.usrStorage.GetAll(ctx, query, filters)
	if err != nil {
		log.Error("failed to get users", logger.Err(err))
		return nil, domain.Metadata{}, fmt.Errorf("%s: %w", op, err)
	}

	return users, metadata, nil

}
