package domain

import (
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain/models"
	"time"
)

type User struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"second_name"`
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
