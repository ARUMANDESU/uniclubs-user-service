package auth

import (
	"context"
	"errors"
	"fmt"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	token2 "github.com/ARUMANDESU/uniclubs-user-service/pkg/token"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Auth struct {
	log            *slog.Logger
	usrStorage     UserStorage
	sessionStorage SessionStorage
}

type UserStorage interface {
	SaveUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, userID int64) (user *models.User, err error)
	GetUserByEmail(ctx context.Context, email string) (user *models.User, err error)
}

type SessionStorage interface {
	Create(ctx context.Context, sessionToken string, userID int64) error
	Get(ctx context.Context, sessionToken string) (userID int64, err error)
	Delete(ctx context.Context, sessionToken string) error
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotExist       = errors.New("user does not exist")
	ErrSessionNotExists   = errors.New("session does not exists")
)

func New(log *slog.Logger, usrStorage UserStorage, sessionStorage SessionStorage) *Auth {
	return &Auth{log: log, usrStorage: usrStorage, sessionStorage: sessionStorage}
}

func (a Auth) Login(ctx context.Context, email string, password string) (token string, err error) {
	const op = "authService.Login"
	log := a.log.With(slog.String("op", op))

	user, err := a.usrStorage.GetUserByEmail(ctx, email)
	if err != nil {

		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			return "", fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			return "", fmt.Errorf("%s: %w", op, err)
		}

	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	token, err = token2.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = a.sessionStorage.Create(ctx, token, user.ID)
	if err != nil {
		log.Info("can not save session", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
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
		log.Error("failed to generate password hash", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = a.usrStorage.SaveUser(ctx, modelUser)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExists):
			log.Error("user already exists", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)

		default:
			log.Error("failed to save user", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	// TODO: implement message broker sending user created

	return modelUser.ID, nil
}

func (a Auth) Logout(ctx context.Context, sessionToken string) error {
	const op = "authService.Logout"
	log := a.log.With(slog.String("op", op))

	err := a.sessionStorage.Delete(ctx, sessionToken)
	if err != nil {
		log.Error("failed to delete session", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a Auth) Authenticate(ctx context.Context, sessionToken string) (userID int64, err error) {
	const op = "authService.Authenticate"
	log := a.log.With(slog.String("op", op))

	userID, err = a.sessionStorage.Get(ctx, sessionToken)
	if err != nil {
		log.Error("failed to get session", slog.Attr{
			Key:   "error",
			Value: slog.StringValue(err.Error()),
		})
		switch {
		case errors.Is(err, storage.ErrSessionNotExists):
			return 0, fmt.Errorf("%s, %w", op, ErrSessionNotExists)
		default:
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}

	return userID, nil
}

func (a Auth) CheckUserRole(userId int64, roles []userv1.Role) (bool, error) {
	//TODO implement me
	panic("implement me")
}
