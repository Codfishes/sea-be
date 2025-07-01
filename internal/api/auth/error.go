package auth

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrPhoneAlreadyExists  = errors.New("phone number already exists")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrUserNotVerified     = errors.New("user not verified")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidOTP          = errors.New("invalid OTP")
	ErrOTPExpired          = errors.New("OTP expired")
	ErrOTPNotFound         = errors.New("OTP not found")
	ErrSamePassword        = errors.New("new password cannot be the same as current password")
	ErrWeakPassword        = errors.New("password does not meet strength requirements")
	ErrInvalidPhone        = errors.New("invalid phone number format")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("forbidden access")
	ErrInvalidImageFormat  = errors.New("invalid image format")
	ErrImageTooLarge       = errors.New("image file too large")
)
