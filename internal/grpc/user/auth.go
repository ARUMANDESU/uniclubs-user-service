package user

import (
	"context"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=Auth
type Auth interface {
	Login(ctx context.Context, email string, password string) (dto dtos.UserCredentialsDTO, err error)
	Register(ctx context.Context, user *dtos.UserRegisterDTO) (userID int64, err error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, rtToken, jwtToken string) (dtos.UserCredentialsDTO, error)
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
		validation.Field(&req.Password, validation.Required, validation.By(validatePassword)),
		validation.Field(&req.Barcode, validation.Required),
		validation.Field(&req.FirstName, validation.Required, validation.By(validateName)),
		validation.Field(&req.LastName, validation.Required, validation.By(validateName)),
		validation.Field(&req.Major, validation.By(validateMajor)),
		validation.Field(&req.Year, validation.Required, validation.By(validateYear)),
		validation.Field(&req.GroupName, validation.By(validateGroupName)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := s.auth.Register(ctx, dtos.RegisterRequestToDTO(req))
	if err != nil {
		return nil, handleError(err)
	}

	return &userv1.RegisterResponse{UserId: userID}, nil
}

func (s serverApi) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dto, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, handleError(err)
	}

	return &userv1.LoginResponse{User: dto.User.ToUserObject(), JwtToken: dto.JWTToken, RtToken: dto.RtToken}, nil
}

func (s serverApi) Logout(ctx context.Context, req *userv1.LogoutRequest) (*empty.Empty, error) {
	err := validation.Validate(&req.RtToken, validation.Required)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.Logout(ctx, req.GetRtToken())
	if err != nil {
		return nil, handleError(err)
	}

	return &empty.Empty{}, nil
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
		return nil, handleError(err)
	}

	return &userv1.CheckUserRoleResponse{HasRole: hasRole}, nil

}

func (s serverApi) ActivateUser(ctx context.Context, req *userv1.ActivateUserRequest) (*empty.Empty, error) {
	err := validation.Validate(&req.VerificationToken, validation.Required)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.auth.ActivateUser(ctx, req.GetVerificationToken())
	if err != nil {
		return nil, handleError(err)
	}

	return &empty.Empty{}, nil

}

func (s serverApi) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.RtToken, validation.Required),
		validation.Field(&req.JwtToken, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	dto, err := s.auth.RefreshToken(ctx, req.GetRtToken(), req.GetJwtToken())
	if err != nil {
		return nil, handleError(err)
	}

	return &userv1.RefreshTokenResponse{
		JwtToken: dto.JWTToken,
		RtToken:  dto.RtToken,
		User:     dto.User.ToUserObject(),
	}, nil

}
