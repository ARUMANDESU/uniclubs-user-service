package user

import (
	"errors"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/token/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverApi struct {
	userv1.UnimplementedUserServer
	auth       Auth
	management Management
}

func Register(gRPC *grpc.Server, auth Auth, management Management) {
	userv1.RegisterUserServer(gRPC, &serverApi{auth: auth, management: management})
}

func handleError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrActivationTokenNotFound),
		errors.Is(err, domain.ErrRefreshTokenNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, jwt.ErrUserIDMismatch),
		errors.Is(err, jwt.ErrTokenIsNotValid),
		errors.Is(err, jwt.ErrInvalidTokenClaims),
		errors.Is(err, jwt.ErrUserIDClaimNotFound),
		errors.Is(err, jwt.ErrTokenSignatureIsInvalid),
		errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrUserNonAuthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
