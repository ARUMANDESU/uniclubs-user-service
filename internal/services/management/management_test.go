package management

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestManagement_GetUser(t *testing.T) {
	t.Log("TestManagement_GetUser")

	suite := Setup(t)

	suite.MockUserStorage.On("GetUserByID", mock.Anything, int64(1)).Return(&domain.User{}, nil)

	user, err := suite.Management.GetUser(context.Background(), 1)

	require.NoError(t, err)

	assert.Equal(t, &domain.User{}, user)

	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, int64(1))
}

func TestManagement_GetUser_NotFound(t *testing.T) {
	suite := Setup(t)

	suite.MockUserStorage.On("GetUserByID", mock.Anything, int64(1)).Return(&domain.User{}, storage.ErrUserNotExists)

	user, err := suite.Management.GetUser(context.Background(), 1)
	require.ErrorIs(t, err, ErrUserNotExist)

	assert.Nil(t, user)

	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, int64(1))
}

func TestManagement_UpdateUser_UserExists(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.MockUserStorage.On("UpdateUser", mock.Anything, user).Return(nil)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.UpdateUser(context.Background(), user)

	require.NoError(t, err)
	suite.MockUserStorage.AssertCalled(t, "UpdateUser", mock.Anything, user)
}

func TestManagement_UpdateUser_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(nil, storage.ErrUserNotExists)
	suite.MockUserStorage.On("UpdateUser", mock.Anything, user).Return(storage.ErrUserNotExists)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.UpdateUser(context.Background(), user)

	assert.ErrorIs(t, err, ErrUserNotExist)

	suite.MockAmqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.MockUserStorage.AssertCalled(t, "UpdateUser", mock.Anything, user)
}

func TestManagement_DeleteUser_UserExists(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.MockUserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(nil)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.NoError(t, err)
	suite.MockUserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
	suite.MockAmqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestManagement_DeleteUser_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.MockUserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(storage.ErrUserNotExists)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotExist)

	suite.MockAmqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.MockUserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
}

func TestManagement_DeleteUser_ErrUnexpected(t *testing.T) {
	suite := Setup(t)

	user := &domain.User{ID: 1}
	suite.MockUserStorage.On("DeleteUserByID", mock.Anything, user.ID).Return(errors.New("unexpected"))
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Management.DeleteUser(context.Background(), user.ID)

	require.Error(t, err)

	suite.MockAmqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.MockUserStorage.AssertCalled(t, "DeleteUserByID", mock.Anything, user.ID)
}

func TestManagement_UpdateAvatar(t *testing.T) {
	suite := Setup(t)

	avatar := "avatar1"
	user := &domain.User{ID: 1, AvatarURL: avatar}
	image, err := os.ReadFile("../../../tests/image/test_image.jpg")
	require.NoError(t, err, "failed to read test image")

	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.MockImageStorage.On("DeleteImage", mock.Anything, avatar).Return(nil)
	suite.MockImageStorage.On("UploadImage", mock.Anything, mock.Anything, mock.Anything).Return("avatar2", nil)
	suite.MockUserStorage.On("UpdateUser", mock.Anything, user).Return(nil)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	userGot, err := suite.Management.UpdateAvatar(context.Background(), user.ID, image)

	require.NoError(t, err)
	assert.Equal(t, "avatar2", userGot.AvatarURL)

	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.MockImageStorage.AssertCalled(t, "DeleteImage", mock.Anything, avatar)
	suite.MockImageStorage.AssertCalled(t, "UploadImage", mock.Anything, mock.Anything, mock.Anything)
	suite.MockUserStorage.AssertCalled(t, "UpdateUser", mock.Anything, user)
	suite.MockAmqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

}

func TestManagement_UpdateAvatar_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	avatar := "avatar1"
	user := &domain.User{ID: 1, AvatarURL: avatar}
	image, err := os.ReadFile("../../../tests/image/test_image.jpg")
	require.NoError(t, err, "failed to read test image")

	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(&domain.User{}, storage.ErrUserNotExists)

	userGot, err := suite.Management.UpdateAvatar(context.Background(), user.ID, image)

	require.ErrorIs(t, err, ErrUserNotExist)
	assert.Nil(t, userGot)

	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.MockImageStorage.AssertNotCalled(t, "DeleteImage", mock.Anything, avatar)
	suite.MockImageStorage.AssertNotCalled(t, "UploadImage", mock.Anything, mock.Anything, mock.Anything)
	suite.MockUserStorage.AssertNotCalled(t, "UpdateUser", mock.Anything, user)
	suite.MockAmqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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

	suite.MockUserStorage.On("GetAll", mock.Anything, query, filters).Return(users, metadata, nil)

	usersGot, metadataGot, err := suite.Management.SearchUsers(context.Background(), query, filters)

	require.NoError(t, err)
	assert.Equal(t, users, usersGot)
	assert.Equal(t, metadata, metadataGot)

	suite.MockUserStorage.AssertCalled(t, "GetAll", mock.Anything, query, filters)
}

func TestManagement_SearchUsers_Err(t *testing.T) {
	suite := Setup(t)

	query := ""
	filters := domain.Filters{
		Page:     1,
		PageSize: 1,
	}

	suite.MockUserStorage.On("GetAll", mock.Anything, query, filters).Return([]*domain.User{}, domain.Metadata{}, errors.New("unexpected"))

	usersGot, metadataGot, err := suite.Management.SearchUsers(context.Background(), query, filters)

	require.Error(t, err)
	assert.Nil(t, usersGot)
	assert.Equal(t, domain.Metadata{}, metadataGot)

	suite.MockUserStorage.AssertCalled(t, "GetAll", mock.Anything, query, filters)
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
			user:   &domain.User{ID: 1, Role: "DSVR"},
			target: &domain.User{ID: 2, Role: "ADMIN"},
			role:   "USER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 1, Role: "DSVR"},
			target: &domain.User{ID: 2, Role: "ADMIN"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 1, Role: "DSVR"},
			target: &domain.User{ID: 2, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 1, Role: "DSVR"},
			target: &domain.User{ID: 2, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to MODER",
			user:   &domain.User{ID: 1, Role: "ADMIN"},
			target: &domain.User{ID: 2, Role: "USER"},
			role:   "MODER",
		},
		{
			name:   "Change role to USER",
			user:   &domain.User{ID: 1, Role: "ADMIN"},
			target: &domain.User{ID: 2, Role: "MODER"},
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
			suite.MockUserStorage.On("GetUserByID", mock.Anything, tt.user.ID).Return(tt.user, nil)
			suite.MockUserStorage.On("GetUserByID", mock.Anything, tt.target.ID).Return(tt.target, nil)
			suite.MockUserStorage.On("UpdateUserRole", mock.Anything, tt.target.ID, tt.role).Return(nil)
			suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			err := suite.Management.ChangeUserRole(context.Background(), dto)

			require.NoError(t, err)

			suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, tt.user.ID)
			suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, tt.target.ID)
			suite.MockUserStorage.AssertCalled(t, "UpdateUserRole", mock.Anything, tt.target.ID, tt.role)
			suite.MockAmqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to MODER, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "MODER",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "ADMIN",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "Moder change target role from DSVR to USER, user unauthorized",
			user:        &domain.User{ID: 1, Role: "MODER"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "USER",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from USER to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "USER"},
			role:        "ADMIN",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from USER to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "USER"},
			role:        "DSVR",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from MODER to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "DSVR",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from MODER to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "ADMIN",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from DSVR to ADMIN, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "DSVR"},
			role:        "ADMIN",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "ADMIN change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "ADMIN"},
			target:      &domain.User{ID: 2, Role: "ADMIN"},
			role:        "DSVR",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "DSVR change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "DSVR"},
			target:      &domain.User{ID: 2, Role: "ADMIN"},
			role:        "DSVR",
			expectedErr: ErrUserNonAuthorized,
		},
		{
			name:        "DSVR change target role from ADMIN to DSVR, user unauthorized",
			user:        &domain.User{ID: 1, Role: "DSVR"},
			target:      &domain.User{ID: 2, Role: "MODER"},
			role:        "DSVR",
			expectedErr: ErrUserNonAuthorized,
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
			suite.MockUserStorage.On("GetUserByID", mock.Anything, tt.user.ID).Return(tt.user, nil)
			suite.MockUserStorage.On("GetUserByID", mock.Anything, tt.target.ID).Return(tt.target, nil)

			err := suite.Management.ChangeUserRole(context.Background(), dto)
			require.ErrorIs(t, err, tt.expectedErr)

			suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, mock.AnythingOfType("int64"))
			suite.MockUserStorage.AssertNotCalled(t, "UpdateUserRole", mock.Anything, tt.target.ID, tt.role)
			suite.MockAmqp.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
