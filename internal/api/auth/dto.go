package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strong_password"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,phone_id"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        UserInfo  `json:"user"`
}

type UserInfo struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           *string   `json:"phone"`
	IsVerified      bool      `json:"is_verified"`
	ProfileImageURL *string   `json:"profile_image_url"`
	Role            string    `json:"role"`
}

type UpdateProfileRequest struct {
	Name  string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone *string `json:"phone,omitempty" validate:"omitempty,phone_id"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,strong_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTP         string `json:"otp" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,strong_password"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
	Type  string `json:"type" validate:"required,oneof=email_verification password_reset"`
}

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=email_verification password_reset"`
}

type UploadProfileImageResponse struct {
	ImageURL string `json:"image_url"`
	Message  string `json:"message"`
}

type LogoutRequest struct {
	Token string `json:"token" validate:"required"`
}
