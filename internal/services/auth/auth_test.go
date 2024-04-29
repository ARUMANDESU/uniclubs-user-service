package auth

import (
	"context"
	"errors"
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/token/jwt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestAuth_Login(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	password := gofakeit.Password(true, true, true, false, false, 12)

	user := &domain.User{
		ID:    1,
		Email: gofakeit.Email(),
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user.PasswordHash = passwordHash
	suite.MockTokenStorage.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	suite.MockUserStorage.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)

	result, err := suite.Auth.Login(context.Background(), user.Email, password)

	// Assert that no error was returned
	assert.NoError(t, err)
	assert.Equal(t, user, result.User)

	// Assert that the GetUserByEmail function was called with the correct parameters
	suite.MockUserStorage.AssertCalled(t, "GetUserByEmail", mock.Anything, user.Email)
}

func TestAuth_Login_InvalidCredentials(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	password := gofakeit.Password(true, true, true, false, false, 12)

	user := &domain.User{
		ID:    1,
		Email: gofakeit.Email(),
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user.PasswordHash = passwordHash

	suite.MockUserStorage.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil)

	_, err = suite.Auth.Login(context.Background(), user.Email, "wrong_password")

	// Assert that an error was returned
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCredentials)

	// Assert that the GetUserByEmail function was called with the correct parameters
	suite.MockUserStorage.AssertCalled(t, "GetUserByEmail", mock.Anything, user.Email)

}

func TestAuth_CheckUserRole_UserExistsAndHasRole(t *testing.T) {
	suite := Setup(t)

	userID := int64(1)
	roles := []userv1.Role{userv1.Role_ADMIN}

	suite.MockUserStorage.On("GetUserRoleByID", mock.Anything, userID).Return("ADMIN", nil)

	result, err := suite.Auth.CheckUserRole(context.Background(), userID, roles)

	assert.NoError(t, err)
	assert.True(t, result)

	suite.MockUserStorage.AssertCalled(t, "GetUserRoleByID", mock.Anything, userID)
}

func TestAuth_CheckUserRole_UserExistsAndDoesNotHaveRole(t *testing.T) {
	suite := Setup(t)

	userID := int64(1)
	roles := []userv1.Role{userv1.Role_USER}

	suite.MockUserStorage.On("GetUserRoleByID", mock.Anything, userID).Return("ADMIN", nil)

	result, err := suite.Auth.CheckUserRole(context.Background(), userID, roles)

	assert.NoError(t, err)
	assert.False(t, result)

	suite.MockUserStorage.AssertCalled(t, "GetUserRoleByID", mock.Anything, userID)
}

func TestAuth_CheckUserRole_UserDoesNotExist(t *testing.T) {
	suite := Setup(t)

	userID := int64(1)
	roles := []userv1.Role{userv1.Role_ADMIN}

	suite.MockUserStorage.On("GetUserRoleByID", mock.Anything, userID).Return("", storage.ErrUserNotExists)

	_, err := suite.Auth.CheckUserRole(context.Background(), userID, roles)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotExist)

	suite.MockUserStorage.AssertCalled(t, "GetUserRoleByID", mock.Anything, userID)
}

func TestAuth_CheckUserRole_ErrorGettingRole(t *testing.T) {
	suite := Setup(t)

	userID := int64(1)
	roles := []userv1.Role{userv1.Role_ADMIN}

	suite.MockUserStorage.On("GetUserRoleByID", mock.Anything, userID).Return("", errors.New("database error"))

	_, err := suite.Auth.CheckUserRole(context.Background(), userID, roles)

	assert.Error(t, err)

	suite.MockUserStorage.AssertCalled(t, "GetUserRoleByID", mock.Anything, userID)
}

func TestAuth_Register(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	password := gofakeit.Password(true, true, true, false, false, 12)

	dto := &dtos.UserRegisterDTO{Email: gofakeit.Email(), Password: password}
	user := dto.ToDomain()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user.PasswordHash = passwordHash

	suite.MockUserStorage.On("SaveUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	suite.MockTokenStorage.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, err := suite.Auth.Register(context.Background(), dto)

	// Assert that no error was returned
	assert.NoError(t, err)
	assert.Equal(t, user.ID, result)

	// Assert that the SaveUser function was called with the correct parameters
	suite.MockUserStorage.AssertCalled(t, "SaveUser", mock.Anything, mock.AnythingOfType("*domain.User"))
	suite.MockTokenStorage.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	suite.MockAmqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuth_Register_ErrorUserExists(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	password := gofakeit.Password(true, true, true, false, false, 12)

	dto := &dtos.UserRegisterDTO{Email: gofakeit.Email(), Password: password}
	user := dto.ToDomain()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user.PasswordHash = passwordHash

	suite.MockUserStorage.On("SaveUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(storage.ErrUserExists)

	_, err = suite.Auth.Register(context.Background(), dto)

	assert.ErrorIs(t, err, ErrUserExists)

	suite.MockUserStorage.AssertCalled(t, "SaveUser", mock.Anything, mock.AnythingOfType("*domain.User"))
}

func TestAuth_Register_ErrorSavingUser(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	password := gofakeit.Password(true, true, true, false, false, 12)

	dto := &dtos.UserRegisterDTO{Email: gofakeit.Email(), Password: password}
	user := dto.ToDomain()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user.PasswordHash = passwordHash

	suite.MockUserStorage.On("SaveUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

	_, err = suite.Auth.Register(context.Background(), dto)

	assert.Error(t, err)

	suite.MockUserStorage.AssertCalled(t, "SaveUser", mock.Anything, mock.AnythingOfType("*domain.User"))
}

func TestAuth_Activate(t *testing.T) {
	suite := Setup(t)

	token := gofakeit.UUID()

	user := &domain.User{ID: 1}

	suite.MockUserStorage.On("ActivateUser", mock.Anything, user.ID).Return(nil)
	suite.MockTokenStorage.On("Get", mock.Anything, token).Return(user.ID, nil)
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.MockTokenStorage.On("Delete", mock.Anything, token).Return(nil)
	suite.MockAmqp.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := suite.Auth.ActivateUser(context.Background(), token)

	// Assert that no error was returned
	assert.NoError(t, err)

	// Assert that the ActivateUser function was called with the correct parameters
	suite.MockUserStorage.AssertCalled(t, "ActivateUser", mock.Anything, user.ID)
	suite.MockTokenStorage.AssertCalled(t, "Get", mock.Anything, token)
	suite.MockTokenStorage.AssertCalled(t, "Delete", mock.Anything, token)
	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.MockAmqp.AssertCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuth_Activate_ErrorActivatingUser(t *testing.T) {
	suite := Setup(t)

	token := gofakeit.UUID()

	user := &domain.User{ID: 1}

	suite.MockUserStorage.On("ActivateUser", mock.Anything, user.ID).Return(errors.New("database error"))
	suite.MockTokenStorage.On("Get", mock.Anything, token).Return(user.ID, nil)

	err := suite.Auth.ActivateUser(context.Background(), token)

	assert.Error(t, err)

	suite.MockUserStorage.AssertCalled(t, "ActivateUser", mock.Anything, user.ID)
	suite.MockTokenStorage.AssertCalled(t, "Get", mock.Anything, token)
}

func TestAuth_RefreshToken(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	user := &domain.User{ID: 1}

	tokenPair, err := jwt.GenerateTokenPair(user.ID, suite.Auth.JwtCfg)
	require.NoError(t, err)

	suite.MockTokenStorage.On("Get", mock.Anything, tokenPair["rt_token"]).Return(user.ID, nil)
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(user, nil)
	suite.MockTokenStorage.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	result, err := suite.Auth.RefreshToken(context.Background(), tokenPair["rt_token"], tokenPair["access_token"])

	// Assert that no error was returned
	require.NoError(t, err)
	assert.Equal(t, user, result.User)

	// Assert that the GetUserByID function was called with the correct parameters
	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.MockTokenStorage.AssertCalled(t, "Get", mock.Anything, tokenPair["rt_token"])
	suite.MockTokenStorage.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuth_RefreshToken_ErrorGettingUser(t *testing.T) {
	suite := Setup(t)

	// Define the expected user
	user := &domain.User{ID: 1}

	tokenPair, err := jwt.GenerateTokenPair(user.ID, suite.Auth.JwtCfg)
	require.NoError(t, err)

	suite.MockTokenStorage.On("Get", mock.Anything, tokenPair["rt_token"]).Return(user.ID, nil)
	suite.MockUserStorage.On("GetUserByID", mock.Anything, user.ID).Return(&domain.User{}, storage.ErrUserNotExists)

	_, err = suite.Auth.RefreshToken(context.Background(), tokenPair["rt_token"], tokenPair["access_token"])

	assert.ErrorIs(t, err, ErrUserNotExist)

	suite.MockUserStorage.AssertCalled(t, "GetUserByID", mock.Anything, user.ID)
	suite.MockTokenStorage.AssertCalled(t, "Get", mock.Anything, tokenPair["rt_token"])
}

func TestAuth_RefreshToken_ErrTokenNotFound(t *testing.T) {
	suite := Setup(t)

	tokenPair, err := jwt.GenerateTokenPair(1, suite.Auth.JwtCfg)
	require.NoError(t, err)

	suite.MockTokenStorage.On("Get", mock.Anything, tokenPair["rt_token"]).Return(int64(0), storage.ErrTokenNotExists)

	_, err = suite.Auth.RefreshToken(context.Background(), tokenPair["rt_token"], tokenPair["access_token"])

	assert.ErrorIs(t, err, ErrRefreshTokenNotExists)

	suite.MockTokenStorage.AssertCalled(t, "Get", mock.Anything, tokenPair["rt_token"])
}

func TestAuth_Logout(t *testing.T) {
	suite := Setup(t)

	token := gofakeit.UUID()

	suite.MockTokenStorage.On("Delete", mock.Anything, token).Return(nil)

	err := suite.Auth.Logout(context.Background(), token)

	// Assert that no error was returned
	assert.NoError(t, err)

	// Assert that the Delete function was called with the correct parameters
	suite.MockTokenStorage.AssertCalled(t, "Delete", mock.Anything, token)
}

func TestAuth_Logout_ErrorDeletingToken(t *testing.T) {
	suite := Setup(t)

	token := gofakeit.UUID()

	suite.MockTokenStorage.On("Delete", mock.Anything, token).Return(errors.New("database error"))

	err := suite.Auth.Logout(context.Background(), token)

	assert.Error(t, err)

	suite.MockTokenStorage.AssertCalled(t, "Delete", mock.Anything, token)
}
