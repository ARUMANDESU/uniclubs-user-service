package auth

import (
	"context"
	"errors"
	"fmt"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/logger"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/token/activate"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/token/jwt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log                    *slog.Logger
	JwtCfg                 config.JWTConfig
	usrStorage             UserStorage
	sessionStorage         TokenStorage
	activationTokenStorage TokenStorage
	amqp                   Amqp
}

type Amqp interface {
	Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error
}

type UserStorage interface {
	SaveUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	GetUserByEmail(ctx context.Context, email string) (user *domain.User, err error)
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
	ErrRefreshTokenNotExists    = errors.New("refresh token not found")
	ErrActivationTokenNotExists = errors.New("activation token does not exists")
)

func New(
	log *slog.Logger,
	JwtCfg config.JWTConfig,
	usrStorage UserStorage,
	sessionStorage TokenStorage,
	activateTokenStorage TokenStorage,
	amqp Amqp,
) *Auth {
	return &Auth{
		log:                    log,
		JwtCfg:                 JwtCfg,
		usrStorage:             usrStorage,
		sessionStorage:         sessionStorage,
		activationTokenStorage: activateTokenStorage,
		amqp:                   amqp,
	}
}

func (a Auth) Login(ctx context.Context, email string, password string) (dtos.UserCredentialsDTO, error) {
	const op = "authService.Login"
	log := a.log.With(slog.String("op", op))

	user, err := a.usrStorage.GetUserByEmail(ctx, email)
	if err != nil {

		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, err)
		}

	}
	// compare password and hash from db
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	// Generate a pair of Access and Refresh tokens
	tokenPair, err := jwt.GenerateTokenPair(user.ID, a.JwtCfg)
	if err != nil {
		log.Error("failed to generate token pair", logger.Err(err))
		return dtos.UserCredentialsDTO{}, err
	}

	// save rt token
	err = a.sessionStorage.Create(ctx, tokenPair["refresh_token"], user.ID, time.Hour*24*30)
	if err != nil {
		log.Info("failed to save refresh token", logger.Err(err))
		return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return dtos.UserCredentialsDTO{
		User:     user,
		JWTToken: tokenPair["access_token"],
		RtToken:  tokenPair["refresh_token"],
	}, nil
}

func (a Auth) Register(ctx context.Context, dto *dtos.UserRegisterDTO) (userID int64, err error) {
	const op = "authService.Register"

	log := a.log.With(slog.String("op", op))

	user := dto.ToDomain()

	user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", logger.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = a.usrStorage.SaveUser(ctx, user)
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

	token, err := activate.GenerateToken()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = a.activationTokenStorage.Create(ctx, token, user.ID, time.Hour*24)
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
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Token:     token,
	}

	err = a.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.UserRegisteredEventRoutingKey, msg)
	if err != nil {
		log.Error("failed to publish", logger.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, nil
}

func (a Auth) Logout(ctx context.Context, refreshToken string) error {
	const op = "authService.Logout"
	log := a.log.With(slog.String("op", op))

	err := a.sessionStorage.Delete(ctx, refreshToken)
	if err != nil {
		log.Error("failed to delete refresh token", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a Auth) RefreshToken(ctx context.Context, rtToken, jwtToken string) (dtos.UserCredentialsDTO, error) {
	const op = "authService.RefreshToken"
	log := a.log.With(slog.String("op", op))

	userID, err := a.sessionStorage.Get(ctx, rtToken)
	if err != nil {
		log.Error("failed to get session", logger.Err(err))
		switch {
		case errors.Is(err, storage.ErrTokenNotExists):
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s, %w", op, ErrRefreshTokenNotExists)
		default:
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	userIDFromToken, err := jwt.GetUserIDFromToken(jwtToken, a.JwtCfg.AccessTokenSecret)
	if err != nil && !errors.Is(err, domain.ErrTokenIsExpired) {
		return dtos.UserCredentialsDTO{}, err
	}

	if userIDFromToken != userID {
		return dtos.UserCredentialsDTO{}, domain.ErrUserIDMismatch
	}

	user, err := a.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, err)
		}

	}

	tokenPair, err := jwt.GenerateTokenPair(userID, a.JwtCfg)
	if err != nil {
		log.Error("failed to generate token pair", logger.Err(err))
		return dtos.UserCredentialsDTO{}, err
	}

	// save rt token
	err = a.sessionStorage.Create(ctx, tokenPair["refresh_token"], user.ID, time.Hour*24*30)
	if err != nil {
		log.Info("failed to save refresh token", logger.Err(err))
		return dtos.UserCredentialsDTO{}, fmt.Errorf("%s: %w", op, err)
	}

	return dtos.UserCredentialsDTO{
		User:     user,
		JWTToken: tokenPair["access_token"],
		RtToken:  tokenPair["refresh_token"],
	}, nil
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
		case errors.Is(err, storage.ErrTokenNotExists):
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

	user, err := a.usrStorage.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}

	}
	msg := struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Barcode   string `json:"barcode"`
		AvatarURL string `json:"avatar_url"`
	}{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Barcode:   user.Barcode,
		AvatarURL: user.AvatarURL,
	}

	err = a.amqp.Publish(ctx, "", rabbitmq.UserActivatedEventRoutingKey, msg)
	if err != nil {
		log.Error("failed to publish user.activated", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.activationTokenStorage.Delete(ctx, token)
	if err != nil {
		log.Error("failed to delete token", logger.Err(err))
	}

	return nil
}
