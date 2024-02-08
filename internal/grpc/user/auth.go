package user

import (
	"context"
	"errors"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/auth"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
	) (token string, err error)
	Register(ctx context.Context, user domain.User) (userID int64, err error)
	Logout(ctx context.Context, sessionToken string) error
	Authenticate(ctx context.Context, sessionToken string) (userID int64, err error)
	CheckUserRole(
		ctx context.Context,
		userId int64,
		roles []userv1.Role,
	) (bool, error)
	ActivateUser(ctx context.Context, token string) error
}

func (s serverApi) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {

	err := validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(6, 64)),
		validation.Field(&req.Barcode, validation.Required),
		validation.Field(&req.FirstName, validation.Required),
		validation.Field(&req.LastName, validation.Required),
		validation.Field(&req.Major, validation.Required),
		validation.Field(&req.Year, validation.Required),
		validation.Field(&req.GroupName, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user := domain.User{
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		Barcode:   req.GetBarcode(),
		Major:     req.GetMajor(),
		GroupName: req.GetGroupName(),
		Year:      int(req.GetYear()),
	}

	userID, err := s.auth.Register(ctx, user)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, ErrUserAlreadyExists.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s serverApi) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(6, 64)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserNotExist):
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		case errors.Is(err, auth.ErrInvalidCredentials):
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}

	}

	return &userv1.LoginResponse{SessionToken: token}, nil
}

func (s serverApi) Logout(ctx context.Context, req *userv1.LogoutRequest) (*empty.Empty, error) {
	err := validation.Validate(&req.SessionToken, validation.Required)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.Logout(ctx, req.GetSessionToken())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) Authenticate(ctx context.Context, req *userv1.AuthenticateRequest) (*userv1.AuthenticateResponse, error) {
	err := validation.Validate(&req.SessionToken, validation.Required)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := s.auth.Authenticate(ctx, req.GetSessionToken())
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotExists) {
			return nil, status.Error(codes.NotFound, ErrSessionNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.AuthenticateResponse{UserId: userID}, nil
}

func (s serverApi) CheckUserRole(ctx context.Context, req *userv1.CheckUserRoleRequest) (*userv1.CheckUserRoleResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.UserId, validation.Required),
		validation.Field(&req.Roles, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hasRole, err := s.auth.CheckUserRole(ctx, req.GetUserId(), req.GetRoles())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotExist) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.CheckUserRoleResponse{HasRole: hasRole}, nil

}

func (s serverApi) ActivateUser(ctx context.Context, req *userv1.ActivateUserRequest) (*empty.Empty, error) {
	err := validation.Validate(&req.VerificationToken, validation.Required, validation.Length(31, 33))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.ActivateUser(ctx, req.GetVerificationToken())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrActivationTokenNotExists):
			return nil, status.Error(codes.NotFound, ErrActivationTokenNotFound.Error())
		case errors.Is(err, auth.ErrUserNotExist):
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}

	}

	return &empty.Empty{}, nil

}
