package auth

import (
	"context"
	"errors"
	"fmt"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Auth struct {
	log        *slog.Logger
	usrStorage UserStorage
}

type UserStorage interface {
	SaveUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, userID int64) (user models.User, err error)
}

//TODO: interface for session management

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotExist       = errors.New("user does not exist")
)

func New(log *slog.Logger, usrStorage UserStorage) *Auth {
	return &Auth{log: log, usrStorage: usrStorage}
}

func (a Auth) Login(ctx context.Context, email string, password string) (token string, err error) {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (a Auth) CheckUserRole(userId int64, roles []userv1.Role) (bool, error) {
	//TODO implement me
	panic("implement me")
}
