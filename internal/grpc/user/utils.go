package user

import (
	"errors"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-user-service/pkg/tokens/jwt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
)

func handleError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrActivationTokenNotFound),
		errors.Is(err, domain.ErrRefreshTokenNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, jwt.ErrUserIDMismatch),
		errors.Is(err, jwt.ErrTokenIsNotValid),
		errors.Is(err, jwt.ErrInvalidTokenClaims),
		errors.Is(err, jwt.ErrUserIDClaimNotFound),
		errors.Is(err, jwt.ErrTokenSignatureIsInvalid),
		errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrUserNonAuthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrRateLimitExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func validateMajor(value any) error {
	major, err := value.(string)
	if !err {
		return errors.New("major must be a string")
	}

	validMajors := []any{"STAFF", "SE", "MT", "IT", "CS", "BDA", "BDH", "ITM", "ITE", "EE", "IoT", "ST", "DJ", "MCs"}

	return validation.Validate(major, validation.In(validMajors...).
		Error("major must be one of the following: STAFF, SE, MT, IT, CS, BDA, BDH, ITM, ITE, EE, IoT, ST, DJ, MCs"))
}

func validateName(value any) error {
	name, err := value.(string)
	if !err {
		return errors.New("name must be a string")
	}

	return validation.Validate(name,
		validation.Match(regexp.MustCompile("^[a-zA-Z]+(([',. -][a-zA-Z ])?[a-zA-Z]*)*$")).Error("name must contain only letters"),
		validation.Length(2, 75).Error("name must be between 2 and 75 characters"),
	)
}

func validateGroupName(value any) error {
	groupName, err := value.(string)
	if !err {
		return errors.New("group name must be a string")
	}

	return validation.Validate(groupName,
		validation.Match(regexp.MustCompile("^[0-9]{4}$")).Error("group name must be a 4 digit number"),
	)
}

func validateYear(value any) error {
	year, err := value.(int32)
	if !err {
		return errors.New("year must be an integer")
	}

	return validation.Validate(year, validation.Min(1).Error("year must be greater than 0"), validation.Max(8).Error("year must be less than 8"))
}

func validatePassword(value any) error {
	password, err := value.(string)
	if !err {
		return errors.New("password must be a string")
	}

	var (
		hasMinLen    = validation.Length(6, 64).Error("password must be between 6 and 64 characters")
		hasUppercase = validation.Match(regexp.MustCompile(`[A-Z]`)).Error("password must contain at least one uppercase letter")
		hasLowercase = validation.Match(regexp.MustCompile(`[a-z]`)).Error("password must contain at least one lowercase letter")
		hasNumber    = validation.Match(regexp.MustCompile(`[0-9]`)).Error("password must contain at least one number")
	)

	return validation.Validate(password, hasMinLen, hasUppercase, hasLowercase, hasNumber)
}
