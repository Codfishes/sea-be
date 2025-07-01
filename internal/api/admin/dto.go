package admin

import (
	"github.com/google/uuid"
	"time"
)

type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AdminLoginResponse struct {
	AccessToken      string        `json:"access_token"`
	TokenType        string        `json:"token_type"`
	ExpiresInMinutes float64       `json:"expires_in_minutes"`
	Admin            AdminResponse `json:"admin"`
}

type AdminResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type DashboardStatsResponse struct {
	TotalSubscriptions     int     `json:"total_subscriptions"`
	ActiveSubscriptions    int     `json:"active_subscriptions"`
	NewSubscriptions       int     `json:"new_subscriptions"`
	MonthlyRevenue         float64 `json:"monthly_revenue"`
	Reactivations          int     `json:"reactivations"`
	SubscriptionGrowth     float64 `json:"subscription_growth_percentage"`
	RevenueGrowth          float64 `json:"revenue_growth_percentage"`
	PendingTestimonials    int     `json:"pending_testimonials"`
	TotalUsers             int     `json:"total_users"`
	CancelledSubscriptions int     `json:"cancelled_subscriptions"`
}

type DateRangeFilter struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required"`
}

type UserListRequest struct {
	Page    int    `query:"page" validate:"omitempty,min=1"`
	Limit   int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Search  string `query:"search" validate:"omitempty,max=100"`
	Status  string `query:"status" validate:"omitempty,oneof=active inactive all"`
	Role    string `query:"role" validate:"omitempty,oneof=user admin all"`
	SortBy  string `query:"sort_by" validate:"omitempty,oneof=name email created_at last_login_at"`
	SortDir string `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           *string    `json:"email"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone"`
	IsVerified      bool       `json:"is_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`
	ProfileImageURL *string    `json:"profile_image_url"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Additional admin info
	SubscriptionCount int     `json:"subscription_count"`
	TotalSpent        float64 `json:"total_spent"`
}

type UserListResponse struct {
	Users []UserResponse  `json:"users"`
	Meta  *PaginationMeta `json:"meta"`
}

type UpdateUserStatusRequest struct {
	IsActive bool   `json:"is_active"`
	Reason   string `json:"reason,omitempty" validate:"omitempty,max=255"`
}

type PaginationMeta struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Subscription management DTOs
type SubscriptionSearchRequest struct {
	Page       int     `query:"page" validate:"omitempty,min=1"`
	Limit      int     `query:"limit" validate:"omitempty,min=1,max=100"`
	Search     string  `query:"search" validate:"omitempty,max=100"`
	Status     string  `query:"status" validate:"omitempty,oneof=active paused cancelled all"`
	MealPlanID string  `query:"meal_plan_id" validate:"omitempty"`
	UserID     string  `query:"user_id" validate:"omitempty"`
	MinPrice   float64 `query:"min_price" validate:"omitempty,min=0"`
	MaxPrice   float64 `query:"max_price" validate:"omitempty,min=0"`
	DateFrom   string  `query:"date_from" validate:"omitempty"`
	DateTo     string  `query:"date_to" validate:"omitempty"`
	SortBy     string  `query:"sort_by" validate:"omitempty,oneof=created_at updated_at total_price"`
	SortDir    string  `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type SubscriptionSearchResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	UserName     string     `json:"user_name"`
	UserEmail    *string    `json:"user_email"`
	UserPhone    string     `json:"user_phone"`
	MealPlanID   string     `json:"meal_plan_id"`
	MealPlanName string     `json:"meal_plan_name"`
	MealTypes    []string   `json:"meal_types"`
	DeliveryDays []string   `json:"delivery_days"`
	TotalPrice   float64    `json:"total_price"`
	Status       string     `json:"status"`
	PauseStart   *time.Time `json:"pause_start_date,omitempty"`
	PauseEnd     *time.Time `json:"pause_end_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SubscriptionSearchListResponse struct {
	Subscriptions []SubscriptionSearchResponse `json:"subscriptions"`
	Meta          *PaginationMeta              `json:"meta"`
}

type ForceCancelSubscriptionRequest struct {
	Reason        string   `json:"reason" validate:"required,min=5,max=500"`
	RefundAmount  *float64 `json:"refund_amount,omitempty" validate:"omitempty,min=0"`
	NotifyUser    bool     `json:"notify_user"`
	AdminComments string   `json:"admin_comments,omitempty" validate:"omitempty,max=1000"`
}

type ForceCancelSubscriptionResponse struct {
	SubscriptionID string    `json:"subscription_id"`
	CancelledAt    time.Time `json:"cancelled_at"`
	Reason         string    `json:"reason"`
	RefundAmount   *float64  `json:"refund_amount,omitempty"`
	RefundID       *string   `json:"refund_id,omitempty"`
	Message        string    `json:"message"`
}
