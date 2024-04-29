package management

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
