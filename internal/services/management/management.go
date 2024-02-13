package management

import (
	"context"
	"errors"
	"fmt"
	imagev1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/filestorage"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/clients/image"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"log/slog"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotExist       = errors.New("user does not exist")
)

type Management struct {
	log         *slog.Logger
	usrStorage  UserStorage
	imageClient *image.Client
	amqp        Amqp
}

type Amqp interface {
	Publish(ctx context.Context, routingKey string, msg any) error
}

type UserStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUserByID(ctx context.Context, userID int64) error
	GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error)
}

func New(log *slog.Logger, storage UserStorage, client *image.Client, amqp Amqp) *Management {
	return &Management{
		log:         log,
		usrStorage:  storage,
		imageClient: client,
		amqp:        amqp,
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

	return user, nil

}

func (m Management) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "Management.UpdateUser"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.UpdateUser(ctx, user)
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

	msg := struct {
		ID        int64   `json:"id"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		AvatarURL *string `json:"avatar_url"`
	}{
		ID:        user.ID,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		AvatarURL: &user.AvatarURL,
	}

	err = m.amqp.Publish(ctx, "user.club.updated", msg)
	if err != nil {
		log.Error("failed to publish user updated event", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
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

	err = m.amqp.Publish(ctx, "user.club.deleted", userID)
	if err != nil {
		log.Error("failed to publish user deleted event", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
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

func (m Management) UpdateAvatar(ctx context.Context, userID int64, image []byte) (*domain.User, error) {
	const op = "Management.UpdateAvatar"
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

	res, err := m.imageClient.UploadImage(ctx, &imagev1.UploadImageRequest{Image: image, Filename: user.Barcode})
	if err != nil {
		log.Error("failed to upload avatar", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	user.AvatarURL = res.GetImageUrl()

	err = m.usrStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user not found", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to update user avatar url", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	msg := struct {
		ID        int64   `json:"id"`
		AvatarURL *string `json:"avatar_url"`
	}{
		ID:        user.ID,
		AvatarURL: &user.AvatarURL,
	}

	err = m.amqp.Publish(ctx, "user.club.updated", msg)
	if err != nil {
		log.Error("failed to publish user updated event", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
