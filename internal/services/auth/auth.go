package auth

import (
	"context"
	"fmt"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"log/slog"
)

type Auth struct {
	log *slog.Logger
}

func New(log *slog.Logger) *Auth {
	return &Auth{log: log}
}

func (a Auth) Login(ctx context.Context, email string, password string) (token string, err error) {
	//TODO implement me
	panic("implement me")
}

func (a Auth) Register(ctx context.Context, user domain.User) (userID int64, err error) {
	const op = "authService.Register"

	log := a.log.With(slog.String("op", op))
	log.Info(fmt.Sprintf("%v", user))
	//TODO Database logic

	return 0, err
}

func (a Auth) Logout(ctx context.Context, sessionToken string) error {
	//TODO implement me
	panic("implement me")
}

func (a Auth) CheckUserRole(userId int64, roles []userv1.Role) (bool, error) {
	//TODO implement me
	panic("implement me")
}
