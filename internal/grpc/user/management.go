package user

import (
	"context"
	"errors"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/management"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Management interface {
	GetUser(ctx context.Context, userID int64) (user *domain.User, err error)
	SearchUsers(
		ctx context.Context,
		query string,
		filters domain.Filters,
	) (users []*domain.User, metadata domain.Metadata, err error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, userID int64) error
	UpdateAvatar(ctx context.Context, userID int64, image []byte) error
}

func (s serverApi) UpdateUser(ctx context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.UserId, validation.Required),
		validation.Field(&req.Year, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.management.GetUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, management.ErrUserNotExist) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	paths := req.GetUpdateMask().GetPaths()
	for _, path := range paths {
		switch path {
		case "first_name":
			user.FirstName = req.GetFirstName()
		case "last_name":
			user.LastName = req.GetLastName()
		case "major":
			user.Major = req.GetMajor()
		case "group_name":
			user.GroupName = req.GetGroupName()
		case "year":
			user.Year = int(req.GetYear())
		}
	}

	err = s.management.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, management.ErrUserNotExist) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.UpdateUserResponse{UserId: user.ID}, nil

}

func (s serverApi) DeleteUser(ctx context.Context, req *userv1.DeleteUserRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.UserId, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.management.DeleteUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, management.ErrUserNotExist) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil

}

func (s serverApi) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	err := validation.Validate(&req.UserId, validation.Required, validation.Min(1))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.management.GetUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, management.ErrUserNotExist) {
			return nil, status.Error(codes.NotFound, ErrUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userv1.GetUserResponse{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarURL,
		Barcode:   user.Barcode,
		Major:     user.Major,
		GroupName: user.GroupName,
		Year:      int32(user.Year),
		CreatedAt: timestamppb.New(user.CreatedAt),
		Role:      user.MapRoleStringToEnum(),
	}, nil

}

func (s serverApi) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.PageNumber, validation.Required, validation.Min(1)),
		validation.Field(&req.PageSize, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	f := domain.Filters{
		Page:     req.GetPageNumber(),
		PageSize: req.GetPageSize(),
	}

	users, metadata, err := s.management.SearchUsers(ctx, req.GetQuery(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &userv1.SearchUsersResponse{
		Users: domain.MapUserArrToUserObjectArr(users),
		Metadata: &userv1.SearchUsersMetadata{
			CurrentPage:  metadata.CurrentPage,
			PageSize:     metadata.PageSize,
			FirstPage:    metadata.FirstPage,
			LastPage:     metadata.LastPage,
			TotalRecords: metadata.TotalRecords,
		},
	}, nil
}

func (s serverApi) UnlockAccount(ctx context.Context, req *userv1.UnlockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) LockAccount(ctx context.Context, req *userv1.LockAccountRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UpdateAvatar(ctx context.Context, req *userv1.UpdateAvatarRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.Image, validation.Required),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.management.UpdateAvatar(ctx, req.GetUserId(), req.GetImage())
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}
