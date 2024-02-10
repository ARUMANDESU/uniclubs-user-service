package auth

import (
	"context"
	"errors"
	"fmt"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/token/activate"
	session "github.com/ARUMANDESU/uniclubs-user-service/pkg/token/session"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log                    *slog.Logger
	usrStorage             UserStorage
	sessionStorage         TokenStorage
	activationTokenStorage TokenStorage
	amqp                   Amqp
}

type Amqp interface {
	Publish(ctx context.Context, routingKey string, msg any) error
}

type UserStorage interface {
	SaveUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID int64) (user *models.User, err error)
	GetUserByEmail(ctx context.Context, email string) (user *models.User, err error)
	GetUserRoleByID(ctx context.Context, userID int64) (role string, err error)
	ActivateUser(ctx context.Context, userID int64) error
}

type TokenStorage interface {
	Create(ctx context.Context, token string, userID int64, duration time.Duration) error
	Get(ctx context.Context, token string) (userID int64, err error)
	Delete(ctx context.Context, sessionToken string) error
}

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserExists               = errors.New("user already exists")
	ErrUserNotExist             = errors.New("user does not exist")
	ErrSessionNotExists         = errors.New("session does not exists")
	ErrActivationTokenNotExists = errors.New("activation token does not exists")
)

func New(
	log *slog.Logger,
	usrStorage UserStorage,
	sessionStorage TokenStorage,
	activateTokenStorage TokenStorage,
	amqp Amqp,
) *Auth {
	return &Auth{
		log:                    log,
		usrStorage:             usrStorage,
		sessionStorage:         sessionStorage,
		activationTokenStorage: activateTokenStorage,
		amqp:                   amqp,
	}
}

func (a Auth) Login(ctx context.Context, email string, password string) (token string, err error) {
	const op = "authService.Login"
	log := a.log.With(slog.String("op", op))

	user, err := a.usrStorage.GetUserByEmail(ctx, email)
	if err != nil {

		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return "", fmt.Errorf("%s: %w", op, err)
		}

	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Info("invalid credentials", logger.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	token, err = session.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.Create(ctx, token, user.ID, time.Hour)
	if err != nil {
		log.Info("can not save session", logger.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a Auth) Register(ctx context.Context, user domain.User) (userID int64, err error) {
	const op = "authService.Register"

	log := a.log.With(slog.String("op", op))

	modelUser := domain.UserToModelUser(user)

	modelUser.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = a.usrStorage.SaveUser(ctx, modelUser)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExists):
			log.Error("user already exists", logger.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)

		default:
			log.Error("failed to save user", logger.Err(err))
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}
	// TODO: implement message broker sending user created

	token, err := activate.GenerateToken()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = a.activationTokenStorage.Create(ctx, token, modelUser.ID, time.Hour*24)
	if err != nil {
		log.Info("can not save activate token", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	msg := struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Token     string `json:"token"`
	}{
		FirstName: modelUser.FirstName,
		LastName:  modelUser.LastName,
		Email:     modelUser.Email,
		Token:     token,
	}

	err = a.amqp.Publish(ctx, "user.registered", msg)
	if err != nil {
		log.Error("failed to publish", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return modelUser.ID, nil
}

func (a Auth) Logout(ctx context.Context, sessionToken string) error {
	const op = "authService.Logout"
	log := a.log.With(slog.String("op", op))

	err := a.sessionStorage.Delete(ctx, sessionToken)
	if err != nil {
		log.Error("failed to delete session", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a Auth) Authenticate(ctx context.Context, sessionToken string) (userID int64, err error) {
	const op = "authService.Authenticate"
	log := a.log.With(slog.String("op", op))

	userID, err = a.sessionStorage.Get(ctx, sessionToken)
	if err != nil {
		log.Error("failed to get session", logger.Err(err))
		switch {
		case errors.Is(err, storage.ErrSessionNotExists):
			return 0, fmt.Errorf("%s, %w", op, ErrSessionNotExists)
		default:
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	return userID, nil
}

func (a Auth) CheckUserRole(ctx context.Context, userId int64, roles []userv1.Role) (bool, error) {
	const op = "authService.CheckUserRole"
	log := a.log.With(slog.String("op", op))

	role, err := a.usrStorage.GetUserRoleByID(ctx, userId)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get role", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}

	for _, r := range roles {
		if r.String() == role {
			return true, nil
		}
	}

	return false, nil
}

func (a Auth) ActivateUser(ctx context.Context, token string) error {
	const op = "authService.ActivateUser"
	log := a.log.With(slog.String("op", op))

	userID, err := a.activationTokenStorage.Get(ctx, token)
	if err != nil {
		log.Error("failed to get activation token", logger.Err(err))
		switch {
		case errors.Is(err, storage.ErrSessionNotExists):
			return fmt.Errorf("%s, %w", op, ErrActivationTokenNotExists)
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = a.usrStorage.ActivateUser(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to activate user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = a.activationTokenStorage.Delete(ctx, token)
	if err != nil {
		log.Error("failed to delete token", logger.Err(err))
	}

	return nil
}
