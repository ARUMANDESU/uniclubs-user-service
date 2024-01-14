package user

import (
	"context"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type serverApi struct {
	userv1.UnimplementedUserServer
}

func Register(gRPC *grpc.Server) {
	userv1.RegisterUserServer(gRPC, &serverApi{})
}

func (s serverApi) Register(ctx context.Context, in *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) DeleteUser(ctx context.Context, in *userv1.DeleteUserRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) Login(ctx context.Context, in *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) Logout(ctx context.Context, in *userv1.LogoutRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) CheckUserRole(ctx context.Context, in *userv1.CheckUserRoleRequest) (*userv1.CheckUserRoleResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) ListUsers(ctx context.Context, in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) VerifyEmail(ctx context.Context, in *userv1.VerifyEmailRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UnlockAccount(ctx context.Context, in *userv1.UnlockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) LockAccount(ctx context.Context, in *userv1.LockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}
