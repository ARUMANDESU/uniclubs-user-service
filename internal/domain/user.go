package domain

import (
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type User struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	Role        string    `json:"role"`
	Barcode     string    `json:"barcode"`
	PhoneNumber string    `json:"phone_number"`
	Major       string    `json:"major"`
	GroupName   string    `json:"group_name"`
	Year        int       `json:"year"`
}

func (u User) MapRoleStringToEnum() userv1.Role {
	switch u.Role {
	case "GUEST":
		return userv1.Role_GUEST
	case "USER":
		return userv1.Role_USER
	case "MODER":
		return userv1.Role_MODER
	case "ADMIN":
		return userv1.Role_ADMIN
	case "DSVR":
		return userv1.Role_DSVR
	default:
		return userv1.Role_GUEST // or any default value
	}
}

func UserToModelUser(dUser User) *models.User {
	return &models.User{
		ID:           dUser.ID,
		FirstName:    dUser.FirstName,
		LastName:     dUser.LastName,
		Email:        dUser.Email,
		PasswordHash: nil,   // Assume that we don't have the password hash in domain.User
		Activated:    false, // or determine activation status if possible
		CreatedAt:    dUser.CreatedAt,
		Role:         dUser.Role,
		Barcode:      dUser.Barcode,
		PhoneNumber:  dUser.PhoneNumber,
		Major:        dUser.Major,
		GroupName:    dUser.GroupName,
		Year:         dUser.Year,
	}
}

func ModelUserToDomainUser(mUser models.User) *User {
	return &User{
		ID:          mUser.ID,
		FirstName:   mUser.FirstName,
		LastName:    mUser.LastName,
		Email:       mUser.Email,
		Password:    "", // Assume that we don't convert the password hash back to a password
		CreatedAt:   mUser.CreatedAt,
		Role:        mUser.Role,
		Barcode:     mUser.Barcode,
		PhoneNumber: mUser.PhoneNumber,
		Major:       mUser.Major,
		GroupName:   mUser.GroupName,
		Year:        mUser.Year,
	}
}

func mapUserToUserObject(user *User) *userv1.UserObject {
	return &userv1.UserObject{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Barcode:   user.Barcode,
		Major:     user.Major,
		GroupName: user.GroupName,
		Year:      int32(user.Year),
		CreatedAt: timestamppb.New(user.CreatedAt),
		Role:      user.MapRoleStringToEnum(),
	}
}

func MapUserArrToUserObjectArr(users []*User) []*userv1.UserObject {
	userObjects := make([]*userv1.UserObject, len(users))
	for i, user := range users {
		userObjects[i] = mapUserToUserObject(user)
	}
	return userObjects
}
