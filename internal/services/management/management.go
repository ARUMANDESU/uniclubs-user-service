package management

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
)

var (
	ErrUserNotExist      = errors.New("user does not exist")
	ErrUserNonAuthorized = errors.New("user is not authorized")
)

type Management struct {
	log        *slog.Logger
	usrStorage UserStorage
	amqp       Amqp
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=Amqp
type Amqp interface {
	Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=UserStorage
type UserStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateUserRole(ctx context.Context, userID int64, role string) error
	DeleteUserByID(ctx context.Context, userID int64) error
	GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error)
	DeleteNonActivatedUsers(ctx context.Context, days int) error
}

func New(log *slog.Logger, storage UserStorage, amqp Amqp) *Management {
	return &Management{
		log:        log,
		usrStorage: storage,
		amqp:       amqp,
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

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdatedEventRoutingKey, user)
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

func (m Management) UpdateAvatar(ctx context.Context, userID int64, imageUrl string) (user *domain.User, prevImage string, err error) {
	const op = "service.management.updateAvatar"
	log := m.log.With(slog.String("op", op))

	user, err = m.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return nil, "", ErrUserNotExist
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, "", err
		}
	}

	if user.AvatarURL != "" {
		prevImage = user.AvatarURL
	}

	user.AvatarURL = imageUrl

	err = m.usrStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Warn("user not found while updating avatar", logger.Err(err))
			return nil, "", ErrUserNotExist
		default:
			log.Error("failed to update user avatar url", logger.Err(err))
			return nil, "", err
		}
	}

	go func() {
		user := user
		err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdatedEventRoutingKey, user)
		if err != nil {
			log.Error("failed to publish user updated event", logger.Err(err))
		}
	}()

	return user, prevImage, nil
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

	userRolePosition := domain.RoleMap[user.Role]
	targetRolePosition := domain.RoleMap[target.Role]
	rolePosition := domain.RoleMap[dto.Role]

	if userRolePosition <= targetRolePosition || userRolePosition <= rolePosition {
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

func (m Management) DeleteNonActivatedUsers(ctx context.Context, days int) error {
	const op = "service.management.deleteNonActivatedUsers"
	log := m.log.With(slog.String("op", op))

	err := m.usrStorage.DeleteNonActivatedUsers(ctx, days)
	if err != nil {
		return handleError(err, log, "failed to delete inactived users")
	}

	return nil
}

func handleError(err error, log *slog.Logger, op string) error {
	switch {
	case errors.Is(err, storage.ErrUserNotExists):
		return ErrUserNotExist
	default:
		log.Error(op, logger.Err(err))
		return err
	}
}
