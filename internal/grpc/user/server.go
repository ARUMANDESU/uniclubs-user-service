package user

import (
	"errors"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"google.golang.org/grpc"
)

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrActivationTokenNotFound = errors.New("activation token not found")
	ErrSessionNotFound         = errors.New("session not found")
	ErrInternal                = errors.New("internal error")
)

type serverApi struct {
	userv1.UnimplementedUserServer
	auth       Auth
	management Management
}

func Register(gRPC *grpc.Server, auth Auth, management Management) {
	userv1.RegisterUserServer(gRPC, &serverApi{auth: auth, management: management})
}
