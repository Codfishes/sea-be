package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	Name            string     `json:"name" db:"name"`
	Phone           *string    `json:"phone" db:"phone"`
	Password        string     `json:"-" db:"password"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" db:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at" db:"phone_verified_at"`
	ProfileImageURL *string    `json:"profile_image_url" db:"profile_image_url"`
	Role            string     `json:"role" db:"role"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Phone           *string    `json:"phone"`
	IsVerified      bool       `json:"is_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`
	ProfileImageURL *string    `json:"profile_image_url"`
	Role            string     `json:"role"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		Phone:           u.Phone,
		IsVerified:      u.IsVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		PhoneVerifiedAt: u.PhoneVerifiedAt,
		ProfileImageURL: u.ProfileImageURL,
		Role:            u.Role,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil
}

const (
	RoleUser = "user"
)
