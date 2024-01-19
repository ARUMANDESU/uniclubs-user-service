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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context,
		email string,
		password string,
	) (token string, err error)
	Register(ctx context.Context,
		user domain.User,
	) (userID int64, err error)
	Logout(ctx context.Context, sessionToken string) error
	CheckUserRole(userId int64, roles []userv1.Role) (bool, error)
}

type serverApi struct {
	userv1.UnimplementedUserServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	userv1.RegisterUserServer(gRPC, &serverApi{auth: auth})
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
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s serverApi) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
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
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.LoginResponse{SessionToken: token}, nil
}

func (s serverApi) Logout(ctx context.Context, req *userv1.LogoutRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) CheckUserRole(ctx context.Context, req *userv1.CheckUserRoleRequest) (*userv1.CheckUserRoleResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) VerifyEmail(ctx context.Context, req *userv1.VerifyEmailRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UnlockAccount(ctx context.Context, req *userv1.UnlockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) LockAccount(ctx context.Context, req *userv1.LockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}
