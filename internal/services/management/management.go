package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/tokens"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
)

const rateLimitCooldown = time.Minute * 5

type Management struct {
	log              *slog.Logger
	userStorage      UserStorage
	tokenStorage     TokenStorage
	rateLimitStorage RateLimitStorage
	amqp             Amqp
}

type ServiceConfig struct {
	Log              *slog.Logger
	UserStorage      UserStorage
	TokenStorage     TokenStorage
	RateLimitStorage RateLimitStorage
	Amqp             Amqp
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=Amqp
type Amqp interface {
	Publish(ctx context.Context, exchangeName rabbitmq.ExchangeName, routingKey rabbitmq.RoutingKey, msg any) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=UserStorage
type UserStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetAll(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error)

	UpdateUser(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, userID int64, role string) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash []byte) error

	DeleteUserByID(ctx context.Context, userID int64) error
	DeleteNonActivatedUsers(ctx context.Context, days int) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=TokenStorage
type TokenStorage interface {
	Create(ctx context.Context, token string, userID int64, duration time.Duration) error
	Get(ctx context.Context, token string) (userID int64, err error)
	Delete(ctx context.Context, sessionToken string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=RateLimitStorage
type RateLimitStorage interface {
	GetLastRequestTime(ctx context.Context, email string) (time.Time, error)
	UpdateLastRequestTime(ctx context.Context, email string, t time.Time) error
}

func New(config ServiceConfig) *Management {
	return &Management{
		log:              config.Log,
		userStorage:      config.UserStorage,
		amqp:             config.Amqp,
		tokenStorage:     config.TokenStorage,
		rateLimitStorage: config.RateLimitStorage,
	}
}

func (m Management) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	const op = "service.management.getUser"
	log := m.log.With(slog.String("op", op))

	user, err := m.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, err
		}

	}

	return user, nil

}

func (m Management) UpdateUser(ctx context.Context, dto dtos.UpdateUserDTO) (domain.User, error) {
	const op = "service.management.updateUser"
	log := m.log.With(slog.String("op", op))

	user, err := m.userStorage.GetUserByID(ctx, dto.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.User{}, err
		default:
			log.Error("failed to get user", logger.Err(err))
			return domain.User{}, domain.ErrInternal
		}
	}

	for _, path := range dto.Paths {
		switch path {
		case "first_name":
			user.FirstName = dto.FirstName
		case "last_name":
			user.LastName = dto.LastName
		case "major":
			user.Major = dto.Major
		case "group_name":
			user.GroupName = dto.GroupName
		case "year":
			user.Year = dto.Year
		}
	}

	err = m.userStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.User{}, err
		default:
			log.Error("failed to update user", logger.Err(err))
			return domain.User{}, domain.ErrInternal
		}
	}

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdated, user)
	if err != nil {
		log.Error("failed to publish user updated event", logger.Err(err))
		return domain.User{}, err
	}

	return *user, nil
}

func (m Management) DeleteUser(ctx context.Context, userID int64) error {
	const op = "service.management.deleteUser"
	log := m.log.With(slog.String("op", op))

	err := m.userStorage.DeleteUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to delete user", logger.Err(err))
			return err
		}
	}

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserDeleted, userID)
	if err != nil {
		log.Error("failed to publish user deleted event", logger.Err(err))
		return err
	}

	return nil
}

func (m Management) SearchUsers(ctx context.Context, query string, filters domain.Filters) ([]*domain.User, domain.Metadata, error) {
	const op = "service.management.searchUsers"
	log := m.log.With(slog.String("op", op))

	users, metadata, err := m.userStorage.GetAll(ctx, query, filters)
	if err != nil {
		log.Error("failed to get users", logger.Err(err))
		return nil, domain.Metadata{}, err
	}

	return users, metadata, nil

}

func (m Management) UpdateAvatar(ctx context.Context, userID int64, imageUrl string) (user *domain.User, prevImage string, err error) {
	const op = "service.management.updateAvatar"
	log := m.log.With(slog.String("op", op))

	user, err = m.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, "", domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return nil, "", err
		}
	}

	if user.AvatarURL != "" {
		prevImage = user.AvatarURL
	}

	user.AvatarURL = imageUrl

	err = m.userStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			log.Warn("user not found while updating avatar", logger.Err(err))
			return nil, "", domain.ErrUserNotFound
		default:
			log.Error("failed to update user avatar url", logger.Err(err))
			return nil, "", err
		}
	}

	go func() {
		user := user
		err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserUpdated, user)
		if err != nil {
			log.Error("failed to publish user updated event", logger.Err(err))
		}
	}()

	return user, prevImage, nil
}

func (m Management) ChangeUserRole(ctx context.Context, dto *dtos.ChangeRoleDTO) error {
	const op = "service.management.changeUserRole"
	log := m.log.With(slog.String("op", op))

	user, err := m.userStorage.GetUserByID(ctx, dto.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	if user.Role != "DSVR" && user.Role != "ADMIN" {
		return domain.ErrUserNonAuthorized
	}

	target, err := m.userStorage.GetUserByID(ctx, dto.TargetID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	userRolePosition := domain.RoleMap[user.Role]
	targetRolePosition := domain.RoleMap[target.Role]
	rolePosition := domain.RoleMap[dto.Role]

	if userRolePosition <= targetRolePosition || userRolePosition <= rolePosition {
		return domain.ErrUserNonAuthorized
	}

	err = m.userStorage.UpdateRole(ctx, target.ID, dto.Role)
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

	err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotification, msg)
	if err != nil {
		log.Error("failed to publish push notification", logger.Err(err))
		return err
	}

	return nil
}

func (m Management) DeleteNonActivatedUsers(ctx context.Context, days int) error {
	const op = "service.management.deleteNonActivatedUsers"
	log := m.log.With(slog.String("op", op))

	err := m.userStorage.DeleteNonActivatedUsers(ctx, days)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to delete non activated users", logger.Err(err))
			return domain.ErrInternal
		}
	}

	return nil
}

func (m Management) ChangePassword(ctx context.Context, dto dtos.ChangeUserPasswordDTO) error {
	const op = "service.management.changePassword"
	log := m.log.With(slog.String("op", op))

	user, err := m.userStorage.GetUserByID(ctx, dto.UserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	// compare old password and hash from db
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(dto.OldPass)); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidCredentials, "old password is incorrect")
	}

	// generate new password hash
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.NewPass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))
		return domain.ErrInternal
	}

	err = m.userStorage.UpdatePassword(ctx, user.ID, passwordHash)
	if err != nil {
		log.Error("failed to update user password", logger.Err(err))
		return domain.ErrInternal
	}

	return nil
}

func (m Management) ForgotPassword(ctx context.Context, email, barcode string) error {
	const op = "service.management.forgotPassword"
	log := m.log.With(slog.String("op", op))

	err := m.checkRateLimit(ctx, email)
	if err != nil {
		return err
	}

	user, err := m.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound
		default:
			log.Error("failed to get user", logger.Err(err))
			return err
		}
	}

	if user.Barcode != barcode {
		return domain.ErrInvalidCredentials
	}

	token, err := tokens.Generate()
	if err != nil {
		return err
	}

	err = m.tokenStorage.Create(ctx, token, user.ID, time.Minute*15)
	if err != nil {
		log.Error("failed to save token", logger.Err(err))
		return err
	}

	go func() {
		msg := struct {
			Email      string    `json:"email"`
			Token      string    `json:"token"`
			ValidUntil time.Time `json:"valid_until"`
		}{
			Email:      email,
			Token:      token,
			ValidUntil: time.Now().Add(time.Minute * 15),
		}

		err = m.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserForgotPassword, msg)
		if err != nil {
			log.Error("failed to publish push notification", logger.Err(err))
		}
	}()

	return nil
}

func (m Management) ResetPassword(ctx context.Context, token, newPassword string) error {
	const op = "service.management.resetPassword"
	log := m.log.With(slog.String("op", op))

	userID, err := m.tokenStorage.Get(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTokenNotFound):
			return domain.ErrTokenNotFound
		default:
			log.Error("failed to get token", logger.Err(err))
			return err
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))
		return domain.ErrInternal
	}

	err = m.userStorage.UpdatePassword(ctx, userID, passwordHash)
	if err != nil {
		log.Error("failed to update user password", logger.Err(err))
		return domain.ErrInternal
	}

	return nil

}

func (m Management) checkRateLimit(ctx context.Context, email string) error {
	lastRequestTime, err := m.rateLimitStorage.GetLastRequestTime(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return m.rateLimitStorage.UpdateLastRequestTime(ctx, email, time.Now())
		}
		return err
	}

	if time.Since(lastRequestTime) < rateLimitCooldown {
		return domain.ErrRateLimitExceeded
	}

	// Update last request time
	return m.rateLimitStorage.UpdateLastRequestTime(ctx, email, time.Now())
}
