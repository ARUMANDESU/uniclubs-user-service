package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	imageUtils "github.com/ARUMANDESU/uniclubs-user-service/pkg/image"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"log/slog"
	"path"
	"time"
)

var (
	ErrUserNotExist      = errors.New("user does not exist")
	ErrUserNonAuthorized = errors.New("user is not authorized")
)

type Management struct {
	log          *slog.Logger
	usrStorage   UserStorage
	imageStorage ImageStorage
	amqp         Amqp
}

type Amqp interface {
	Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error
}

type ImageStorage interface {
	UploadImage(ctx context.Context, image []byte, filename string) (string, error)
	DeleteImage(ctx context.Context, filename string) error
}

type UserStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateUserRole(ctx context.Context, userID int64, role string) error
	DeleteUserByID(ctx context.Context, userID int64) error
	GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error)
}

func New(log *slog.Logger, storage UserStorage, imageStorage ImageStorage, amqp Amqp) *Management {
	return &Management{
		log:          log,
		usrStorage:   storage,
		imageStorage: imageStorage,
		amqp:         amqp,
	}
}

func (m Management) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	const op = "service.management.getUser"
	log := m.log.With(slog.String("op", op))

	user, err := m.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return nil, ErrUserNotExist
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, err
		}

	}

	return user, nil

}

func (m Management) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "service.management.updateUser"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return ErrUserNotExist
		default:
			log.Error("failed to update user", logger.Err(err))
			return err
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

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdatedEventRoutingKey, msg)
	if err != nil {
		log.Error("failed to publish user updated event", logger.Err(err))
		return err
	}

	return nil
}

func (m Management) DeleteUser(ctx context.Context, userID int64) error {
	const op = "service.management.deleteUser"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.DeleteUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return ErrUserNotExist
		default:
			log.Error("failed to delete user", logger.Err(err))
			return err
		}
	}

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserDeletedEventRoutingKey, userID)
	if err != nil {
		log.Error("failed to publish user deleted event", logger.Err(err))
		return err
	}

	return nil
}

func (m Management) SearchUsers(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error) {
	const op = "service.management.searchUsers"
	log := m.log.With(slog.String("op", op))

	users, metadata, err := m.usrStorage.GetAll(ctx, query, filters)
	if err != nil {
		log.Error("failed to get users", logger.Err(err))
		return nil, domain.Metadata{}, err
	}

	return users, metadata, nil

}

func (m Management) UpdateAvatar(ctx context.Context, userID int64, image []byte) (*domain.User, error) {
	const op = "service.management.updateAvatar"
	log := m.log.With(slog.String("op", op))

	user, err := m.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return nil, ErrUserNotExist
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, err
		}
	}

	// Delete previous avatar
	if user.AvatarURL != "" {
		// path.Base returns the last element of the path: object key
		err = m.imageStorage.DeleteImage(ctx, path.Base(user.AvatarURL))
		if err != nil {
			log.Error("failed to delete previous avatar", logger.Err(err))
			return nil, err
		}
	}

	// Compress image
	compressImage, filename, err := imageUtils.CompressImage(image, 75)
	if err != nil {
		log.Error("failed to compress image", logger.Err(err))
		return nil, err
	}

	imageCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	url, err := m.imageStorage.UploadImage(imageCtx, compressImage, filename)
	if err != nil {
		log.Error("failed to upload avatar", logger.Err(err))
		return nil, err
	}
	user.AvatarURL = url

	err = m.usrStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Warn("user not found while updating avatar", logger.Err(err))
			return nil, ErrUserNotExist
		default:
			log.Error("failed to update user avatar url", logger.Err(err))
			return nil, err
		}
	}

	msg := struct {
		ID        int64   `json:"id"`
		AvatarURL *string `json:"avatar_url"`
	}{
		ID:        user.ID,
		AvatarURL: &user.AvatarURL,
	}

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdatedEventRoutingKey, msg)
	if err != nil {
		log.Error("failed to publish user updated event", logger.Err(err))
		return nil, err
	}

	return user, nil
}

func (m Management) ChangeUserRole(ctx context.Context, dto *dtos.ChangeRoleDTO) error {
	const op = "service.management.changeUserRole"
	log := m.log.With(slog.String("op", op))

	user, err := m.usrStorage.GetUserByID(ctx, dto.UserID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return ErrUserNotExist
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	if user.Role != "DSVR" && user.Role != "ADMIN" {
		return ErrUserNonAuthorized
	}

	target, err := m.usrStorage.GetUserByID(ctx, dto.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return ErrUserNotExist
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	if user.Role == "DSVR" {
		if target.Role == "DSVR" || dto.Role == "DSVR" {
			return ErrUserNonAuthorized
		}
	} else if user.Role == "ADMIN" {
		if target.Role == "DSVR" || target.Role == "ADMIN" || dto.Role == "DSVR" || dto.Role == "ADMIN" {
			return ErrUserNonAuthorized
		}
	} else {
		return ErrUserNonAuthorized
	}

	err = m.usrStorage.UpdateUserRole(ctx, target.ID, dto.Role)
	if err != nil {
		log.Error("failed to update user's role", logger.Err(err))
		return err
	}

	msg := domain.Notification{
		UserID:      target.ID,
		Message:     fmt.Sprintf("Your role has been changed to %s", dto.Role),
		Description: fmt.Sprintf("Your role has been changed to %s by %s", dto.Role, user.FirstName+" "+user.LastName),
		Status:      "NEW",
		Severity:    "INFO",
		Source:      "user-service",
		DisplayType: "INBOX",
		CreatedAt:   time.Now().String(),
		ExpiryAt:    "",
	}

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, msg)
	if err != nil {
		log.Error("failed to publish push notification", logger.Err(err))
		return err
	}

	return nil
}
