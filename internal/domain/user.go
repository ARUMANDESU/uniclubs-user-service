package domain

import (
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AvatarURL    string    `json:"avatar_url"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	Activated    bool      `json:"activated"`
	CreatedAt    time.Time `json:"created_at"`
	Role         string    `json:"role"`
	Barcode      string    `json:"barcode"`
	PhoneNumber  string    `json:"phone_number"`
	Major        string    `json:"major"`
	GroupName    string    `json:"group_name"`
	Year         int32     `json:"year"`
}

func (u *User) MapRoleStringToEnum() userv1.Role {
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

func (u *User) ToUserObject() *userv1.UserObject {
	return &userv1.UserObject{
		UserId:    u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarUrl: u.AvatarURL,
		Barcode:   u.Barcode,
		Major:     u.Major,
		GroupName: u.GroupName,
		Year:      u.Year,
		CreatedAt: timestamppb.New(u.CreatedAt),
		Role:      u.MapRoleStringToEnum(),
	}
}

func MapUserArrToUserObjectArr(users []*User) []*userv1.UserObject {
	userObjects := make([]*userv1.UserObject, len(users))
	for i, user := range users {
		userObjects[i] = user.ToUserObject()
	}
	return userObjects
}
