package dtos

import (
	userv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/user"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
)

type UserRegisterDTO struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Barcode   string
	Major     string
	GroupName string
	Year      int32
}

func (u *UserRegisterDTO) ToDomain() *domain.User {
	return &domain.User{
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		PasswordHash: nil,
		Barcode:      u.Barcode,
		Major:        u.Major,
		GroupName:    u.GroupName,
		Year:         u.Year,
	}
}

func RegisterRequestToDTO(req *userv1.RegisterRequest) *UserRegisterDTO {
	return &UserRegisterDTO{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Barcode:   req.Barcode,
		Major:     req.Major,
		GroupName: req.GroupName,
		Year:      req.Year,
	}
}
