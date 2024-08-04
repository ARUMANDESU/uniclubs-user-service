package management

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"testing"

	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/services/management/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type Suite struct {
	Management *Management
	*mocks.UserStorage
	*mocks.TokenStorage
	*mocks.RateLimitStorage
	*mocks.Amqp
}

func Setup(t *testing.T) *Suite {
	t.Helper()
	userStorage := mocks.NewUserStorage(t)
	amqp := mocks.NewAmqp(t)
	tokenStorage := mocks.NewTokenStorage(t)
	rateLimitStorage := mocks.NewRateLimitStorage(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &Suite{
		Management: New(ServiceConfig{
			Log:              logger,
			UserStorage:      userStorage,
			TokenStorage:     tokenStorage,
			RateLimitStorage: rateLimitStorage,
			Amqp:             amqp,
		}),
		UserStorage:      userStorage,
		TokenStorage:     tokenStorage,
		RateLimitStorage: rateLimitStorage,
		Amqp:             amqp,
	}
}

func TestManagement_GetUser(t *testing.T) {
	t.Log("TestManagement_GetUser")

	suite := Setup(t)

	suite.UserStorage.On("GetUserByID", mock.Anything, int64(1)).Return(&domain.User{}, nil)

	user, err := suite.Management.GetUser(context.Background(), 1)

	require.NoError(t, err)

	assert.Equal(t, &domain.User{}, user)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, int64(1))
}

func TestManagement_GetUser_NotFound(t *testing.T) {
	suite := Setup(t)

	suite.UserStorage.On("GetUserByID", mock.Anything, int64(1)).Return(&domain.User{}, domain.ErrUserNotFound)

	user, err := suite.Management.GetUser(context.Background(), 1)
	require.ErrorIs(t, err, domain.ErrUserNotFound)

	assert.Nil(t, user)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, int64(1))
}

func TestManagement_UpdateUser_UserExists(t *testing.T) {
	s := Setup(t)

	dto := dtos.UpdateUserDTO{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe",
		Paths:     []string{"first_name"},
	}
	user := &domain.User{ID: 1}
	s.UserStorage.On("GetUserByID", mock.Anything, dto.UserID).Return(user, nil)
	s.UserStorage.On("UpdateUser", mock.Anything, mock.Anything).Return(nil)
	s.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	updatedUser, err := s.Management.UpdateUser(context.Background(), dto)
	require.NoError(t, err)

	assert.Equal(t, user.FirstName, updatedUser.FirstName)

	s.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, dto.UserID)
	s.UserStorage.AssertCalled(t, "UpdateUser", mock.Anything, user)
}

func TestManagement_UpdateUser_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	dto := dtos.UpdateUserDTO{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe",
		Paths:     []string{"first_name"},
	}
	user := &domain.User{ID: 1}
	suite.UserStorage.On("GetUserByID", mock.Anything, dto.UserID).Return(nil, domain.ErrUserNotFound)

	_, err := suite.Management.UpdateUser(context.Background(), dto)

	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, dto.UserID)
	suite.Amqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.UserStorage.AssertNotCalled(t, "UpdateUser", mock.Anything, user)
}

func TestManagement_DeleteUser_UserExists(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.UserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.UserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(nil)
	suite.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.NoError(t, err)
	suite.UserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
	suite.Amqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestManagement_DeleteUser_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.UserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(domain.ErrUserNotFound)
	suite.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)

	suite.Amqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.UserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
}

func TestManagement_DeleteUser_ErrUnexpected(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.UserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(errors.New("unexpected"))
	suite.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.Error(t, err)

	suite.Amqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.UserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
}

func TestManagement_UpdateAvatar(t *testing.T) {
	suite := Setup(t)

	tests := []struct {
		name                  string
		user                  *domain.User
		newAvatarUrl          string
		expectedPrevAvatarUrl string
	}{
		{
			name:                  "Update avatar",
			user:                  &domain.User{ID: 1, AvatarURL: "avatar1"},
			newAvatarUrl:          "imagine_image_url",
			expectedPrevAvatarUrl: "avatar1",
		},
		{
			name:                  "user doesn't have previous avatar",
			user:                  &domain.User{ID: 2},
			newAvatarUrl:          "imagine_image_url",
			expectedPrevAvatarUrl: "",
		},
		{
			name:                  "user doesn't have previous avatar 2",
			user:                  &domain.User{ID: 3},
			newAvatarUrl:          "imagine_image_url_2",
			expectedPrevAvatarUrl: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			suite.UserStorage.On("GetUserByID", mock.Anything, tt.user.ID).Return(tt.user, nil)
			suite.UserStorage.On("UpdateUser", mock.Anything, tt.user).Return(nil)
			suite.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			userGot, prevImageUrl, err := suite.Management.UpdateAvatar(context.Background(), tt.user.ID, tt.newAvatarUrl)

			require.NoError(t, err)
			assert.Equal(t, tt.newAvatarUrl, userGot.AvatarURL)
			assert.Equal(t, tt.expectedPrevAvatarUrl, prevImageUrl)

			suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, tt.user.ID)
			suite.UserStorage.AssertCalled(t, "UpdateUser", mock.Anything, tt.user)
		})
	}

}

func TestManagement_UpdateAvatar_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	avatar := "avatar1"
	user := &domain.User{ID: 1, AvatarURL: avatar}
	imageUrl := "imagine_image_url"

	suite.UserStorage.On("GetUserByID", mock.Anything, user.ID).Return(&domain.User{}, domain.ErrUserNotFound)

	userGot, _, err := suite.Management.UpdateAvatar(context.Background(), user.ID, imageUrl)

	require.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, userGot)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.UserStorage.AssertNotCalled(t, "UpdateUser", mock.Anything, user)
}

func TestManagement_SearchUsers(t *testing.T) {
	suite := Setup(t)

	query := ""
	filters := domain.Filters{
		Page:     1,
		PageSize: 1,
	}
	users := []*domain.User{{ID: 1}}
	metadata := domain.Metadata{
		CurrentPage:  1,
		PageSize:     1,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 1,
	}

	suite.UserStorage.On("GetAll", mock.Anything, query, filters).Return(users, metadata, nil)

	usersGot, metadataGot, err := suite.Management.SearchUsers(context.Background(), query, filters)

	require.NoError(t, err)
	assert.Equal(t, users, usersGot)
	assert.Equal(t, metadata, metadataGot)

	suite.UserStorage.AssertCalled(t, "GetAll", mock.Anything, query, filters)
}

func TestManagement_SearchUsers_Err(t *testing.T) {
	suite := Setup(t)

	query := ""
	filters := domain.Filters{
		Page:     1,
		PageSize: 1,
	}

	suite.UserStorage.On("GetAll", mock.Anything, query, filters).Return([]*domain.User{}, domain.Metadata{}, errors.New("unexpected"))

	usersGot, metadataGot, err := suite.Management.SearchUsers(context.Background(), query, filters)

	require.Error(t, err)
	assert.Nil(t, usersGot)
	assert.Equal(t, domain.Metadata{}, metadataGot)

	suite.UserStorage.AssertCalled(t, "GetAll", mock.Anything, query, filters)
}

func TestManagement_ChangeUserRole_HappyPath(t *testing.T) {
	suite := Setup(t)

	tests := []struct {
		name   string
		user   *domain.User
		target *domain.User
		role   string
	}{
		{
			name:   "Change role to ADMIN",
			user:   &domain.User{ID: 1, Role: "DSVR"},
			target: &domain.User{ID: 2, Role: "USER"},
			role:   "ADMIN",
		},
		{
			name:   "Change role to USER",
			user:   &domain.User{ID: 3, Role: "DSVR"},
			target: &domain.User{ID: 4, Role: "ADMIN"},
			role:   "USER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 5, Role: "DSVR"},
			target: &domain.User{ID: 6, Role: "ADMIN"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 7, Role: "DSVR"},
			target: &domain.User{ID: 8, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 9, Role: "DSVR"},
			target: &domain.User{ID: 10, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 11, Role: "ADMIN"},
			target: &domain.User{ID: 12, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to USER",
			user:   &domain.User{ID: 13, Role: "ADMIN"},
			target: &domain.User{ID: 14, Role: "MODER"},
			role:   "USER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dto := &dtos.ChangeRoleDTO{
				UserID:   tt.user.ID,
				TargetID: tt.target.ID,
				Role:     tt.role,
			}
			suite.UserStorage.On("GetUserByID", mock.Anything, tt.user.ID).Return(tt.user, nil)
			suite.UserStorage.On("GetUserByID", mock.Anything, tt.target.ID).Return(tt.target, nil)
			suite.UserStorage.On("UpdateRole", mock.Anything, tt.target.ID, tt.role).Return(nil)
			suite.Amqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			err := suite.Management.ChangeUserRole(context.Background(), dto)

			require.NoError(t, err)

			suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, tt.user.ID)
			suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, tt.target.ID)
			suite.UserStorage.AssertCalled(t, "UpdateRole", mock.Anything, tt.target.ID, tt.role)
			suite.Amqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestManagement_ChangeUserRole_FailPath(t *testing.T) {
	suite := Setup(t)

	tests := []struct {
		name        string
		user        *domain.User
		target      *domain.User
		role        string
		expectedErr error
	}{
		{
			name:        "Moder change target role from ADMIN to MODER, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "ADMIN"},
			role:        "MODER",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to MODER, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "MODER",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "ADMIN",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to USER, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "USER",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from USER to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "USER"},
			role:        "ADMIN",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from USER to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "USER"},
			role:        "DSVR",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from MODER to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "DSVR",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from MODER to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "ADMIN",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from DSVR to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "ADMIN",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "ADMIN"},
			role:        "DSVR",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "DSVR change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "DSVR"},
			target:      &domain.User{ID: 2, Role: "ADMIN"},
			role:        "DSVR",
			expectedErr: domain.ErrUserNonAuthorized,
		},
		{
			name:        "DSVR change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "DSVR"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "DSVR",
			expectedErr: domain.ErrUserNonAuthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dto := &dtos.ChangeRoleDTO{
				UserID:   tt.user.ID,
				TargetID: tt.target.ID,
				Role:     tt.role,
			}
			suite.UserStorage.On("GetUserByID", mock.Anything, tt.user.ID).Return(tt.user, nil)
			suite.UserStorage.On("GetUserByID", mock.Anything, tt.target.ID).Return(tt.target, nil)

			err := suite.Management.ChangeUserRole(context.Background(), dto)
			require.ErrorIs(t, err, tt.expectedErr)

			suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, mock.AnythingOfType("int64"))
			suite.UserStorage.AssertNotCalled(t, "UpdateRole", mock.Anything, tt.target.ID, tt.role)
			suite.Amqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestManagement_ChangePassword_Success(t *testing.T) {
	suite := Setup(t)

	dto := dtos.ChangeUserPasswordDTO{
		UserID:  1,
		OldPass: "oldPassword",
		NewPass: "newPassword",
	}

	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(dto.OldPass), bcrypt.DefaultCost)
	user := &domain.User{ID: dto.UserID, PasswordHash: hashedOldPassword}

	suite.UserStorage.On("GetUserByID", mock.Anything, dto.UserID).Return(user, nil)
	suite.UserStorage.On("UpdatePassword", mock.Anything, dto.UserID, mock.Anything).Return(nil)

	err := suite.Management.ChangePassword(context.Background(), dto)

	require.NoError(t, err)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, dto.UserID)
	suite.UserStorage.AssertCalled(t, "UpdatePassword", mock.Anything, dto.UserID, mock.Anything)
}

func TestManagement_ChangePassword_UserNotFound(t *testing.T) {
	suite := Setup(t)

	dto := dtos.ChangeUserPasswordDTO{
		UserID:  1,
		OldPass: "oldPassword",
		NewPass: "newPassword",
	}

	suite.UserStorage.On("GetUserByID", mock.Anything, dto.UserID).Return(nil, domain.ErrUserNotFound)

	err := suite.Management.ChangePassword(context.Background(), dto)

	require.ErrorIs(t, err, domain.ErrUserNotFound)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, dto.UserID)
	suite.UserStorage.AssertNotCalled(t, "UpdatePassword", mock.Anything, dto.UserID, mock.Anything)
}

func TestManagement_ChangePassword_WrongOldPassword(t *testing.T) {
	suite := Setup(t)

	dto := dtos.ChangeUserPasswordDTO{
		UserID:  1,
		OldPass: "wrongOldPassword",
		NewPass: "newPassword",
	}

	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte("oldPassword"), bcrypt.DefaultCost)
	user := &domain.User{ID: dto.UserID, PasswordHash: hashedOldPassword}

	suite.UserStorage.On("GetUserByID", mock.Anything, dto.UserID).Return(user, nil)

	err := suite.Management.ChangePassword(context.Background(), dto)

	require.Error(t, err)

	suite.UserStorage.AssertCalled(t, "GetUserByID", mock.Anything, dto.UserID)
	suite.UserStorage.AssertNotCalled(t, "UpdatePassword", mock.Anything, dto.UserID, mock.Anything)
}
